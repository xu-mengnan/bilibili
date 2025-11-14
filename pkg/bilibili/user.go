package bilibili

import (
	"encoding/json"
	"fmt"
	"net/url"
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
	apiURL := "https://api.bilibili.com/x/space/acc/info"

	// 构造查询参数
	params := url.Values{}
	params.Add("mid", fmt.Sprintf("%d", mid))

	// 完整URL
	fullURL := apiURL + "?" + params.Encode()

	// 使用公共客户端发送请求
	client := NewBilibiliClient()
	body, err := client.SendRequest(fullURL)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var userResp UserResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 检查API是否返回错误
	if userResp.Code != 0 {
		return nil, fmt.Errorf("API返回错误，错误码: %d, 错误信息: %s", userResp.Code, userResp.Message)
	}

	return &userResp, nil
}
