package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinHelloHandler 处理 /hello 路径的请求
func GinHelloHandler(c *gin.Context) {
	c.String(http.StatusOK, "Hello, Bilibili!")
}
