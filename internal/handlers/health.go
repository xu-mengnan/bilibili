package handlers

import (
	"net/http"

	"bilibili/internal/services"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	commentService  *services.CommentService
	exportService   *services.ExportService
	analysisService *services.AnalysisService
}

func NewHealthHandler(
	commentService *services.CommentService,
	exportService *services.ExportService,
	analysisService *services.AnalysisService,
) *HealthHandler {
	return &HealthHandler{
		commentService:  commentService,
		exportService:   exportService,
		analysisService: analysisService,
	}
}

func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	status := http.StatusOK
	checks := gin.H{}

	if h.commentService == nil {
		status = http.StatusServiceUnavailable
		checks["comment_service"] = gin.H{"status": "error", "error": "unavailable"}
	} else if err := h.commentService.Ready(); err != nil {
		status = http.StatusServiceUnavailable
		checks["comment_service"] = gin.H{"status": "error", "error": err.Error()}
	} else {
		checks["comment_service"] = gin.H{"status": "ok"}
	}

	if h.exportService == nil {
		status = http.StatusServiceUnavailable
		checks["export_service"] = gin.H{"status": "error", "error": "unavailable"}
	} else if err := h.exportService.Ready(); err != nil {
		status = http.StatusServiceUnavailable
		checks["export_service"] = gin.H{"status": "error", "error": err.Error()}
	} else {
		checks["export_service"] = gin.H{"status": "ok"}
	}

	analysisMode := "unavailable"
	if h.analysisService == nil {
		status = http.StatusServiceUnavailable
		checks["analysis_service"] = gin.H{"status": "error", "error": "unavailable"}
	} else {
		analysisMode = h.analysisService.Mode()
		if err := h.analysisService.Ready(); err != nil {
			status = http.StatusServiceUnavailable
			checks["analysis_service"] = gin.H{"status": "error", "error": err.Error(), "mode": analysisMode}
		} else {
			checks["analysis_service"] = gin.H{"status": "ok", "mode": analysisMode}
		}
	}

	queue := services.ScrapeQueueStats{}
	if h.commentService != nil {
		queue = h.commentService.QueueStats()
	}

	overall := "ready"
	if status != http.StatusOK {
		overall = "not_ready"
	}

	c.JSON(status, gin.H{
		"status":       overall,
		"services":     checks,
		"scrape_queue": queue,
		"analysis_mode": analysisMode,
	})
}

// HealthCheck is kept as a compatibility alias for readiness.
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	h.Ready(c)
}
