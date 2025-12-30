package service

import (
	"context"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/libs/databases"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/libs/redis"
	"github.com/stardustagi/TopLib/libs/server"
	"github.com/stardustagi/TopLib/protocol"
	"github.com/stardustagi/TopModelsLogs/backend"
	"github.com/stardustagi/TopModelsLogs/constants"
	"github.com/stardustagi/TopModelsLogs/models"
	"github.com/stardustagi/TopModelsLogs/protocol/requests"
	"github.com/stardustagi/TopModelsLogs/protocol/responses"
	"go.uber.org/zap"
)

type LogService struct {
	logger    *zap.Logger
	ctx       context.Context
	cancelCtx context.CancelFunc
	dao       databases.BaseDao
	rds       redis.RedisCli
	app       *backend.Application
}

var (
	logServiceInstance *LogService
	logServiceOnce     sync.Once
)

// GetLogServiceInstance 获取日志服务实例
func GetLogServiceInstance() *LogService {
	logServiceOnce.Do(func() {
		logServiceInstance = NewLogService()
	})
	return logServiceInstance
}

// NewLogService 创建新的日志服务
func NewLogService() *LogService {
	ctx, cancel := context.WithCancel(context.Background())
	return &LogService{
		logger:    logs.GetLogger("LogService"),
		ctx:       ctx,
		cancelCtx: cancel,
		dao:       databases.GetDao(),
		rds: redis.NewRedisView(redis.GetRedisDb(),
			constants.ApplicationPrefix,
			logs.GetLogger("LogRedis")),
	}
}

func (s *LogService) Start(app *backend.Application) {
	if app == nil {
		panic("请设置后端应用")
	}
	s.app = app
	s.initialization()
	s.logger.Info("Starting LogService...")
}

func (s *LogService) Stop() {
	s.logger.Info("Stopping LogService...")
	s.cancelCtx()
	s.logger.Info("LogService stopped.")
}

func (s *LogService) initialization() {
	s.app.AddGroup("log", server.Request())

	s.app.AddPostHandler("log", server.NewHandler(
		"createApiLog",
		[]string{"log", "api"},
		s.CreateApiLog))

	s.app.AddPostHandler("log", server.NewHandler(
		"getApiLogList",
		[]string{"log", "api"},
		s.GetApiLogList))

	s.app.AddPostHandler("log", server.NewHandler(
		"getApiLogDetail",
		[]string{"log", "api"},
		s.GetApiLogDetail))

	s.app.AddPostHandler("log", server.NewHandler(
		"createModelTrainingLog",
		[]string{"log", "training"},
		s.CreateModelTrainingLog))

	s.app.AddPostHandler("log", server.NewHandler(
		"getModelTrainingLogList",
		[]string{"log", "training"},
		s.GetModelTrainingLogList))

	s.app.AddPostHandler("log", server.NewHandler(
		"getModelTrainingLogDetail",
		[]string{"log", "training"},
		s.GetModelTrainingLogDetail))
}

// CreateApiLog 创建API调用日志
// @Summary 创建API调用日志
// @Description 记录API调用日志
// @Tags Log
// @Accept json
// @Produce json
// @Param request body requests.CreateApiLogReq true "创建API日志请求"
// @Success 200 {object} responses.DefaultResponse
// @Router /log/createApiLog [post]
func (s *LogService) CreateApiLog(ctx echo.Context,
	req requests.CreateApiLogReq, resp responses.DefaultResponse) error {
	s.logger.Info("创建API调用日志",
		zap.Int64("userId", req.UserId),
		zap.String("apiPath", req.ApiPath))

	session := s.dao.NewSession()
	defer session.Close()

	apiLog := &models.ApiLog{
		UserId:       req.UserId,
		ApiPath:      req.ApiPath,
		Method:       req.Method,
		RequestBody:  req.RequestBody,
		ResponseBody: req.ResponseBody,
		StatusCode:   req.StatusCode,
		Duration:     req.Duration,
		ClientIP:     req.ClientIP,
		UserAgent:    req.UserAgent,
		CreatedAt:    req.CreatedAt,
	}

	_, err := session.InsertOne(apiLog)
	if err != nil {
		s.logger.Error("创建API日志失败", zap.Error(err))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}

	return protocol.Response(ctx, nil, map[string]interface{}{
		"id":      apiLog.Id,
		"message": "创建API日志成功",
	})
}

// GetApiLogList 获取API日志列表
// @Summary 获取API日志列表
// @Description 分页查询API日志列表
// @Tags Log
// @Accept json
// @Produce json
// @Param request body requests.GetApiLogListReq true "获取API日志列表请求"
// @Success 200 {object} responses.GetApiLogListResp
// @Router /log/getApiLogList [post]
func (s *LogService) GetApiLogList(ctx echo.Context,
	req requests.GetApiLogListReq, resp responses.GetApiLogListResp) error {
	s.logger.Info("获取API日志列表", zap.Int64("userId", req.UserId))

	session := s.dao.NewSession()
	defer session.Close()

	// 默认分页
	if req.PageInfo.Limit <= 0 {
		req.PageInfo.Limit = 20
	}
	if req.PageInfo.Sort == "" {
		req.PageInfo.Sort = "id desc"
	}

	query := session.Native().NewSession()
	defer query.Close()

	// 可选条件
	if req.UserId > 0 {
		query = query.Where("user_id = ?", req.UserId)
	}
	if req.ApiPath != "" {
		query = query.And("api_path LIKE ?", "%"+req.ApiPath+"%")
	}
	if req.StartTime > 0 {
		query = query.And("created_at >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		query = query.And("created_at <= ?", req.EndTime)
	}

	var logs []models.ApiLog
	total, err := query.
		OrderBy(req.PageInfo.Sort).
		Limit(req.PageInfo.Limit, req.PageInfo.Skip).
		FindAndCount(&logs)
	if err != nil {
		s.logger.Error("查询API日志列表失败", zap.Error(err))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}

	resp.Logs = logs
	resp.Total = int(total)

	return protocol.Response(ctx, nil, resp)
}

// GetApiLogDetail 获取API日志详情
// @Summary 获取API日志详情
// @Description 根据ID获取API日志详情
// @Tags Log
// @Accept json
// @Produce json
// @Param request body requests.GetApiLogDetailReq true "获取API日志详情请求"
// @Success 200 {object} responses.DefaultResponse
// @Router /log/getApiLogDetail [post]
func (s *LogService) GetApiLogDetail(ctx echo.Context,
	req requests.GetApiLogDetailReq, resp responses.DefaultResponse) error {
	s.logger.Info("获取API日志详情", zap.Int64("id", req.Id))

	session := s.dao.NewSession()
	defer session.Close()

	apiLog := &models.ApiLog{Id: req.Id}
	ok, err := session.FindOne(apiLog)
	if err != nil {
		s.logger.Error("查询API日志失败", zap.Error(err))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}
	if !ok {
		return protocol.Response(ctx, constants.ErrNotDataSet, nil)
	}

	return protocol.Response(ctx, nil, apiLog)
}

// CreateModelTrainingLog 创建模型训练日志
// @Summary 创建模型训练日志
// @Description 记录模型训练日志
// @Tags Log
// @Accept json
// @Produce json
// @Param request body requests.CreateModelTrainingLogReq true "创建模型训练日志请求"
// @Success 200 {object} responses.DefaultResponse
// @Router /log/createModelTrainingLog [post]
func (s *LogService) CreateModelTrainingLog(ctx echo.Context,
	req requests.CreateModelTrainingLogReq, resp responses.DefaultResponse) error {
	s.logger.Info("创建模型训练日志",
		zap.Int64("userId", req.UserId),
		zap.String("modelName", req.ModelName))

	session := s.dao.NewSession()
	defer session.Close()

	trainingLog := &models.ModelTrainingLog{
		UserId:       req.UserId,
		ModelId:      req.ModelId,
		ModelName:    req.ModelName,
		Status:       req.Status,
		LogLevel:     req.LogLevel,
		LogMessage:   req.LogMessage,
		Epoch:        req.Epoch,
		Loss:         req.Loss,
		Accuracy:     req.Accuracy,
		TrainingTime: req.TrainingTime,
		CreatedAt:    req.CreatedAt,
	}

	_, err := session.InsertOne(trainingLog)
	if err != nil {
		s.logger.Error("创建模型训练日志失败", zap.Error(err))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}

	return protocol.Response(ctx, nil, map[string]interface{}{
		"id":      trainingLog.Id,
		"message": "创建模型训练日志成功",
	})
}

// GetModelTrainingLogList 获取模型训练日志列表
// @Summary 获取模型训练日志列表
// @Description 分页查询模型训练日志列表
// @Tags Log
// @Accept json
// @Produce json
// @Param request body requests.GetModelTrainingLogListReq true "获取模型训练日志列表请求"
// @Success 200 {object} responses.GetModelTrainingLogListResp
// @Router /log/getModelTrainingLogList [post]
func (s *LogService) GetModelTrainingLogList(ctx echo.Context,
	req requests.GetModelTrainingLogListReq, resp responses.GetModelTrainingLogListResp) error {
	s.logger.Info("获取模型训练日志列表", zap.Int64("modelId", req.ModelId))

	session := s.dao.NewSession()
	defer session.Close()

	// 默认分页
	if req.PageInfo.Limit <= 0 {
		req.PageInfo.Limit = 20
	}
	if req.PageInfo.Sort == "" {
		req.PageInfo.Sort = "id desc"
	}

	query := session.Native().NewSession()
	defer query.Close()

	// 可选条件
	if req.UserId > 0 {
		query = query.Where("user_id = ?", req.UserId)
	}
	if req.ModelId > 0 {
		query = query.And("model_id = ?", req.ModelId)
	}
	if req.Status != "" {
		query = query.And("status = ?", req.Status)
	}
	if req.LogLevel != "" {
		query = query.And("log_level = ?", req.LogLevel)
	}
	if req.StartTime > 0 {
		query = query.And("created_at >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		query = query.And("created_at <= ?", req.EndTime)
	}

	var logs []models.ModelTrainingLog
	total, err := query.
		OrderBy(req.PageInfo.Sort).
		Limit(req.PageInfo.Limit, req.PageInfo.Skip).
		FindAndCount(&logs)
	if err != nil {
		s.logger.Error("查询模型训练日志列表失败", zap.Error(err))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}

	resp.Logs = logs
	resp.Total = int(total)

	return protocol.Response(ctx, nil, resp)
}

// GetModelTrainingLogDetail 获取模型训练日志详情
// @Summary 获取模型训练日志详情
// @Description 根据ID获取模型训练日志详情
// @Tags Log
// @Accept json
// @Produce json
// @Param request body requests.GetModelTrainingLogDetailReq true "获取模型训练日志详情请求"
// @Success 200 {object} responses.DefaultResponse
// @Router /log/getModelTrainingLogDetail [post]
func (s *LogService) GetModelTrainingLogDetail(ctx echo.Context,
	req requests.GetModelTrainingLogDetailReq, resp responses.DefaultResponse) error {
	s.logger.Info("获取模型训练日志详情", zap.Int64("id", req.Id))

	session := s.dao.NewSession()
	defer session.Close()

	trainingLog := &models.ModelTrainingLog{Id: req.Id}
	ok, err := session.FindOne(trainingLog)
	if err != nil {
		s.logger.Error("查询模型训练日志失败", zap.Error(err))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}
	if !ok {
		return protocol.Response(ctx, constants.ErrNotDataSet, nil)
	}

	return protocol.Response(ctx, nil, trainingLog)
}
