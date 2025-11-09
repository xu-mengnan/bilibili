package bilibili

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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
	apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发起请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON
	var videoResp VideoResponse
	if err := json.Unmarshal(body, &videoResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &videoResp, nil
}

// GetVideoByAID 通过AID获取视频信息
func GetVideoByAID(aid int64) (*VideoResponse, error) {
	// 构造API URL
	apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?aid=%d", aid)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发起请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON
	var videoResp VideoResponse
	if err := json.Unmarshal(body, &videoResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &videoResp, nil
}
