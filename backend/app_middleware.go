package backend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/libs/jwt"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/libs/redis"
	"github.com/stardustagi/TopModelsLogs/constants"
	"go.uber.org/zap"
)

// LogUserAccess 日志服务用户访问中间件
func LogUserAccess() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			jwtstr := c.Request().Header.Get("jwt")
			UserId := c.Request().Header.Get("id")
			secret := fmt.Sprintf("%s-%s-%s", constants.AppName, constants.AppVersion, UserId)
			if jwtstr != "" {
				jwtobj, ok := jwt.JWTDecrypt(jwtstr, secret)
				if !ok || jwtobj == nil || jwtobj["token"] == nil || jwtobj["id"] == nil {
					return c.JSON(401, map[string]interface{}{
						"errcode": 2,
						"errmsg":  "jwt解析错误",
					})
				}

				id, ok := jwtobj["id"].(string)
				if !ok {
					return c.JSON(401, map[string]interface{}{
						"errcode": 2,
						"errmsg":  "用户信息获取失败",
					})
				}
				intId, err := strconv.ParseInt(id, 10, 64)
				if err != nil {
					return c.JSON(401, map[string]interface{}{
						"errcode": 2,
						"errmsg":  "用户信息获取失败",
					})
				}
				tokenKey := fmt.Sprintf("%s:%s:user:%s", constants.AppName, constants.AppVersion, constants.LogUserTokenKey(intId))
				logs.Info("redis key is:", zap.String("tokenKey", tokenKey))
				redisCmd := redis.GetRedisDb()
				oldToken, err1 := redisCmd.Get(context.Background(), tokenKey).Result()
				if err1 != nil {
					return c.JSON(401, map[string]interface{}{
						"errcode": 2,
						"errmsg":  "获取token失败",
					})
				}
				if jwtobj["token"] != oldToken {
					return c.JSON(401, map[string]interface{}{
						"errcode": 2,
						"errmsg":  "token不匹配",
					})
				}
				c.Request().Header.Set("id", id)
			}
			return next(c)
		}
	}
}
