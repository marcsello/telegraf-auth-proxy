package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/influxdata/telegraf"
	telegrafParsers "github.com/influxdata/telegraf/plugins/parsers"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
	"github.com/marcsello/telegraf-tag-auth-proxy/middleware"
	"gitlab.com/MikeTTh/env"
	"go.uber.org/zap"
	"strings"
)

var authenticatedTagName = env.String("AUTH_TAG", "host")
var maxBodyLen = env.Int("MAX_BODY_LEN", 10737418240)  // 10M by default
var loadParsers = env.String("LOAD_PARSERS", "influx") // load parsers by their names

func RegisterHandlers(r *gin.Engine, logger *zap.Logger) {

	upstreamURL := env.StringOrPanic("PROXY_UPSTREAM_URL")

	r.Use(middleware.GoodLoggerMiddleware(logger))
	r.Use(middleware.BasicAuthMiddleware(
		env.String("HTPASSWD_PATH", ".htpasswd"),
		env.String("BASIC_AUTH_REALM", "restricted-area"),
	))

	pList := strings.Split(loadParsers, ",")
	loadedParserCnt := 0
	for _, name := range pList {
		logger.Debug("Registering parser", zap.String("parser", name))
		creator, ok := telegrafParsers.Parsers[name]
		if !ok {
			logger.Error("Invalid parser. Skipping...", zap.String("parser", name))
			continue
		}

		parser := creator("")

		// https://github.com/influxdata/telegraf/blob/6814d7af8a4134d8e05bee47f597df4e930eba69/config/config_test.go#L781-L783
		initer, ok := parser.(telegraf.Initializer)
		if ok {
			logger.Debug("Parser require initialization...", zap.String("parser", name))
			err := initer.Init()
			if err != nil {
				logger.Error("Failed to initialize parser. Skipping...", zap.String("parser", name), zap.Error(err))
				continue
			}
		}

		handler := createHandler(parser, upstreamURL)
		r.POST("/"+name, handler)
		r.PUT("/"+name, handler)
		loadedParserCnt++
	}

	if loadedParserCnt == 0 {
		panic("no parsers loaded")
	}
	logger.Debug("Loaded parsers", zap.Int("count", loadedParserCnt))
}
