package bilibili

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// UserInfo 用户信息
type UserInfo struct {
	ID       int64  `json:"mid"`
	Name     string `json:"name"`
	Sex      string `json:"sex"`
	Sign     string `json:"sign"`
	Level    int    `json:"level"`
	Face     string `json:"face"` // 头像URL
	Coins    int    `json:"coins"`
	Birthday string `json:"birthday"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    UserInfo `json:"data"`
}

// GetUser 获取用户信息
func GetUser(mid int64) (*UserResponse, error) {
	// 构造API URL
	apiURL := fmt.Sprintf("https://api.bilibili.com/x/space/acc/info?mid=%d", mid)

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
	var userResp UserResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &userResp, nil
}
