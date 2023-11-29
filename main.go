package main

import (
	"github.com/gin-gonic/gin"
	"github.com/marcsello/telegraf-auth-proxy/handlers"
	"gitlab.com/MikeTTh/env"
	"go.uber.org/zap"
)

var DEBUG = env.Bool("DEBUG", false)

func main() {
	// setup stuff
	var err error
	var logger *zap.Logger
	if DEBUG {
		gin.SetMode(gin.DebugMode)
		logger, err = zap.NewDevelopment()
		logger.Warn("RUNNING IN DEBUG MODE!")
	} else {
		gin.SetMode(gin.ReleaseMode)
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}

	r := gin.New()
	handlers.RegisterHandlers(r, logger)

	// start server
	err = r.Run(env.String("BIND_ADDR", ":8000"))
	if err != nil {
		panic(err)
	}
}
