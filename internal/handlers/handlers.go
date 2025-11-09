package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HelloHandler 处理 /hello 路径的请求 (标准http)
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, Bilibili!"))
}

// GinHelloHandler 处理 /hello 路径的请求 (Gin框架)
func GinHelloHandler(c *gin.Context) {
	c.String(http.StatusOK, "Hello, Bilibili!")
}