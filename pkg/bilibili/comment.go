package bilibili

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// AuthOption 认证选项类型
type AuthOption func(*BilibiliClient)

// WithCookie Cookie认证选项
func WithCookie(sessdata string) AuthOption {
	return func(client *BilibiliClient) {
		client.SetCookies(map[string]string{
			"SESSDATA": sessdata,
		})
	}
}

// WithAppAuth APP认证选项
func WithAppAuth(appkey, appsec string) AuthOption {
	return func(client *BilibiliClient) {
		client.SetAppAuth(appkey, appsec)
	}
}

// GetComments 获取视频评论 (使用wbi/main端点)
// next: 用于翻页的游标值（从上一页响应的 Cursor.Next 获取）
// nextOffset: 可选的 next_offset 字符串（从上一页响应的 Cursor.PaginationReply.NextOffset 获取）
func GetComments(oid int64, pn int, ps int, next int, authOptions ...AuthOption) (*CommentResponse, error) {
	return GetCommentsWithOffset(oid, pn, ps, next, "", authOptions...)
}

// GetCommentsWithOffset 获取视频评论（支持 next_offset 字符串）
func GetCommentsWithOffset(oid int64, pn int, ps int, next int, nextOffset string, authOptions ...AuthOption) (*CommentResponse, error) {
	// 构造API URL (使用wbi/main端点)
	apiURL := "https://api.bilibili.com/x/v2/reply/main"

	// 构造查询参数
	params := url.Values{}
	params.Add("oid", fmt.Sprintf("%d", oid))

	// 构造 pagination_str 参数
	if pn == 1 || (next == 0 && nextOffset == "") {
		// 第一页时 pagination_str 使用默认格式
		params.Add("pagination_str", `{"offset":"{\"type\":1,\"direction\":1,\"data\":{}}"}`)
	} else if nextOffset != "" {
		// 如果提供了 next_offset 字符串，直接使用它
		// next_offset 可能已经是 JSON 编码的字符串，需要检查格式
		// 如果 nextOffset 已经是完整的 pagination_str 格式，直接使用
		// 否则包装成正确的格式
		var paginationStr string
		if len(nextOffset) > 0 && nextOffset[0] == '{' {
			// 如果已经是 JSON 对象格式，直接使用
			paginationStr = nextOffset
		} else {
			// 否则包装成 {"offset":"..."} 格式，需要对 nextOffset 进行 JSON 转义
			// 使用 JSON 编码确保特殊字符被正确转义
			offsetJSON, _ := json.Marshal(nextOffset)
			paginationStr = fmt.Sprintf(`{"offset":%s}`, string(offsetJSON))
		}
		params.Add("pagination_str", paginationStr)
	} else {
		// 使用 next 值构造 pagination_str
		// 尝试使用 cursor 格式
		paginationStr := fmt.Sprintf(`{"offset":"{\"type\":1,\"direction\":1,\"data\":{\"cursor\":%d}}"}`, next)
		params.Add("pagination_str", paginationStr)
	}

	params.Add("type", "1") // 视频评论类型
	params.Add("mode", "2") // 按时间排序

	// 使用公共客户端发送请求
	client := NewBilibiliClient()

	// 应用认证选项
	for _, option := range authOptions {
		option(client)
	}

	// 获取WBI密钥并签名参数
	wbiKey := GetWBIKey()
	signedParams := SignParams(params, wbiKey)

	// 完整URL
	fullURL := apiURL + "?" + signedParams.Encode()

	body, err := client.SendRequest(fullURL)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var commentResp CommentResponse
	if err := json.Unmarshal(body, &commentResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 检查API是否返回错误
	if commentResp.Code != 0 {
		// 如果是权限错误，尝试使用备用接口
		if commentResp.Code == -403 {
			fmt.Println("权限错误，尝试使用备用接口获取评论...")
			return GetCommentsFallback(oid, pn, ps, authOptions...)
		}
		return nil, fmt.Errorf("API返回错误，错误码: %d, 错误信息: %s", commentResp.Code, commentResp.Message)
	}

	return &commentResp, nil
}

// GetCommentsFallback 备用方法，使用原始reply接口
func GetCommentsFallback(oid int64, pn int, ps int, authOptions ...AuthOption) (*CommentResponse, error) {
	// 构造API URL (使用原始reply接口)
	apiURL := "https://api.bilibili.com/x/v2/reply"

	// 构造查询参数
	params := url.Values{}
	params.Add("oid", fmt.Sprintf("%d", oid))
	params.Add("pn", fmt.Sprintf("%d", pn)) // fallback接口使用pn参数
	params.Add("ps", fmt.Sprintf("%d", ps))
	params.Add("type", "1") // 视频评论类型
	params.Add("sort", "2") // 按时间倒序排序

	// 使用公共客户端发送请求
	client := NewBilibiliClient()

	// 应用认证选项
	for _, option := range authOptions {
		option(client)
	}

	// 获取WBI密钥并签名参数
	wbiKey := GetWBIKey()
	signedParams := SignParams(params, wbiKey)

	// 完整URL
	fullURL := apiURL + "?" + signedParams.Encode()

	body, err := client.SendRequest(fullURL)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var commentResp CommentResponse
	if err := json.Unmarshal(body, &commentResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 检查API是否返回错误
	if commentResp.Code != 0 {
		return nil, fmt.Errorf("API返回错误，错误码: %d, 错误信息: %s", commentResp.Code, commentResp.Message)
	}

	return &commentResp, nil
}

// GetHotComments 获取视频的热门评论 (使用main端点)
func GetHotComments(oid int64, pn int, ps int, authOptions ...AuthOption) (*CommentResponse, error) {
	// 构造API URL (使用main端点获取热门评论)
	apiURL := "https://api.bilibili.com/x/v2/reply/main"

	// 构造查询参数
	params := url.Values{}
	params.Add("oid", fmt.Sprintf("%d", oid))
	params.Add("pn", fmt.Sprintf("%d", pn))
	params.Add("ps", fmt.Sprintf("%d", ps))
	params.Add("type", "1") // 视频评论类型

	// 获取WBI密钥并签名参数
	wbiKey := GetWBIKey()
	signedParams := SignParams(params, wbiKey)

	// 完整URL
	fullURL := apiURL + "?" + signedParams.Encode()

	// 使用公共客户端发送请求
	client := NewBilibiliClient()

	// 应用认证选项
	for _, option := range authOptions {
		option(client)
	}

	body, err := client.SendRequest(fullURL)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var commentResp CommentResponse
	if err := json.Unmarshal(body, &commentResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 检查API是否返回错误
	if commentResp.Code != 0 {
		return nil, fmt.Errorf("API返回错误，错误码: %d, 错误信息: %s", commentResp.Code, commentResp.Message)
	}

	return &commentResp, nil
}

// GetAllComments 获取视频的所有评论
func GetAllComments(oid int64, authOptions ...AuthOption) ([]CommentData, error) {
	// 使用map来去重，以RPID为键
	uniqueComments := make(map[int64]CommentData)
	var allComments []CommentData

	// 先使用main接口获取第一页热门评论（这个接口似乎更稳定）
	firstPage, err := GetHotComments(oid, 1, 20, authOptions...)
	if err != nil {
		return nil, fmt.Errorf("获取第一页热门评论失败: %v", err)
	}

	// 添加第一页评论到结果中（去重）
	for _, comment := range firstPage.Data.Replies {
		if _, exists := uniqueComments[comment.RPID]; !exists {
			uniqueComments[comment.RPID] = comment
			allComments = append(allComments, comment)
		}
	}

	// 计算总评论数
	totalCount := firstPage.Data.Cursor.AllCount
	pageSize := 20
	totalPages := (totalCount + pageSize - 1) / pageSize // 向上取整

	fmt.Printf("总评论数: %d, 总页数: %d\n", totalCount, totalPages)

	// 获取剩余页的评论
	for page := 2; page <= totalPages && page <= 100; page++ { // 限制最多获取100页以避免过多请求
		// 添加延迟避免请求过于频繁
		time.Sleep(300 * time.Millisecond)

		fmt.Printf("正在获取第 %d 页热门评论...\n", page)

		resp, err := GetHotComments(oid, page, pageSize, authOptions...)
		if err != nil {
			// 如果某页获取失败，记录错误并继续获取下一页
			fmt.Printf("获取第%d页热门评论失败: %v\n", page, err)
			continue
		}

		// 输出调试信息
		fmt.Printf("第%d页返回%d条热门评论\n", page, len(resp.Data.Replies))

		// 添加评论到结果中（去重）
		addedCount := 0
		for _, comment := range resp.Data.Replies {
			if _, exists := uniqueComments[comment.RPID]; !exists {
				uniqueComments[comment.RPID] = comment
				allComments = append(allComments, comment)
				addedCount++
			}
		}

		fmt.Printf("第%d页获取到%d条不重复的热门评论\n", page, addedCount)

		// 如果某页没有返回数据，跳出循环
		if len(resp.Data.Replies) == 0 {
			fmt.Printf("第%d页没有返回数据，停止获取\n", page)
			break
		}
	}

	return allComments, nil
}
