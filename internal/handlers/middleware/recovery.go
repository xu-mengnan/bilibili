package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"bilibili/pkg/utils"
	"github.com/gin-gonic/gin"
)

// Recovery 捕获panic并优雅恢复的中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 信息
				stack := debug.Stack()
				utils.LogError(fmt.Sprintf("Panic recovered: %v\n%s", err, string(stack)))

				// 返回错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
