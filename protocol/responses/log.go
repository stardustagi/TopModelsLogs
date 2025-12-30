package responses

import "github.com/stardustagi/TopModelsLogs/models"

// DefaultResponse 默认响应
type DefaultResponse struct {
	Message string `json:"message"`
}

// GetApiLogListResp 获取API日志列表响应
type GetApiLogListResp struct {
	Logs  []models.ApiLog `json:"logs"`
	Total int             `json:"total"`
}

// GetModelTrainingLogListResp 获取模型训练日志列表响应
type GetModelTrainingLogListResp struct {
	Logs  []models.ModelTrainingLog `json:"logs"`
	Total int                       `json:"total"`
}

// GetModelsCallLogListResp 获取模型调用日志列表响应
type GetModelsCallLogListResp struct {
	Logs  []models.StatusReport `json:"logs"`
	Total int                   `json:"total"`
}
