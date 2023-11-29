package handlers

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/golang/snappy"
	"io"
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
