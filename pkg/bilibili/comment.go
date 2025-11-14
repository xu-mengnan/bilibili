package bilibili

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// GetComments 获取视频评论 (使用原始reply接口)
func GetComments(oid int64, pn int, ps int) (*CommentResponse, error) {
	// 构造API URL (使用原始reply接口)
	apiURL := "https://api.bilibili.com/x/v2/reply"

	// 构造查询参数
	params := url.Values{}
	params.Add("oid", fmt.Sprintf("%d", oid))
	params.Add("pn", fmt.Sprintf("%d", pn))
	params.Add("ps", fmt.Sprintf("%d", ps))
	params.Add("type", "1") // 视频评论类型
	params.Add("sort", "2") // 按时间倒序排序

	// 完整URL
	fullURL := apiURL + "?" + params.Encode()

	// 使用公共客户端发送请求
	client := NewBilibiliClient()
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
func GetHotComments(oid int64, pn int, ps int) (*CommentResponse, error) {
	// 构造API URL (使用main端点获取热门评论)
	apiURL := "https://api.bilibili.com/x/v2/reply/wbi/main"

	// 构造查询参数
	params := url.Values{}
	params.Add("oid", fmt.Sprintf("%d", oid))
	params.Add("pn", fmt.Sprintf("%d", pn))
	params.Add("ps", fmt.Sprintf("%d", ps))
	params.Add("type", "1") // 视频评论类型

	// 完整URL
	fullURL := apiURL + "?" + params.Encode()

	// 使用公共客户端发送请求
	client := NewBilibiliClient()
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
func GetAllComments(oid int64) ([]CommentData, error) {
	// 使用map来去重，以RPID为键
	uniqueComments := make(map[int64]CommentData)
	var allComments []CommentData

	// 先使用main接口获取第一页热门评论（这个接口似乎更稳定）
	firstPage, err := GetHotComments(oid, 1, 20)
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
	for page := 2; page <= totalPages && page <= 5; page++ { // 限制最多获取20页以避免过多请求
		// 添加延迟避免请求过于频繁
		time.Sleep(500 * time.Millisecond)

		fmt.Printf("正在获取第 %d 页热门评论...\n", page)

		resp, err := GetHotComments(oid, page, pageSize)
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
