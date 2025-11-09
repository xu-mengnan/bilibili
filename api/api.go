package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bilibili/internal/handlers"
	"bilibili/internal/services"
	"bilibili/pkg/utils"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to Bilibili API")
	})

	r.GET("/hello", func(c *gin.Context) {
		handlers.GinHelloHandler(c)
	})

	r.GET("/user/:id", func(c *gin.Context) {
		// 获取URL中的用户ID
		userIDStr := c.Param("id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// 获取用户信息
		user, err := services.GetUserByID(userID)
		if err != nil {
			utils.LogError("Failed to get user: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// 返回JSON格式的用户信息
		c.JSON(http.StatusOK, user)
	})

	return r
}
