package main

import (
	"fmt"
	"log"

	"bilibili/pkg/bilibili"
	"bilibili/pkg/file"
)

func main() {
	// 示例：获取视频信息
	fmt.Println("获取视频信息示例:")
	videoResp, err := bilibili.GetVideoByBVID("BV1Qr1QBLEL5") // 使用之前测试过的视频
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

	// 示例：获取评论信息
	fmt.Println("获取评论信息示例:")
	// 注意：这里需要使用真实的视频aid，可以通过上面获取视频信息得到
	if videoResp != nil && videoResp.Code == 0 {
		commentsResp, err := bilibili.GetComments(videoResp.Data.AID, 1, 10)
		if err != nil {
			log.Printf("获取评论失败: %v", err)
		} else if commentsResp.Code != 0 {
			log.Printf("获取评论失败，错误码: %d, 错误信息: %s", commentsResp.Code, commentsResp.Message)
		} else {
			fmt.Printf("获取到 %d 条评论:\n", len(commentsResp.Data.Replies))
			for i, comment := range commentsResp.Data.Replies {
				// 显示前3条评论
				if i < 3 {
					fmt.Printf("%d. %s: %s (点赞: %d)\n", i+1, comment.Member.Uname, comment.Content.Message, comment.Like)
				}
			}
			// 显示总评论数
			fmt.Printf("总评论数: %d\n", commentsResp.Data.Page.Count)
		}
	}

	fmt.Println("\n" + "=========================" + "\n")

	// 示例：获取所有评论并保存到Excel
	fmt.Println("获取所有评论并保存到Excel示例:")
	if videoResp != nil && videoResp.Code == 0 {
		fmt.Println("正在获取所有评论，请稍候...")
		allComments, err := bilibili.GetAllComments(videoResp.Data.AID)
		if err != nil {
			log.Printf("获取所有评论失败: %v", err)
		} else {
			fmt.Printf("总共获取到 %d 条不重复的评论:\n", len(allComments))
			// 显示前5条评论
			for i, comment := range allComments {
				if i < 5 {
					fmt.Printf("%d. %s: %s (点赞: %d)\n", i+1, comment.Member.Uname, comment.Content.Message, comment.Like)
				}
			}

			// 将评论数据保存到Excel文件
			if len(allComments) > 0 {
				// 准备Excel数据
				excelData := make([][]string, len(allComments)+1)

				// 添加表头
				excelData[0] = []string{"用户名", "评论内容", "点赞数", "评论时间"}

				// 添加评论数据
				for i, comment := range allComments {
					excelData[i+1] = []string{
						comment.Member.Uname,
						comment.Content.Message,
						fmt.Sprintf("%d", comment.Like),
						fmt.Sprintf("%d", comment.Ctime),
					}
				}

				// 保存到Excel文件
				filename := fmt.Sprintf("video_%d_comments.xlsx", videoResp.Data.AID)
				err := file.WriteExcel(excelData, filename)
				if err != nil {
					log.Printf("保存Excel文件失败: %v", err)
				} else {
					fmt.Printf("评论数据已保存到文件: %s\n", filename)
				}
			}
		}
	}
}
