package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/stardustagi/TopModelsLogs/constants"
	"github.com/stardustagi/TopModelsLogs/models"
	"go.uber.org/zap"
)

// AlarmConfigValue Redis存储的告警配置值结构
type AlarmConfigValue struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// getAlarmConfig 从数据库获取告警配置并同步到Redis
func (s *LogService) getAlarmConfig() error {
	s.logger.Info("开始同步告警配置到Redis")

	session := s.dao.NewSession()
	defer session.Close()

	// 查询 status=1 的告警配置
	var alarmConfigs []models.AlarmConfig
	err := session.Native().
		Where("status = ?", 1).
		Find(&alarmConfigs)
	if err != nil {
		s.logger.Error("查询告警配置失败", zap.Error(err))
		return err
	}

	s.logger.Info("查询到告警配置", zap.Int("count", len(alarmConfigs)))

	// 遍历配置，写入Redis
	for _, config := range alarmConfigs {
		// 生成 field: user_id + type
		field := fmt.Sprintf("%d_%s", config.UserId, config.Type)

		// 生成 value: {"min":xxx,"max":xxx}
		value := AlarmConfigValue{
			Min: config.Min,
			Max: config.Max,
		}
		valueJson, err := json.Marshal(value)
		if err != nil {
			s.logger.Error("序列化告警配置失败",
				zap.Error(err),
				zap.Int64("userId", config.UserId),
				zap.String("type", config.Type))
			continue
		}

		// 写入Redis HSet
		err = s.rds.HSet(s.ctx, constants.AlarmKey, field, valueJson)
		if err != nil {
			s.logger.Error("写入Redis告警配置失败",
				zap.Error(err),
				zap.String("field", field))
			continue
		}

		s.logger.Debug("同步告警配置成功",
			zap.String("field", field),
			zap.Int64("min", config.Min),
			zap.Int64("max", config.Max))
	}

	s.logger.Info("告警配置同步完成", zap.Int("count", len(alarmConfigs)))
	return nil
}

// StartAlarmConfigSyncTask 启动告警配置定时同步任务
// 每3分钟同步一次数据库告警配置到Redis
func (s *LogService) StartAlarmConfigSyncTask() {
	s.logger.Info("启动告警配置定时同步任务，间隔3分钟")

	// 启动时立即执行一次
	if err := s.getAlarmConfig(); err != nil {
		s.logger.Error("初始同步告警配置失败", zap.Error(err))
	}

	// 定时任务
	ticker := time.NewTicker(constants.RenewAlarmConfig)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := s.getAlarmConfig(); err != nil {
					s.logger.Error("定时同步告警配置失败", zap.Error(err))
				}
			case <-s.ctx.Done():
				ticker.Stop()
				s.logger.Info("告警配置定时同步任务已停止")
				return
			}
		}
	}()
}

func (s *LogService) checkAlarm(userId, value int64, alarmType string) bool {
	// 生成 field: user_id + type
	field := fmt.Sprintf("%d_%s", userId, alarmType)

	// 从Redis读取报警配置
	configData, err := s.rds.HGet(s.ctx, constants.AlarmKey, field)
	if err != nil {
		s.logger.Debug("获取报警配置失败",
			zap.Error(err),
			zap.String("field", field))
		return false
	}

	if len(configData) == 0 {
		s.logger.Debug("报警配置不存在,勿略报警", zap.String("field", field))
		return true
	}

	// 解析配置
	var config AlarmConfigValue
	if err := json.Unmarshal(configData, &config); err != nil {
		s.logger.Error("解析报警配置失败",
			zap.Error(err),
			zap.String("field", field))
		return false
	}

	// 检查 value 是否在 min 和 max 之间
	if value >= int64(config.Min) && value <= int64(config.Max) {
		return true
	}

	return false
}

func (s *LogService) userAlarm(userId int64, stats ModelReportStats) {
	if s.checkAlarm(userId, int64(stats.TokensPerSecCount), "token") {
		s.logger.Error("tokens_per_sec 超出阈值报警", zap.Any("stats", stats))
	}
	if s.checkAlarm(userId, int64(stats.LatencyCount), "latency") {
		s.logger.Error("latency 超出阈值报警", zap.Any("stats", stats))
	}
	if s.checkAlarm(userId, stats.FailureCount, "failure") {
		s.logger.Error("failure_count 超出阈值报警", zap.Any("stats", stats))
	}
	// todo: 确定需求以后发对应的报警消息给用户
}

func (s *LogService) systemAlarm(stats ModelReportStats) {
	if s.checkAlarm(0, int64(stats.TokensPerSecCount), "token") {
		s.logger.Error("系统级 tokens_per_sec 超出阈值报警", zap.Any("stats", stats))
	}
	if s.checkAlarm(0, int64(stats.LatencyCount), "latency") {
		s.logger.Error("系统级 latency 超出阈值报警", zap.Any("stats", stats))
	}
	if s.checkAlarm(0, stats.FailureCount, "failure") {
		s.logger.Error("系统级 failure_count 超出阈值报警", zap.Any("stats", stats))
	}
	// todo: 确定需求以后发对应的报警消息给管理员
}
