package bilibili

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

func GetVideoByBVID(bvid string) (*VideoResponse, error) {
	return GetVideoByBVIDContext(context.Background(), bvid)
}

func GetVideoByBVIDContext(ctx context.Context, bvid string) (*VideoResponse, error) {
	return getVideoByBVIDContext(ctx, DefaultClient(), bvid)
}

func getVideoByBVIDContext(ctx context.Context, client *BilibiliClient, bvid string) (*VideoResponse, error) {
	params := url.Values{}
	params.Set("bvid", bvid)

	body, err := client.SendRequestContext(ctx, client.videoViewURL+"?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var resp VideoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析视频 JSON 失败: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("视频 API 返回错误，错误码: %d, 错误信息: %s", resp.Code, resp.Message)
	}
	return &resp, nil
}

func GetVideoByAID(aid int64) (*VideoResponse, error) {
	return GetVideoByAIDContext(context.Background(), aid)
}

func GetVideoByAIDContext(ctx context.Context, aid int64) (*VideoResponse, error) {
	return getVideoByAIDContext(ctx, DefaultClient(), aid)
}

func getVideoByAIDContext(ctx context.Context, client *BilibiliClient, aid int64) (*VideoResponse, error) {
	params := url.Values{}
	params.Set("aid", fmt.Sprintf("%d", aid))

	body, err := client.SendRequestContext(ctx, client.videoViewURL+"?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var resp VideoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析视频 JSON 失败: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("视频 API 返回错误，错误码: %d, 错误信息: %s", resp.Code, resp.Message)
	}
	return &resp, nil
}
