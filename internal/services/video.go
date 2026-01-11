package services

import (
	"fmt"
	"regexp"
	"strings"

	"bilibili/pkg/bilibili"
)

// VideoService 视频服务
type VideoService struct{}

// NewVideoService 创建视频服务
func NewVideoService() *VideoService {
	return &VideoService{}
}

// VideoInfo 视频信息
type VideoInfo struct {
	BVID          string `json:"bvid"`
	AID           int64  `json:"aid"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Views         int    `json:"views"`
	CommentsTotal int    `json:"comments_total"`
	Likes         int    `json:"likes"`
	CreatedTime   int64  `json:"created_time"`
	PicURL        string `json:"pic_url"`
	Description   string `json:"description"`
}

// ParseVideoInput 解析视频输入（支持BV号、AV号、URL）
func (vs *VideoService) ParseVideoInput(input string) (videoID string, videoType string, err error) {
	input = strings.TrimSpace(input)

	// 匹配 BV 号（直接输入或URL中）
	bvPattern := regexp.MustCompile(`(BV[a-zA-Z0-9]+)`)
	if matches := bvPattern.FindStringSubmatch(input); len(matches) > 0 {
		return matches[1], "bv", nil
	}

	// 匹配 AV 号（直接输入或URL中）
	avPattern := regexp.MustCompile(`[aA][vV](\d+)`)
	if matches := avPattern.FindStringSubmatch(input); len(matches) > 0 {
		return "av" + matches[1], "av", nil
	}

	// 匹配纯数字（假定为AV号）
	numPattern := regexp.MustCompile(`^\d+$`)
	if numPattern.MatchString(input) {
		return "av" + input, "av", nil
	}

	return "", "", fmt.Errorf("invalid video ID format: %s", input)
}

// GetVideoInfo 获取视频信息
func (vs *VideoService) GetVideoInfo(input string) (*VideoInfo, error) {
	videoID, videoType, err := vs.ParseVideoInput(input)
	if err != nil {
		return nil, err
	}

	var videoResp *bilibili.VideoResponse

	// 根据类型获取视频信息
	if videoType == "bv" {
		videoResp, err = bilibili.GetVideoByBVID(videoID)
	} else {
		// AV号需要去掉"av"前缀
		videoResp, err = bilibili.GetVideoByBVID(videoID) // 优先尝试BV
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %v", err)
	}

	if videoResp.Code != 0 {
		return nil, fmt.Errorf("video API error: %s", videoResp.Message)
	}

	// 构造VideoInfo
	info := &VideoInfo{
		BVID:        videoResp.Data.BVID,
		AID:         videoResp.Data.AID,
		Title:       videoResp.Data.Title,
		Author:      videoResp.Data.Owner.Name,
		Views:       videoResp.Data.Stat.View,
		Likes:       videoResp.Data.Stat.Like,
		PicURL:      videoResp.Data.Pic,
		Description: videoResp.Data.Desc,
	}

	return info, nil
}
