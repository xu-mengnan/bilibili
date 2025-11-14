package bilibili

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// BilibiliClient Bilibili API客户端
type BilibiliClient struct {
	client *http.Client
}

// NewBilibiliClient 创建新的Bilibili客户端
func NewBilibiliClient() *BilibiliClient {
	return &BilibiliClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendRequest 发送HTTP GET请求
func (c *BilibiliClient) SendRequest(url string) ([]byte, error) {
	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置通用请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://www.bilibili.com/")

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return body, nil
}
