package api

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bilibili/internal/config"
	"bilibili/internal/handlers"
	"bilibili/internal/handlers/middleware"
	svc "bilibili/internal/services"
	"bilibili/pkg/storage"
	"bilibili/pkg/utils"
)

// Services 所有服务实例
type Services struct {
	CommentService  *svc.CommentService
	ExportService   *svc.ExportService
	AnalysisService *svc.AnalysisService
}

// SetupRoutes 设置路由
func SetupRoutes(ctx context.Context) (*gin.Engine, *Services) {
	r := gin.New() // 不使用默认中间件，手动注册

	// 注册全局中间件
	r.Use(middleware.Recovery()) // Panic 恢复
	r.Use(middleware.Logging())  // 请求日志
	r.Use(middleware.CORS())     // 跨域支持

	// 加载配置
	cfg, err := config.LoadDefault()
	if err != nil {
		log.Printf("警告: 加载配置文件失败: %v，使用默认配置", err)
		cfg, _ = config.LoadDefault()
	}

	// 初始化服务
	// 初始化存储层
	taskStorage := storage.NewJSONStorage(cfg.Storage.DataDir)

	// 初始化服务（传递 context）
	commentService := svc.NewCommentService(ctx, taskStorage)
	videoService := svc.NewVideoService()
	exportService := svc.NewExportService(ctx, "./exports")
	analysisService := svc.NewAnalysisService(
		cfg.AI.APIURL,
		cfg.AI.APIKey,
		cfg.AI.Model,
	)

	services := &Services{
		CommentService:  commentService,
		ExportService:   exportService,
		AnalysisService: analysisService,
	}

	// 初始化处理器
	commentHandlers := handlers.NewCommentHandlers(commentService, exportService)
	videoHandlers := handlers.NewVideoHandlers(videoService)
	analysisHandlers := handlers.NewAnalysisHandlers(commentService, analysisService)
	v2Handlers := handlers.NewV2Handlers(commentService, analysisService)
	healthHandler := handlers.NewHealthHandler()

	// 静态文件服务
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/tasks", "./static/tasks.html")
	r.StaticFile("/analysis", "./static/analysis.html")

	// 健康检查
	r.GET("/health", healthHandler.HealthCheck)

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
		user, err := svc.GetUserByID(userID)
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
		apiGroup.GET("/tasks/all", commentHandlers.GetAllTasksHandler) // 获取所有任务

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

	// V2 API - 为新版前端页面服务（更简洁的响应格式）
	v2Group := r.Group("/api/v2")
	{
		// 任务相关
		v2Group.GET("/tasks", v2Handlers.GetTasksHandler)
		v2Group.GET("/tasks/:id", v2Handlers.GetTaskHandler)

		// 模板相关
		v2Group.GET("/templates", v2Handlers.GetTemplatesHandler)

		// 分析相关
		v2Group.POST("/analyze-stream", v2Handlers.AnalyzeStreamHandlerV2)
		v2Group.POST("/preview", v2Handlers.PreviewPromptHandlerV2)
	}

	return r, services
}

// ShutdownServices 关闭所有服务
func ShutdownServices(ctx context.Context, services *Services) {
	// 按照依赖顺序关闭服务

	// 1. 关闭 ExportService
	if err := services.ExportService.Shutdown(ctx); err != nil {
		utils.LogError("Failed to shutdown ExportService: " + err.Error())
	}

	// 2. 关闭 CommentService
	if err := services.CommentService.Shutdown(ctx); err != nil {
		utils.LogError("Failed to shutdown CommentService: " + err.Error())
	}

	utils.LogInfo("All services shutdown complete")
}
