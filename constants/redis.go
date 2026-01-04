package constants

import (
	"fmt"
	"time"
)

const (
	ReportKeyExpire  = "3h" // 报表key过期时间3小时
	AlarmKey         = "config:alarm"
	RenewAlarmConfig = 3 * time.Minute
)

// GetReportKey 生成报表Redis key
func GetReportKey(userId int64) string {
	return fmt.Sprintf("cache:report:%d", userId)
}

// GetReportField 生成报表hash field
func GetReportField(modelId int, actualProviderId string) string {
	return fmt.Sprintf("%d_%s", modelId, actualProviderId)
}
