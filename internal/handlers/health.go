package handlers

import (
	"net/http"

	"bilibili/internal/services"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	commentService *services.CommentService
}

func NewHealthHandler(commentService *services.CommentService) *HealthHandler {
	return &HealthHandler{commentService: commentService}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	queue := services.ScrapeQueueStats{}
	if h.commentService != nil {
		queue = h.commentService.QueueStats()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"services": gin.H{
			"comment_service":  "ok",
			"export_service":   "ok",
			"analysis_service": "ok",
			"video_service":    "ok",
		},
		"scrape_queue": queue,
	})
}
