package service

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/protocol"
	"github.com/stardustagi/TopModelsLogs/constants"
	"github.com/stardustagi/TopModelsLogs/models"
	"github.com/stardustagi/TopModelsLogs/protocol/requests"
	"github.com/stardustagi/TopModelsLogs/protocol/responses"
	"go.uber.org/zap"
)

// CreateModelsCallLog 创建模型调用日志
// @Summary 创建模型调用日志
// @Description 记录模型调用状态日志
// @Tags Log
// @Accept json
// @Produce json
// @Param request body requests.CreateModelsCallLogReq true "创建模型调用日志请求"
// @Success 200 {object} responses.DefaultResponse
// @Router /log/createModelsCallLog [post]
func (s *LogService) CreateModelsCallLog(ctx echo.Context,
	req requests.CreateModelsCallLogReq, resp responses.DefaultResponse) error {
	s.logger.Info("创建模型调用日志",
		zap.String("traceId", req.TraceId),
		zap.String("model", req.Model),
		zap.String("step", req.Step))

	session := s.dao.NewSession()
	defer session.Close()

	// 处理 Stream 字段转换
	stream := 0
	if req.Stream {
		stream = 1
	}

	// 处理 CreatedAt
	var createdAt time.Time
	if req.CreatedAt > 0 {
		createdAt = time.Unix(req.CreatedAt, 0)
	} else {
		createdAt = time.Now()
	}

	// 处理 Latency 转换为字符串
	latency := fmt.Sprintf("%.4f", req.Latency)

	statusReport := &models.StatusReport{
		TraceId:          req.TraceId,
		NodeAddr:         req.NodeAddr,
		Model:            req.Model,
		ModelId:          req.ModelId,
		ActualModel:      req.ActualModel,
		Provider:         req.Provider,
		ActualProvider:   req.ActualProvider,
		ActualProviderId: req.ActualProviderId,
		CallerKey:        req.CallerKey,
		Stream:           stream,
		ReportType:       req.ReportType,
		TokensPerSec:     req.TokensPerSec,
		Latency:          latency,
		Step:             req.Step,
		StatusCode:       req.StatusCode,
		StatusMessage:    req.StatusMessage,
		CreatedAt:        createdAt,
	}

	// 获取日分表表名
	tbName := statusReport.GetSliceDateDayTable()

	// 检查表是否存在，不存在则创建
	exist, err := session.Native().IsTableExist(tbName)
	if err != nil {
		s.logger.Error("检查日志表是否存在失败", zap.Error(err), zap.String("table", tbName))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}
	if !exist {
		// 创建日分表
		err = session.Native().Table(tbName).Sync2(new(models.StatusReport))
		if err != nil {
			s.logger.Error("创建日志表失败", zap.Error(err), zap.String("table", tbName))
			return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
		}
		s.logger.Info("创建日志表成功", zap.String("table", tbName))
	}

	// 插入数据到日分表
	_, err = session.Native().Table(tbName).InsertOne(statusReport)
	if err != nil {
		s.logger.Error("创建模型调用日志失败", zap.Error(err), zap.String("table", tbName))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}

	return protocol.Response(ctx, nil, map[string]interface{}{
		"id":      statusReport.Id,
		"message": "创建模型调用日志成功",
	})
}

// GetModelsCallLogList 获取模型调用日志列表
// @Summary 获取模型调用日志列表
// @Description 分页查询模型调用日志列表
// @Tags Log
// @Accept json
// @Produce json
// @Param request body requests.GetModelsCallLogListReq true "获取模型调用日志列表请求"
// @Success 200 {object} responses.GetModelsCallLogListResp
// @Router /log/getModelsCallLogList [post]
func (s *LogService) GetModelsCallLogList(ctx echo.Context,
	req requests.GetModelsCallLogListReq, resp responses.GetModelsCallLogListResp) error {
	s.logger.Info("获取模型调用日志列表",
		zap.String("traceId", req.TraceId),
		zap.String("model", req.Model))

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
	if req.TraceId != "" {
		query = query.Where("trace_id = ?", req.TraceId)
	}
	if req.Model != "" {
		query = query.And("model = ?", req.Model)
	}
	if req.CallerKey != "" {
		query = query.And("caller_key = ?", req.CallerKey)
	}
	if req.Step != "" {
		query = query.And("step = ?", req.Step)
	}
	if req.ActualProviderId != "" {
		query = query.And("actual_provider_id = ?", req.ActualProviderId)
	}
	if req.StartTime > 0 {
		query = query.And("created_at >= ?", time.Unix(req.StartTime, 0))
	}
	if req.EndTime > 0 {
		query = query.And("created_at <= ?", time.Unix(req.EndTime, 0))
	}

	var logs []models.StatusReport
	total, err := query.
		OrderBy(req.PageInfo.Sort).
		Limit(req.PageInfo.Limit, req.PageInfo.Skip).
		FindAndCount(&logs)
	if err != nil {
		s.logger.Error("查询模型调用日志列表失败", zap.Error(err))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}

	resp.Logs = logs
	resp.Total = int(total)

	return protocol.Response(ctx, nil, resp)
}

// GetModelsCallLogDetail 获取模型调用日志详情
// @Summary 获取模型调用日志详情
// @Description 根据ID获取模型调用日志详情
// @Tags Log
// @Accept json
// @Produce json
// @Param request body requests.GetModelsCallLogDetailReq true "获取模型调用日志详情请求"
// @Success 200 {object} responses.DefaultResponse
// @Router /log/getModelsCallLogDetail [post]
func (s *LogService) GetModelsCallLogDetail(ctx echo.Context,
	req requests.GetModelsCallLogDetailReq, resp responses.DefaultResponse) error {
	s.logger.Info("获取模型调用日志详情", zap.Uint64("id", req.Id))

	session := s.dao.NewSession()
	defer session.Close()

	statusReport := &models.StatusReport{Id: req.Id}
	ok, err := session.FindOne(statusReport)
	if err != nil {
		s.logger.Error("查询模型调用日志失败", zap.Error(err))
		return protocol.Response(ctx, constants.ErrInternalServer.AppendErrors(err), nil)
	}
	if !ok {
		return protocol.Response(ctx, constants.ErrNotDataSet, nil)
	}

	return protocol.Response(ctx, nil, statusReport)
}
