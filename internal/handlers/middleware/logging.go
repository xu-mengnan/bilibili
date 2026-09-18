package middleware

import (
	"time"

	"bilibili/pkg/utils"
	"github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		queryParamCount := len(c.Request.URL.Query())

		c.Next()

		fields := map[string]interface{}{
			"request_id":  GetRequestID(c),
			"method":      c.Request.Method,
			"path":        path,
			"status":      c.Writer.Status(),
			"ip":          c.ClientIP(),
			"latency_ms":  time.Since(start).Milliseconds(),
			"query_params": queryParamCount,
		}
		if taskID := c.Param("task_id"); taskID != "" {
			fields["task_id"] = taskID
		} else if resourceID := c.Param("id"); resourceID != "" {
			fields["resource_id"] = resourceID
		}

		statusCode := c.Writer.Status()
		if statusCode >= 500 {
			utils.LogErrorFields(nil, fields, "HTTP request")
		} else if statusCode >= 400 {
			utils.LogWarnFields(fields, "HTTP request")
		} else {
			utils.LogInfoFields(fields, "HTTP request")
		}
	}
}
