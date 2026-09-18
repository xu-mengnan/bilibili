package services

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"bilibili/pkg/bilibili"
)

type VideoService struct{}

func NewVideoService() *VideoService {
	return &VideoService{}
}

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

func (vs *VideoService) ParseVideoInput(input string) (videoID string, videoType string, err error) {
	input = strings.TrimSpace(input)

	bvPattern := regexp.MustCompile(`(BV[a-zA-Z0-9]+)`)
	if matches := bvPattern.FindStringSubmatch(input); len(matches) > 0 {
		return matches[1], "bv", nil
	}

	avPattern := regexp.MustCompile(`[aA][vV](\d+)`)
	if matches := avPattern.FindStringSubmatch(input); len(matches) > 0 {
		return matches[1], "av", nil
	}

	if regexp.MustCompile(`^\d+$`).MatchString(input) {
		return input, "av", nil
	}
	return "", "", fmt.Errorf("invalid video ID format: %s", input)
}

func (vs *VideoService) GetVideoInfo(input string) (*VideoInfo, error) {
	return vs.GetVideoInfoContext(context.Background(), input)
}

func (vs *VideoService) GetVideoInfoContext(ctx context.Context, input string) (*VideoInfo, error) {
	videoID, videoType, err := vs.ParseVideoInput(input)
	if err != nil {
		return nil, err
	}

	var resp *bilibili.VideoResponse
	switch videoType {
	case "bv":
		resp, err = bilibili.GetVideoByBVIDContext(ctx, videoID)
	case "av":
		aid, parseErr := strconv.ParseInt(videoID, 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid AV id: %w", parseErr)
		}
		resp, err = bilibili.GetVideoByAIDContext(ctx, aid)
	default:
		return nil, fmt.Errorf("unsupported video type: %s", videoType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	return &VideoInfo{
		BVID:          resp.Data.BVID,
		AID:           resp.Data.AID,
		Title:         resp.Data.Title,
		Author:        resp.Data.Owner.Name,
		Views:         resp.Data.Stat.View,
		CommentsTotal: resp.Data.Stat.Reply,
		Likes:         resp.Data.Stat.Like,
		CreatedTime:   resp.Data.Created,
		PicURL:        resp.Data.Pic,
		Description:   resp.Data.Desc,
	}, nil
}
