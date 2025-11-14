package bilibili

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// VideoInfo 视频信息
type VideoInfo struct {
	BVID      string `json:"bvid"`
	AID       int64  `json:"aid"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	Created   int64  `json:"created"`
	Duration  int    `json:"duration"`
	Pic       string `json:"pic"`
	Owner     Owner  `json:"owner"`
	Stat      Stat   `json:"stat"`
	Copyright int    `json:"copyright"`
}

// Owner 视频所有者信息
type Owner struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

// Stat 视频统计数据
type Stat struct {
	Aid       int64  `json:"aid"`
	View      int    `json:"view"`
	Danmaku   int    `json:"danmaku"`
	Reply     int    `json:"reply"`
	Favorite  int    `json:"favorite"`
	Coin      int    `json:"coin"`
	Share     int    `json:"share"`
	Like      int    `json:"like"`
	Dislike   int    `json:"dislike"`
	NowRank   int    `json:"now_rank"`
	HisRank   int    `json:"his_rank"`
	NoReprint int    `json:"no_reprint"`
	Copyright int    `json:"copyright"`
	ArgueMsg  string `json:"argue_msg"`
}

// VideoResponse 视频信息响应
type VideoResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    VideoInfo `json:"data"`
}

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
