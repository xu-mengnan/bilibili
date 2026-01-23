package middleware

import (
	"time"

	"bilibili/pkg/utils"
	"github.com/gin-gonic/gin"
)

// Logging 记录HTTP请求日志的中间件
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 记录日志
		fields := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"query":      query,
			"status":     statusCode,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"latency":    latency,
		}

		if statusCode >= 500 {
			utils.LogErrorFields(nil, fields, "HTTP request")
		} else if statusCode >= 400 {
			utils.LogWarnFields(fields, "HTTP request")
		} else {
			utils.LogInfoFields(fields, "HTTP request")
		}
	}
}
