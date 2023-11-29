package handlers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/snappy"
	"github.com/influxdata/telegraf"
	"github.com/marcsello/telegraf-auth-proxy/middleware"
	"github.com/marcsello/telegraf-auth-proxy/proxy"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func inflateBodyBytes(originBytes []byte, encoding string) ([]byte, error) {
	// inspired by: https://github.com/influxdata/telegraf/blob/6814d7af8a4134d8e05bee47f597df4e930eba69/plugins/inputs/http_listener_v2/http_listener_v2.go#L252
	switch encoding {
	case "gzip":
		gzipReader, err := gzip.NewReader(bytes.NewReader(originBytes))
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()

		reader := io.LimitReader(gzipReader, int64(maxBodyLen))
		return io.ReadAll(reader)

	case "snappy":
		// snappy block format is only supported by decode/encode not snappy reader/writer
		return snappy.Decode(nil, originBytes)

	default:
		// do nothing
		return originBytes, nil
	}
}

func readBody(ctx *gin.Context) ([]byte, error) {
	r := io.LimitReader(ctx.Request.Body, int64(maxBodyLen))
	return io.ReadAll(r)
}

func validateMetrics(tagToCheck, expectedValue string, metrics []telegraf.Metric) error {
	// validate tag(s)...
	for i, metric := range metrics {
		fieldValue, ok := metric.GetTag(tagToCheck)
		if !ok {
			// tag was not present, consider invalid
			return fmt.Errorf("expected tag %s is not present in metric element %d", tagToCheck, i)
		}
		if fieldValue != expectedValue {
			// tag have unexpected value, consider invalid
			return fmt.Errorf("tag %s have unexpected value %s (expected %s) in metric element %d", tagToCheck, fieldValue, expectedValue, i)
		}
	}

	// all went well
	return nil
}

func createHandler(parser telegraf.Parser) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Parse and verify JSON
		expectedUser, ok := middleware.GetBasicAuthUserFromCtx(ctx)
		if !ok || expectedUser == "" {
			ctx.Status(500) // middleware should have prevented this, so if we get there, something went wrong
			return
		}
		var err error

		// Read body bytes
		var bodyBytes []byte
		bodyBytes, err = readBody(ctx)
		if err != nil {
			middleware.GetLoggerFromCtx(ctx).Error("Failed to read request body", zap.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// handle encoding
		// we have got to keep the original body intact, since we are going to pass it to the upstream
		var inflatedBodyBytes []byte
		encoding := ctx.GetHeader("Content-Encoding")
		inflatedBodyBytes, err = inflateBodyBytes(bodyBytes, encoding)
		if err != nil {
			middleware.GetLoggerFromCtx(ctx).Error("Failed to inflate request body", zap.String("encoding", encoding), zap.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Parse the inflated data
		var metrics []telegraf.Metric
		metrics, err = parser.Parse(inflatedBodyBytes)
		if err != nil {
			middleware.GetLoggerFromCtx(ctx).Error("Failed to parse request body. Are you using the right format?", zap.Error(err))
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}
		if len(metrics) == 0 {
			// nothing to do here
			middleware.GetLoggerFromCtx(ctx).Debug("Got empty metrics. Ignoring...")
			return
		}

		err = validateMetrics(authenticatedTagName, expectedUser, metrics)
		if err != nil {
			middleware.GetLoggerFromCtx(ctx).Warn("Basic auth succeeded, but metric failed authentication! Ignoring...", zap.Error(err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// If all goes well, proxy the intact data
		proxy.ProxyRequest(bodyBytes, ctx)
	}
}
