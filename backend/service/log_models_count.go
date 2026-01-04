package service

import (
	"encoding/json"
	"time"

	"github.com/stardustagi/TopModelsLogs/constants"
	"github.com/stardustagi/TopModelsLogs/models"
	"go.uber.org/zap"
)

// ModelReportStats 模型报表统计结构
type ModelReportStats struct {
	ReportType        string  `json:"report_type"`
	ReportDate        string  `json:"report_date"` // 统计日期，格式：2006-01-02
	TokensPerSecCount float64 `json:"tokens_per_sec_count"`
	LatencyCount      float64 `json:"latency_count"`
	FailureCount      int64   `json:"failure_count"`
	SuccessCount      int64   `json:"success_count"`
	TotalCount        int64   `json:"total_count"` // 用于计算平均值
}

// CountModelsLogs 统计模型调用日志
// 在redis中生成统计报表数据
func (s *LogService) CountModelsLogs(data models.StatusReport) (int64, error) {
	s.logger.Info("统计模型调用日志",
		zap.String("traceId", data.TraceId),
		zap.Int("modelId", data.ModelId),
		zap.String("actualProviderId", data.ActualProviderId))

	// 从 CallerKey 解析 userId（假设 CallerKey 格式包含 userId）
	// 如果没有 userId，使用默认值 0
	userId := int64(0)
	// 这里可以根据实际情况解析 CallerKey 获取 userId

	redisKey := constants.GetReportKey(userId)
	field := constants.GetReportField(data.ModelId, data.ActualProviderId)

	// 获取现有统计数据
	existingData, err := s.rds.HGet(s.ctx, redisKey, field)

	var stats ModelReportStats
	if err == nil && len(existingData) > 0 {
		// 解析现有数据
		if err := json.Unmarshal(existingData, &stats); err != nil {
			s.logger.Warn("解析现有统计数据失败，重新初始化", zap.Error(err))
			stats = ModelReportStats{}
		}
	} else {
		// 初始化新的统计数据
		stats = ModelReportStats{
			ReportType: data.ReportType,
			ReportDate: time.Now().Format("2006-01-02"),
		}
	}

	// 更新统计数据
	stats.TotalCount++
	stats.ReportType = data.ReportType
	stats.ReportDate = time.Now().Format("2006-01-02") // 更新统计日期

	// 计算 tokens_per_sec_count 的累加平均值
	// 新平均值 = (旧平均值 * (总次数-1) + 新值) / 总次数
	if stats.TotalCount == 1 {
		stats.TokensPerSecCount = float64(data.TokensPerSec)
	} else {
		stats.TokensPerSecCount = (stats.TokensPerSecCount*float64(stats.TotalCount-1) + float64(data.TokensPerSec)) / float64(stats.TotalCount)
	}

	// 计算 latency_count 的累加平均值
	latency := s.parseLatency(data.Latency)
	if stats.TotalCount == 1 {
		stats.LatencyCount = latency
	} else {
		stats.LatencyCount = (stats.LatencyCount*float64(stats.TotalCount-1) + latency) / float64(stats.TotalCount)
	}

	// 统计成功/失败次数
	// status_code 为空表示成功，非空表示失败
	if data.StatusCode == "" {
		stats.SuccessCount++
	} else {
		stats.FailureCount++
	}

	// 序列化统计数据
	statsJson, err := json.Marshal(stats)
	if err != nil {
		s.logger.Error("序列化统计数据失败", zap.Error(err))
		return 0, err
	}

	// 写入Redis
	err = s.rds.HSet(s.ctx, redisKey, field, statsJson)
	if err != nil {
		s.logger.Error("写入Redis统计数据失败", zap.Error(err), zap.String("key", redisKey))
		return 0, err
	}
	s.userAlarm(userId, stats)
	s.systemAlarm(stats)

	// 重置过期时间为3小时
	err = s.rds.Expire(s.ctx, redisKey, constants.ReportKeyExpire)
	if err != nil {
		s.logger.Warn("设置过期时间失败", zap.Error(err), zap.String("key", redisKey))
	}

	s.logger.Info("统计模型调用日志成功",
		zap.String("key", redisKey),
		zap.String("field", field),
		zap.Int64("totalCount", stats.TotalCount))

	return stats.TotalCount, nil
}
