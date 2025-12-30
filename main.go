package main

import (
	"github.com/stardustagi/TopLib/libs/conf"
	"github.com/stardustagi/TopLib/libs/databases"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/libs/redis"
	"github.com/stardustagi/TopModelsLogs/backend"
	"github.com/stardustagi/TopModelsLogs/backend/service"
	"github.com/stardustagi/TopModelsLogs/constants"

	_ "github.com/stardustagi/TopModelsLogs/docs"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title TopModelsLogs API
// @version 1.0
// @description LLM 日志服务 API 文档
// @host localhost:8082
// @BasePath /
func main() {
	conf.Init()
	loggerConfig := conf.Get("logger")
	constants.Init()
	logs.Init(loggerConfig)
	logger := logs.GetLogger("main")
	logger.Info("Init logs")
	_, _ = databases.Init(conf.Get("mysql"))
	logger.Info("Init mysql")
	_, _ = redis.Init(conf.Get("redis"))
	logger.Info("Init redis")

	app := backend.NewApplication(conf.Get("websrv"))
	// 添加swagger
	app.AddNativeHandler("GET", "/swagger/*", echoSwagger.WrapHandler)

	// 启动日志服务
	logService := service.GetLogServiceInstance()
	logService.Start(app)
	logger.Info("Log service started")

	app.Start()
	// 停止服务
	app.Stop()
	logService.Stop()
	logger.Info("Services stopped")
}
