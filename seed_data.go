//go:build ignore
// +build ignore

package main

import (
	"TaruApp/database"
	"TaruApp/models"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "modernc.org/sqlite"
)

// 示例数据生成器
func main() {
	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer database.CloseDB()

	fmt.Println("开始生成示例数据...")

	// 创建板块
	boards := []models.CreateBoardRequest{
		{Name: "技术交流", Description: "技术相关的讨论板块"},
		{Name: "生活分享", Description: "分享生活中的点点滴滴"},
		{Name: "游戏天地", Description: "游戏爱好者的聚集地"},
		{Name: "学习园地", Description: "学习资源分享与讨论"},
		{Name: "美食天堂", Description: "美食制作与推荐"},
	}

	boardIDs := []int64{}
	for _, board := range boards {
		result, err := database.DB.Exec(
			"INSERT INTO boards (name, description) VALUES (?, ?)",
			board.Name, board.Description,
		)
		if err != nil {
			log.Println("创建板块失败:", err)
			continue
		}
		id, _ := result.LastInsertId()
		boardIDs = append(boardIDs, id)
		fmt.Printf("✓ 创建板块: %s (ID: %d)\n", board.Name, id)
	}

	// 创建帖子
	publishers := []string{"张三", "李四", "王五", "赵六", "钱七", "孙八", "周九", "吴十"}
	postTitles := [][]string{
		{"Go语言并发编程最佳实践", "深入理解Go接口设计", "Go性能优化技巧", "微服务架构实战分享"},
		{"周末爬山记录", "家庭装修心得", "养宠物的快乐时光", "读书分享：《活着》"},
		{"原神新版本体验", "英雄联盟S13赛季攻略", "塞尔达王国之泪通关心得", "Steam打折游戏推荐"},
		{"如何高效学习编程", "考研经验分享", "英语学习方法总结", "数学建模竞赛心得"},
		{"自制红烧肉教程", "探店：城市美食推荐", "烘焙入门指南", "健康饮食搭配建议"},
	}

	rand.Seed(time.Now().UnixNano())
	postIDs := []int64{}

	for i, boardID := range boardIDs {
		titles := postTitles[i]
		for j, title := range titles {
			publisher := publishers[rand.Intn(len(publishers))]
			content := fmt.Sprintf("这是关于「%s」的详细内容。在这篇文章中，我将分享我的经验和见解...", title)
			now := time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour) // 随机过去30天内的时间

			result, err := database.DB.Exec(
				`INSERT INTO posts (board_id, title, content, publisher, publish_time, 
				coins, favorites, likes, view_count, last_reply_time) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				boardID, title, content, publisher, now,
				rand.Intn(50), rand.Intn(100), rand.Intn(200), rand.Intn(1000), now,
			)
			if err != nil {
				log.Println("创建帖子失败:", err)
				continue
			}
			id, _ := result.LastInsertId()
			postIDs = append(postIDs, id)
			fmt.Printf("✓ 创建帖子: %s (ID: %d, 板块: %s)\n", title, id, boards[i].Name)

			// 为每个帖子创建随机数量的评论
			commentCount := rand.Intn(10) + 1
			for k := 1; k <= commentCount; k++ {
				commentPublisher := publishers[rand.Intn(len(publishers))]
				commentContent := fmt.Sprintf("这是第 %d 条评论。%s 说：很有启发！", k, commentPublisher)
				isAuthor := commentPublisher == publisher
				commentTime := now.Add(time.Duration(rand.Intn(24*30)) * time.Hour)

				_, err := database.DB.Exec(
					`INSERT INTO comments (post_id, content, publisher, publish_time, 
					is_author, floor, likes) VALUES (?, ?, ?, ?, ?, ?, ?)`,
					id, commentContent, commentPublisher, commentTime, isAuthor, k, rand.Intn(50),
				)
				if err == nil {
					// 更新帖子的评论数和最后回复时间
					database.DB.Exec(
						"UPDATE posts SET comment_count = comment_count + 1, last_reply_time = ? WHERE id = ?",
						commentTime, id,
					)
				}
			}
			fmt.Printf("  └─ 创建了 %d 条评论\n", commentCount)
		}
	}

	fmt.Println("\n========================================")
	fmt.Printf("示例数据生成完成！\n")
	fmt.Printf("- 板块数量: %d\n", len(boardIDs))
	fmt.Printf("- 帖子数量: %d\n", len(postIDs))
	fmt.Println("========================================")
	fmt.Println("\n现在可以启动服务器进行测试：")
	fmt.Println("  go run main.go")
	fmt.Println("\n或使用启动脚本：")
	fmt.Println("  Windows: start.bat")
	fmt.Println("  Linux/Mac: ./start.sh")
}
