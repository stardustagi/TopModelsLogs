package models

// ApiLog API调用日志
type ApiLog struct {
	Id           int64  `json:"id" xorm:"'id' pk autoincr BIGINT(20)"`
	UserId       int64  `json:"user_id" xorm:"'user_id' BIGINT(20) index"`
	ApiPath      string `json:"api_path" xorm:"'api_path' VARCHAR(255) index"`
	Method       string `json:"method" xorm:"'method' VARCHAR(10)"`
	RequestBody  string `json:"request_body" xorm:"'request_body' TEXT"`
	ResponseBody string `json:"response_body" xorm:"'response_body' TEXT"`
	StatusCode   int    `json:"status_code" xorm:"'status_code' INT(10)"`
	Duration     int64  `json:"duration" xorm:"'duration' BIGINT(20) comment('请求耗时，毫秒')"`
	ClientIP     string `json:"client_ip" xorm:"'client_ip' VARCHAR(50)"`
	UserAgent    string `json:"user_agent" xorm:"'user_agent' VARCHAR(500)"`
	CreatedAt    int64  `json:"created_at" xorm:"'created_at' BIGINT(20) index"`
}

func (ApiLog) TableName() string {
	return "api_log"
}

// ModelTrainingLog 模型训练日志
type ModelTrainingLog struct {
	Id           int64   `json:"id" xorm:"'id' pk autoincr BIGINT(20)"`
	UserId       int64   `json:"user_id" xorm:"'user_id' BIGINT(20) index"`
	ModelId      int64   `json:"model_id" xorm:"'model_id' BIGINT(20) index"`
	ModelName    string  `json:"model_name" xorm:"'model_name' VARCHAR(128)"`
	Status       string  `json:"status" xorm:"'status' VARCHAR(32) index comment('训练状态: pending, running, completed, failed')"`
	LogLevel     string  `json:"log_level" xorm:"'log_level' VARCHAR(16) comment('日志级别: info, warn, error, debug')"`
	LogMessage   string  `json:"log_message" xorm:"'log_message' TEXT"`
	Epoch        int     `json:"epoch" xorm:"'epoch' INT(10) comment('当前轮次')"`
	Loss         float64 `json:"loss" xorm:"'loss' DOUBLE comment('损失值')"`
	Accuracy     float64 `json:"accuracy" xorm:"'accuracy' DOUBLE comment('准确率')"`
	TrainingTime int64   `json:"training_time" xorm:"'training_time' BIGINT(20) comment('训练时长，秒')"`
	CreatedAt    int64   `json:"created_at" xorm:"'created_at' BIGINT(20) index"`
}

func (ModelTrainingLog) TableName() string {
	return "model_training_log"
}
