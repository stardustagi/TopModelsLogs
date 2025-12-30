# TopModelsLogs服务

## 用途
 - 记录和存储TopModels平台上模型训练和评估的日志信息。
 - 记录调用TopModels API的日志，便于调试和监控。
 - 提供日志查询和分析功能，帮助用户了解模型训练过程中的问题和性能

## 技术栈
    - 编程语言: Golang
    - 数据存储：Mysql
    - 缓存 :redis
    - 消息队列: nats

## 项目结构
- /backend: 后端服务代码
    - /backend/service 服务逻辑代码
      - /backend/service/log_service.go: 日志服务实现
    - app.go: 后端服务入口
    - app_middleware.go: 中间件配置
- /config: 配置文件
- /constants: 常量定义
- /docs: 项目文档
- /logs: 日志文件
- /models: 数据库模型定义
- /protocol: Http请求协议文件
    - /protocol/requests: 请求结构体定义
    - /protocol/responses: 响应结构体定义
- main.go : 项目入口文件

## 项目参考
 TopModelsPlatform