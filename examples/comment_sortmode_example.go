//go:build ignore
// +build ignore

package main

import (
	"bilibili/pkg/bilibili"
)

func main() {
	// 示例视频 BVID
	bvid := "BV1uT4y1P7CX"

	// 1. 获取视频信息
	fmt.Println("=== 获取视频信息 ===")
	videoResp, err := bilibili.GetVideoByBVID(bvid)
	if err != nil {
		fmt.Printf("获取视频信息失败: %v\n", err)
		return
	}

	if videoResp.Code != 0 {
		fmt.Printf("视频API返回错误: %s\n", videoResp.Message)
		return
	}

	fmt.Printf("视频标题: %s\n", videoResp.Data.Title)
	fmt.Printf("视频AID: %d\n", videoResp.Data.AID)
	fmt.Printf("作者: %s\n\n", videoResp.Data.Owner.Name)

	oid := videoResp.Data.AID

	// 2. 按时间排序获取评论（默认方式）
	fmt.Println("=== 按时间排序获取评论 ===")
	timeComments, err := bilibili.GetComments(oid, 1, 10, 0, bilibili.WithSortMode("time"))
	if err != nil {
		fmt.Printf("获取评论失败: %v\n", err)
		return
	}

	if timeComments.Code != 0 {
		fmt.Printf("评论API返回错误: %s\n", timeComments.Message)
		return
	}

	fmt.Printf("获取到 %d 条按时间排序的评论:\n", len(timeComments.Data.Replies))
	for i, comment := range timeComments.Data.Replies {
		if i >= 3 { // 只显示前3条
			break
		}
		fmt.Printf("%d. [%s] %s (点赞: %d)\n",
			i+1,
			comment.Member.Uname,
			comment.Content.Message,
			comment.Like,
		)
	}
	fmt.Println()

	// 3. 按热度排序获取评论
	fmt.Println("=== 按热度排序获取评论 ===")
	hotComments, err := bilibili.GetComments(oid, 1, 10, 0, bilibili.WithSortMode("hot"))
	if err != nil {
		fmt.Printf("获取热门评论失败: %v\n", err)
		return
	}

	if hotComments.Code != 0 {
		fmt.Printf("评论API返回错误: %s\n", hotComments.Message)
		return
	}

	fmt.Printf("获取到 %d 条按热度排序的评论:\n", len(hotComments.Data.Replies))
	for i, comment := range hotComments.Data.Replies {
		if i >= 3 { // 只显示前3条
			break
		}
		fmt.Printf("%d. [%s] %s (点赞: %d)\n",
			i+1,
			comment.Member.Uname,
			comment.Content.Message,
			comment.Like,
		)
	}
	fmt.Println()

	// 4. 也可以使用 GetHotComments (向后兼容的方式)
	fmt.Println("=== 使用 GetHotComments 获取热门评论 ===")
	hotComments2, err := bilibili.GetHotComments(oid, 1, 5)
	if err != nil {
		fmt.Printf("获取热门评论失败: %v\n", err)
		return
	}

	fmt.Printf("获取到 %d 条热门评论\n", len(hotComments2.Data.Replies))

	// 5. 演示如何结合认证和排序模式
	fmt.Println("\n=== 结合认证和排序模式 ===")
	// 如果有 Cookie，可以这样使用：
	// sessdata := "your_sessdata_here"
	// commentsWithAuth, err := bilibili.GetComments(
	//     oid, 1, 10, 0,
	//     bilibili.WithCookie(sessdata),
	//     bilibili.WithSortMode("hot"),
	// )

	fmt.Println("可以通过传入多个选项来同时设置认证和排序模式:")
	fmt.Println("bilibili.GetComments(oid, 1, 10, 0, bilibili.WithCookie(sessdata), bilibili.WithSortMode(\"hot\"))")
}
