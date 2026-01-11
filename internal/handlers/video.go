package handlers

import (
	"net/http"

	"bilibili/internal/services"
	"github.com/gin-gonic/gin"
)

// VideoHandlers 视频处理器集合
type VideoHandlers struct {
	videoService *services.VideoService
}

// NewVideoHandlers 创建视频处理器
func NewVideoHandlers(videoService *services.VideoService) *VideoHandlers {
	return &VideoHandlers{
		videoService: videoService,
	}
}

// VideoInfoRequest 视频信息请求
type VideoInfoRequest struct {
	VideoURLOrID string `json:"video_url_or_id" binding:"required"`
}

// GetVideoInfoHandler 获取视频信息
func (h *VideoHandlers) GetVideoInfoHandler(c *gin.Context) {
	var req VideoInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	videoInfo, err := h.videoService.GetVideoInfo(req.VideoURLOrID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, videoInfo)
}
