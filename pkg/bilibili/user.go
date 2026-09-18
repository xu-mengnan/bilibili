package bilibili

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

func GetUser(mid int64) (*UserResponse, error) {
	return GetUserContext(context.Background(), mid)
}

func GetUserContext(ctx context.Context, mid int64) (*UserResponse, error) {
	client := DefaultClient()
	params := url.Values{}
	params.Set("mid", fmt.Sprintf("%d", mid))

	body, err := client.SendRequestContext(ctx, client.userInfoURL+"?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var resp UserResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析用户 JSON 失败: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("用户 API 返回错误，错误码: %d, 错误信息: %s", resp.Code, resp.Message)
	}
	return &resp, nil
}
