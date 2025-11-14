package bilibili

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// GetVideoByBVID 通过BVID获取视频信息
func GetVideoByBVID(bvid string) (*VideoResponse, error) {
	// 构造API URL
	apiURL := "https://api.bilibili.com/x/web-interface/view"

	// 构造查询参数
	params := url.Values{}
	params.Add("bvid", bvid)

	// 完整URL
	fullURL := apiURL + "?" + params.Encode()

	// 使用公共客户端发送请求
	client := NewBilibiliClient()
	body, err := client.SendRequest(fullURL)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var videoResp VideoResponse
	if err := json.Unmarshal(body, &videoResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 检查API是否返回错误
	if videoResp.Code != 0 {
		return nil, fmt.Errorf("API返回错误，错误码: %d, 错误信息: %s", videoResp.Code, videoResp.Message)
	}

	return &videoResp, nil
}

// GetVideoByAID 通过AID获取视频信息
func GetVideoByAID(aid int64) (*VideoResponse, error) {
	// 构造API URL
	apiURL := "https://api.bilibili.com/x/web-interface/view"

	// 构造查询参数
	params := url.Values{}
	params.Add("aid", fmt.Sprintf("%d", aid))

	// 完整URL
	fullURL := apiURL + "?" + params.Encode()

	// 使用公共客户端发送请求
	client := NewBilibiliClient()
	body, err := client.SendRequest(fullURL)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var videoResp VideoResponse
	if err := json.Unmarshal(body, &videoResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 检查API是否返回错误
	if videoResp.Code != 0 {
		return nil, fmt.Errorf("API返回错误，错误码: %d, 错误信息: %s", videoResp.Code, videoResp.Message)
	}

	return &videoResp, nil
}
