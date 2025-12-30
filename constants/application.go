package constants

import (
	"fmt"
	"os"
)

var (
	ApplicationName   = "TopModelsLogs"
	ApplicationPrefix = "logs"
)

var (
	RedisPrefix   string
	LogsKeyPrefix string
)

var (
	AppName    string
	AppVersion string
)

func Init() {
	AppName = os.Getenv("APP_NAME")
	if AppName == "" {
		AppName = ApplicationName
	}
	AppVersion = os.Getenv("APP_VERSION")
	if AppVersion == "" {
		AppVersion = "v1"
	}
	RedisPrefix = fmt.Sprintf("%s:%s", AppName, AppVersion)
	LogsKeyPrefix = fmt.Sprintf("%s:logs", RedisPrefix)
}

// LogUserTokenKey 用户TokenKey
func LogUserTokenKey(id int64) string {
	return fmt.Sprintf("logUserToken:%d", id)
}
