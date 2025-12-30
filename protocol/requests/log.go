package requests

// PageReq 分页请求
type PageReq struct {
	Skip  int    `json:"skip"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort"`
}

// CreateApiLogReq 创建API日志请求
type CreateApiLogReq struct {
	UserId       int64  `json:"user_id"`
	ApiPath      string `json:"api_path" validate:"required"`
	Method       string `json:"method" validate:"required"`
	RequestBody  string `json:"request_body"`
	ResponseBody string `json:"response_body"`
	StatusCode   int    `json:"status_code"`
	Duration     int64  `json:"duration"`
	ClientIP     string `json:"client_ip"`
	UserAgent    string `json:"user_agent"`
	CreatedAt    int64  `json:"created_at"`
}

// GetApiLogListReq 获取API日志列表请求
type GetApiLogListReq struct {
	PageInfo  PageReq `json:"page_info"`
	UserId    int64   `json:"user_id"`
	ApiPath   string  `json:"api_path"`
	StartTime int64   `json:"start_time"`
	EndTime   int64   `json:"end_time"`
}

// GetApiLogDetailReq 获取API日志详情请求
type GetApiLogDetailReq struct {
	Id int64 `json:"id" validate:"required"`
}

// CreateModelTrainingLogReq 创建模型训练日志请求
type CreateModelTrainingLogReq struct {
	UserId       int64   `json:"user_id"`
	ModelId      int64   `json:"model_id"`
	ModelName    string  `json:"model_name"`
	Status       string  `json:"status"`
	LogLevel     string  `json:"log_level"`
	LogMessage   string  `json:"log_message"`
	Epoch        int     `json:"epoch"`
	Loss         float64 `json:"loss"`
	Accuracy     float64 `json:"accuracy"`
	TrainingTime int64   `json:"training_time"`
	CreatedAt    int64   `json:"created_at"`
}

// GetModelTrainingLogListReq 获取模型训练日志列表请求
type GetModelTrainingLogListReq struct {
	PageInfo  PageReq `json:"page_info"`
	UserId    int64   `json:"user_id"`
	ModelId   int64   `json:"model_id"`
	Status    string  `json:"status"`
	LogLevel  string  `json:"log_level"`
	StartTime int64   `json:"start_time"`
	EndTime   int64   `json:"end_time"`
}

// GetModelTrainingLogDetailReq 获取模型训练日志详情请求
type GetModelTrainingLogDetailReq struct {
	Id int64 `json:"id" validate:"required"`
}

// CreateModelsCallLogReq 创建模型调用日志请求
type CreateModelsCallLogReq struct {
	TraceId          string  `json:"trace_id"`
	NodeAddr         string  `json:"node_addr"`
	Model            string  `json:"model"`
	ModelId          int     `json:"model_id"`
	ActualModel      string  `json:"actual_model"`
	Provider         string  `json:"provider"`
	ActualProvider   string  `json:"actual_provider"`
	ActualProviderId string  `json:"actual_provider_id"`
	CallerKey        string  `json:"caller_key"`
	Stream           bool    `json:"stream"`
	ReportType       string  `json:"report_type"`
	TokensPerSec     int     `json:"tokens_per_sec"`
	Latency          float64 `json:"latency"`
	Step             string  `json:"step"`
	StatusCode       string  `json:"status_code"`
	StatusMessage    string  `json:"status_message"`
	CreatedAt        int64   `json:"created_at"`
}

// GetModelsCallLogListReq 获取模型调用日志列表请求
type GetModelsCallLogListReq struct {
	PageInfo         PageReq `json:"page_info"`
	TraceId          string  `json:"trace_id"`
	Model            string  `json:"model"`
	CallerKey        string  `json:"caller_key"`
	Step             string  `json:"step"`
	ActualProviderId string  `json:"actual_provider_id"`
	StartTime        int64   `json:"start_time"`
	EndTime          int64   `json:"end_time"`
}

// GetModelsCallLogDetailReq 获取模型调用日志详情请求
type GetModelsCallLogDetailReq struct {
	Id uint64 `json:"id" validate:"required"`
}
