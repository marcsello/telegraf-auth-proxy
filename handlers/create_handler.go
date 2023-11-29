package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/influxdata/telegraf"
	"github.com/marcsello/telegraf-auth-proxy/middleware"
	"github.com/marcsello/telegraf-auth-proxy/proxy"
	"go.uber.org/zap"
	"net/http"
)

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
			middleware.GetLoggerFromCtx(ctx).Error("Failed to read request body", zap.String("format", "json"), zap.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// handle encoding
		// we have got to keep the original body intact, since we are going to pass it to the upstream
		var inflatedBodyBytes []byte
		encoding := ctx.GetHeader("Content-Encoding")
		inflatedBodyBytes, err = inflateBodyBytes(bodyBytes, encoding)
		if err != nil {
			middleware.GetLoggerFromCtx(ctx).Error("Failed to inflate request body", zap.String("format", "json"), zap.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Parse the inflated data
		var metrics []telegraf.Metric
		metrics, err = parser.Parse(inflatedBodyBytes)
		if err != nil {
			middleware.GetLoggerFromCtx(ctx).Error("Failed to parse request body. Are you using the right format?", zap.String("format", "json"), zap.Error(err))
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}
		if len(metrics) == 0 {
			// nothing to do here
			middleware.GetLoggerFromCtx(ctx).Debug("Got empty metrics. Ignoring...")
			return
		}

		// validate tag(s)...
		for _, metric := range metrics {
			fieldValue, ok := metric.Tags()[bindFieldName]
			// TODO
		}

		// If all goes well, proxy the intact data
		proxy.ProxyRequest(bodyBytes, ctx)
	}

}
