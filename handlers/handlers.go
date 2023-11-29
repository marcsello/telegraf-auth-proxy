package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/marcsello/telegraf-auth-proxy/middleware"
	"gitlab.com/MikeTTh/env"
	"go.uber.org/zap"
)

var bindFieldName = env.String("BIND_FIELD", "host")
var maxBodyLen = env.Int("MAX_BODY_LEN", 10737418240) // 10M by default

func RegisterHandlers(r *gin.Engine, logger *zap.Logger) {
	r.Use(middleware.GoodLoggerMiddleware(logger))
	r.Use(middleware.BasicAuthMiddleware(env.String("HTPASSWD_PATH", ".htpasswd")))
	r.POST("/line", handleLine)
	r.POST("/json", handleJSON)
	r.POST("/csv", handleCSV)
	r.PUT("/line", handleLine)
	r.PUT("/json", handleJSON)
	r.PUT("/csv", handleCSV)
}
