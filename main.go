package main

import (
	"TaruApp/config"
	"TaruApp/database"
	"TaruApp/handlers"
	"TaruApp/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	config.InitConfig()

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer database.CloseDB()

	// 创建 Gin 路由
	r := gin.Default()

	// 使用中间件
	r.Use(middleware.Logger())
	if config.AppConfig.EnableCORS {
		r.Use(middleware.CORS())
	}
	r.Use(middleware.ErrorHandler())

	// 设置 JSON 响应格式
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// API 路由组
	api := r.Group("/api")
	{
		// 用户相关（不需要认证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register) // 用户注册
			auth.POST("/login", handlers.Login)       // 用户登录
		}

		// 应用市场路由（不需要认证）
		apps := api.Group("/apps")
		{
			apps.GET("", handlers.GetApps)                             // 获取应用列表
			apps.GET("/categories", handlers.GetMainCategories)        // 获取所有大分类
			apps.GET("/subcategories", handlers.GetSubCategories)      // 获取指定大分类下的小分类
			apps.GET("/category", handlers.GetAppsByCategory)          // 根据分类获取应用列表
			apps.GET("/:package_name", handlers.GetAppDetail)          // 获取应用详情
			apps.POST("/:package_name/download", handlers.DownloadApp) // 记录下载
		}

		// 需要认证的路由
		authorized := api.Group("")
		authorized.Use(middleware.AuthRequired())
		{
			// 当前用户信息
			authorized.GET("/me", handlers.GetCurrentUser) // 获取当前用户信息
			authorized.POST("/logout", handlers.Logout)    // 退出登录

			// 用户信息（公开）
			authorized.GET("/users/:id", handlers.GetUserInfo)          // 获取用户信息
			authorized.GET("/users/:id/detail", handlers.GetUserDetail) // 获取用户详情
			authorized.GET("/users/:id/tags", handlers.GetUserTags)     // 获取用户标签
			authorized.GET("/users/:id/stats", handlers.GetUserStats)   // 获取用户统计信息
			authorized.GET("/users", handlers.GetAllUsers)              // 获取所有用户列表

			// 收藏夹功能
			folders := authorized.Group("/folders")
			{
				folders.POST("/create", handlers.CreateFavoriteFolder)               // 创建收藏夹
				folders.GET("/my", handlers.GetMyFavoriteFolders)                    // 获取我的收藏夹列表
				folders.GET("/user/:id", handlers.GetUserFavoriteFolders)            // 获取用户的收藏夹列表
				folders.PUT("/:id", handlers.UpdateFavoriteFolder)                   // 更新收藏夹
				folders.DELETE("/:id", handlers.DeleteFavoriteFolder)                // 删除收藏夹
				folders.GET("/:id/posts", handlers.GetFolderPosts)                   // 获取收藏夹中的帖子
				folders.POST("/:id/posts", handlers.AddPostToFolder)                 // 添加帖子到收藏夹
				folders.DELETE("/:id/posts/:post_id", handlers.RemovePostFromFolder) // 从收藏夹移除帖子
			}

			// 浏览历史
			authorized.GET("/history", handlers.GetViewHistory) // 获取浏览历史

			// 应用市场（需要登录的部分）
			authorized.POST("/apps/:package_name/coin", handlers.CoinApp) // 给应用投币

			// 关注系统
			follow := authorized.Group("/follow")
			{
				follow.POST("/:id", handlers.FollowUser)                // 关注用户
				follow.DELETE("/:id", handlers.UnfollowUser)            // 取消关注用户
				follow.GET("/:id/following", handlers.GetFollowingList) // 获取关注列表
				follow.GET("/:id/followers", handlers.GetFollowerList)  // 获取粉丝列表
			}

			// 签到系统
			checkIn := authorized.Group("/checkin")
			{
				checkIn.POST("", handlers.CheckIn)                      // 每日签到
				checkIn.GET("/status", handlers.GetCheckInStatus)       // 获取签到状态
				checkIn.GET("/rank", handlers.GetCheckInRank)           // 获取签到排行榜
				checkIn.GET("/history/:id", handlers.GetCheckInHistory) // 获取用户签到历史
			}

			// 板块相关
			boards := authorized.Group("/boards")
			{
				boards.POST("/create", handlers.CreateBoard) // 创建板块
				boards.GET("/list", handlers.GetAllBoards)   // 获取所有板块
				boards.GET("/:id", handlers.GetBoardDetail)  // 获取板块详情
				boards.PUT("/:id", handlers.UpdateBoard)     // 更新板块
				boards.DELETE("/:id", handlers.DeleteBoard)  // 删除板块
			}

			// 帖子相关
			posts := authorized.Group("/posts")
			{
				posts.POST("/create", handlers.CreatePost)     // 创建帖子
				posts.GET("/list", handlers.GetPosts)          // 获取帖子列表（支持板块筛选和排序）
				posts.GET("/:id", handlers.GetPostDetail)      // 获取帖子详情
				posts.PUT("/:id", handlers.UpdatePost)         // 更新帖子
				posts.DELETE("/:id", handlers.DeletePost)      // 删除帖子
				posts.POST("/:id/like", handlers.LikePost)     // 点赞帖子
				posts.DELETE("/:id/like", handlers.UnlikePost) // 取消点赞帖子
				posts.POST("/:id/coin", handlers.CoinPost)     // 投币帖子
			}

			// 评论相关
			comments := authorized.Group("/comments")
			{
				comments.POST("/create", handlers.CreateComment)         // 创建评论（支持楼中楼回复）
				comments.GET("/list", handlers.GetComments)              // 获取评论列表（只显示顶级评论）
				comments.GET("/:id/replies", handlers.GetCommentReplies) // 获取评论的子回复列表
				comments.PUT("/:id", handlers.UpdateComment)             // 更新评论
				comments.DELETE("/:id", handlers.DeleteComment)          // 删除评论
				comments.POST("/:id/like", handlers.LikeComment)         // 点赞评论
				comments.POST("/:id/coin", handlers.CoinComment)         // 投币评论
			}

			// 统计相关
			stats := authorized.Group("/stats")
			{
				stats.GET("/boards/:id", handlers.GetBoardStats) // 获取板块统计
				stats.GET("/posts/:id", handlers.GetPostStats)   // 获取帖子统计
			}
		}

		// 管理员路由
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
		{
			admin.PUT("/users/:id/level", handlers.SetUserLevel)    // 设置用户等级
			admin.POST("/users/tags", handlers.CreateUserTag)       // 创建用户标签
			admin.DELETE("/users/tags/:id", handlers.DeleteUserTag) // 删除用户标签
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "TaruApp 服务器运行正常",
		})
	})

	log.Printf("TaruApp 服务器启动在端口 %s", config.AppConfig.ServerPort)
	log.Printf("访问地址: http://localhost:%s", config.AppConfig.ServerPort)
	if err := r.Run(":" + config.AppConfig.ServerPort); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
