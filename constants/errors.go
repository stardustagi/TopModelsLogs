package constants

import (
	topError "github.com/stardustagi/TopLib/libs/errors"
)

var (
	ErrInternalServer = topError.New("Internal server error", 500)
	ErrInvalidParams  = topError.New("无效的请求参数", 501)
	ErrNotDataSet     = topError.New("数据不存在", 1001)
	ErrAuthFailed     = topError.New("认证失败", 1002)
)
