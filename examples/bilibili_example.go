//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"bilibili/pkg/bilibili"
)

func main() {
	// 示例：获取视频信息
	fmt.Println("获取视频信息示例:")
	videoResp, err := bilibili.GetVideoByBVID("BV1shyYBjE9H") // 使用之前测试过的视频
	if err != nil {
		log.Printf("获取视频信息失败: %v", err)
	} else if videoResp.Code != 0 {
		log.Printf("获取视频信息失败，错误码: %d, 错误信息: %s", videoResp.Code, videoResp.Message)
	} else {
		fmt.Printf("视频标题: %s\n", videoResp.Data.Title)
		fmt.Printf("视频作者: %s\n", videoResp.Data.Owner.Name)
		fmt.Printf("播放量: %d\n", videoResp.Data.Stat.View)
		fmt.Printf("点赞数: %d\n", videoResp.Data.Stat.Like)
		fmt.Printf("评论数: %d\n", videoResp.Data.Stat.Reply)
		fmt.Printf("视频AID: %d\n", videoResp.Data.AID)
	}

	fmt.Println("\n" + "=========================" + "\n")

	// 示例：获取评论信息（无认证）
	fmt.Println("获取评论信息示例（无认证）:")
	// 注意：这里需要使用真实的视频aid，可以通过上面获取视频信息得到
	if videoResp != nil && videoResp.Code == 0 {
		// 分页获取评论
		allComments := []bilibili.CommentData{}
		pageSize := 20
		currentPage := 1
		hasMore := true

		fmt.Println("开始分页获取评论...")

		// 用于跟踪已获取的评论ID，防止重复
		seenComments := make(map[int64]bool)
		nextCursor := 0  // 用于翻页的游标
		nextOffset := "" // 用于翻页的 offset 字符串
		for hasMore {
			if currentPage > 10 { // 限制最多获取10页
				break
			}
			fmt.Printf("正在获取第 %d 页评论... (cursor: %d, offset: %s)\n", currentPage, nextCursor, nextOffset)

			// 优先使用 nextOffset，如果没有则使用 nextCursor
			var commentsResp *bilibili.CommentResponse
			var err error
			if nextOffset != "" {
				commentsResp, err = bilibili.GetCommentsWithOffset(videoResp.Data.AID, currentPage, pageSize, nextCursor, nextOffset)
			} else {
				commentsResp, err = bilibili.GetComments(videoResp.Data.AID, currentPage, pageSize, nextCursor)
			}

			if err != nil {
				log.Printf("获取第 %d 页评论失败: %v", currentPage, err)
				break
			} else if commentsResp.Code != 0 {
				log.Printf("获取第 %d 页评论失败，错误码: %d, 错误信息: %s", currentPage, commentsResp.Code, commentsResp.Message)
				break
			} else {
				// 更新 nextCursor 和 nextOffset 为本次响应返回的值，用于下次请求
				nextCursor = commentsResp.Data.Cursor.Next
				nextOffset = commentsResp.Data.Cursor.PaginationReply.NextOffset
				fmt.Printf("本次响应返回 - next cursor: %d, next_offset: %s\n", nextCursor, nextOffset)

				// 添加当前页的评论到总列表（去重）
				newComments := 0
				for _, comment := range commentsResp.Data.Replies {
					if !seenComments[comment.RPID] {
						seenComments[comment.RPID] = true
						allComments = append(allComments, comment)
						newComments++
					}
				}

				fmt.Printf("第 %d 页获取到 %d 条评论 (%d 条新评论)\n", currentPage, len(commentsResp.Data.Replies), newComments)

				// 判断是否还有更多评论
				// 如果 nextCursor 为 0 且 nextOffset 为空，或者返回的评论数少于页面大小，说明没有更多评论了
				if (nextCursor == 0 && nextOffset == "") || len(commentsResp.Data.Replies) < pageSize {
					hasMore = false
					fmt.Printf("没有更多评论了 (nextCursor: %d, nextOffset: %s, 返回评论数: %d)\n", nextCursor, nextOffset, len(commentsResp.Data.Replies))
				} else {
					// 继续下一页
					currentPage++
				}
			}
			// 添加短暂延迟，避免请求过于频繁
			time.Sleep(500 * time.Millisecond)
		}

		// 显示获取到的评论
		fmt.Printf("总共获取到 %d 条不重复评论:\n", len(allComments))
		for i, comment := range allComments {
			fmt.Printf("%d. %s: %s (点赞: %d)\n", i+1, comment.Member.Uname, comment.Content.Message, comment.Like)
		}
		// 显示总评论数
		fmt.Printf("总评论数: %d\n", len(allComments))
	}

}
