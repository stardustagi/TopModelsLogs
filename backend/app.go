package backend

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/libs/databases"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/libs/server"
	"github.com/stardustagi/TopLib/utils"
	"github.com/stardustagi/TopModelsLogs/models"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

type Application struct {
	ctx    context.Context
	b      *server.Backend
	logger *zap.Logger
	config server.HttpServerConfig
}

func NewApplication(configBytes []byte) *Application {
	config, err := utils.Bytes2Struct[server.HttpServerConfig](configBytes)
	if err != nil {
		panic(err)
	}
	b, err := server.NewBackend(configBytes)
	if err != nil {
		panic(err)
	}
	app := &Application{
		ctx:    context.Background(),
		config: config,
		logger: logs.GetLogger("HttpBackend"),
		b:      b,
	}
	// 注册 swagger 路由
	app.b.AddNativeHandler("GET", "/swagger/*", echoSwagger.WrapHandler)
	return app
}

func (h *Application) Start() {
	h.syncDatabaseSchema()
	go func() {
		if err := h.b.Start(); err != nil {
			h.logger.Error("backend.Start error", zap.Error(err))
		}
	}()
}

func (h *Application) Stop() {
	h.logger.Info("Stopping HttpBackend")
	h.b.Stop()
}

func (h *Application) AddGroup(group string, middleware ...echo.MiddlewareFunc) {
	h.b.AddGroup(group, middleware...)
}

func (h *Application) AddPostHandler(group string, handler server.IHandler) {
	h.b.AddPostHandler(group, handler)
}

func (h *Application) AddGetHandler(group string, handler server.IHandler) {
	h.b.AddGetHandler(group, handler)
}

func (h *Application) AddNativeHandler(method, path string, handler echo.HandlerFunc) {
	h.b.AddNativeHandler(method, path, handler)
}

func (h *Application) syncDatabaseSchema() {
	h.logger.Info("Syncing database schema...")
	modelList := []interface{}{
		&models.ApiLog{},
		&models.ModelTrainingLog{},
	}

	dao := databases.GetDao()
	session := dao.NewSession()
	defer session.Close()

	for _, model := range modelList {
		err := session.Native().Sync2(model)
		if err != nil {
			h.logger.Error("Sync database schema failed", zap.Error(err), zap.Any("model", model))
		}
	}
	h.logger.Info("Database schema synced successfully")
}
