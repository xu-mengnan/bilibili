package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bilibili/internal/config"
	"bilibili/internal/handlers"
	"bilibili/internal/services"
	"bilibili/pkg/storage"
	"bilibili/pkg/utils"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// 加载配置
	cfg, err := config.LoadDefault()
	if err != nil {
		log.Printf("警告: 加载配置文件失败: %v，使用默认配置", err)
		cfg, _ = config.LoadDefault()
	}

	// 初始化服务
	// 初始化存储层
	taskStorage := storage.NewJSONStorage(cfg.Storage.DataDir)

	commentService := services.NewCommentService(taskStorage)
	videoService := services.NewVideoService()
	exportService := services.NewExportService("./exports")
	analysisService := services.NewAnalysisService(
		cfg.AI.APIURL,
		cfg.AI.APIKey,
		cfg.AI.Model,
	)

	// 初始化处理器
	commentHandlers := handlers.NewCommentHandlers(commentService, exportService)
	videoHandlers := handlers.NewVideoHandlers(videoService)
	analysisHandlers := handlers.NewAnalysisHandlers(commentService, analysisService)

	// 静态文件服务
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/analysis", "./static/analysis.html")

	// 原有路由
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

	// API路由组
	apiGroup := r.Group("/api")
	{
		// 评论相关
		apiGroup.POST("/comments/scrape", commentHandlers.ScrapeCommentsHandler)
		apiGroup.GET("/comments/progress/:task_id", commentHandlers.GetProgressHandler)
		apiGroup.GET("/comments/result/:task_id", commentHandlers.GetResultHandler)
		apiGroup.POST("/comments/export", commentHandlers.ExportCommentsHandler)
		apiGroup.GET("/comments/stats/:task_id", commentHandlers.GetCommentsStatsHandler)

		// 下载文件
		apiGroup.GET("/download/:file_id", commentHandlers.DownloadFileHandler)

		// 视频相关
		apiGroup.POST("/videos/info", videoHandlers.GetVideoInfoHandler)

		// AI分析相关
		apiGroup.GET("/analysis/templates", analysisHandlers.GetTemplatesHandler)
		apiGroup.POST("/analysis/analyze", analysisHandlers.AnalyzeHandler)
		apiGroup.POST("/analysis/analyze-stream", analysisHandlers.AnalyzeStreamHandler)
		apiGroup.GET("/analysis/tasks/completed", analysisHandlers.CompletedTasksHandler)
		apiGroup.GET("/analysis/tasks/:task_id", analysisHandlers.GetCommentsForAnalysisHandler)
		apiGroup.POST("/analysis/preview", analysisHandlers.PreviewPromptHandler)
	}

	return r
}
