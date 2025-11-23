package handlers

import (
	"TaruApp/database"
	"TaruApp/models"
	"TaruApp/utils"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreatePost 创建帖子
func CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 如果没有指定板块或板块ID为0，默认发到主板块(ID=1)
	if req.BoardID == 0 {
		req.BoardID = 1
	}

	// 如果没有指定类型或类型为空，默认为text
	if req.Type == "" {
		req.Type = "text"
	}

	// 验证类型参数
	if req.Type != "text" && req.Type != "markdown" {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "帖子类型只能是 'text' 或 'markdown'",
		})
		return
	}

	// 获取当前用户信息
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	now := time.Now()
	
	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建帖子失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()
	
	// 插入帖子
	result, err := tx.Exec(
		`INSERT INTO posts (board_id, user_id, title, content, type, publisher, publish_time, image_url, last_reply_time) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.BoardID, userID, req.Title, req.Content, req.Type, username, now, req.ImageURL, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建帖子失败: " + err.Error(),
		})
		return
	}

	// 增加用户经验值（发帖奖励5经验）
	rewardExp := 5
	_, err = tx.Exec(
		"UPDATE users SET exp = exp + ? WHERE id = ?",
		rewardExp, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "增加经验失败: " + err.Error(),
		})
		return
	}
	
	// 获取更新后的经验值
	var exp int
	err = tx.QueryRow("SELECT exp FROM users WHERE id = ?", userID).Scan(&exp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询经验值失败: " + err.Error(),
		})
		return
	}
	
	// 计算新的用户等级
	newUserLevel := utils.CalculateUserLevel(exp)
	
	// 更新用户等级
	_, err = tx.Exec(
		"UPDATE users SET user_level = ? WHERE id = ?",
		newUserLevel, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新用户等级失败: " + err.Error(),
		})
		return
	}
	
	// 提交事务
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建帖子失败: " + err.Error(),
		})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "创建帖子成功",
		Data: gin.H{
			"id":         id,
			"board_id":   req.BoardID,
			"reward_exp": rewardExp,
			"total_exp":  exp,
			"user_level": newUserLevel,
		},
	})
}

// GetPosts 获取帖子列表（支持板块筛选和排序）
func GetPosts(c *gin.Context) {
	var query models.GetPostsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 如果没有指定板块，默认显示主板块(ID=1)的帖子
	if query.BoardID == 0 {
		query.BoardID = 1
	}

	// 默认分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 构建查询语句
	baseQuery := "SELECT id, board_id, user_id, title, content, type, publisher, publish_time, coins, favorites, likes, image_url, attachment_url, attachment_type, comment_count, view_count, last_reply_time, created_at, updated_at FROM posts"
	countQuery := "SELECT COUNT(*) FROM posts"
	whereClause := " WHERE board_id = ?"
	orderClause := ""

	// 板块筛选（现在总是有board_id）
	args := []interface{}{query.BoardID}

	// 排序逻辑
	switch query.Sort {
	case "latest":
		orderClause = " ORDER BY publish_time DESC" // 最新发布
	case "reply":
		orderClause = " ORDER BY last_reply_time DESC" // 最近回复
	case "hot":
		orderClause = " ORDER BY (likes * 3 + favorites * 2 + coins * 5 + comment_count * 2 + view_count) DESC" // 热门（综合权重）
	default:
		orderClause = " ORDER BY publish_time DESC" // 默认最新发布
	}

	// 分页
	offset := (query.Page - 1) * query.PageSize
	limitClause := " LIMIT ? OFFSET ?"
	paginationArgs := append(args, query.PageSize, offset)

	// 查询总数
	var total int
	err := database.DB.QueryRow(countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询帖子总数失败: " + err.Error(),
		})
		return
	}

	// 查询帖子列表
	rows, err := database.DB.Query(baseQuery+whereClause+orderClause+limitClause, paginationArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询帖子列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var imageURL, attachmentURL, attachmentType sql.NullString
		err := rows.Scan(
			&post.ID, &post.BoardID, &post.UserID, &post.Title, &post.Content, &post.Type, &post.Publisher,
			&post.PublishTime, &post.Coins, &post.Favorites, &post.Likes,
			&imageURL, &attachmentURL, &attachmentType, &post.CommentCount, &post.ViewCount, &post.LastReplyTime,
			&post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}
		if attachmentURL.Valid {
			post.AttachmentURL = attachmentURL.String
		}
		if attachmentType.Valid {
			post.AttachmentType = attachmentType.String
		}
		posts = append(posts, post)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取帖子列表成功",
		Data: models.PageData{
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
			List:     posts,
		},
	})
}

// GetPostDetail 获取帖子详情
func GetPostDetail(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 增加浏览数
	database.DB.Exec("UPDATE posts SET view_count = view_count + 1 WHERE id = ?", id)

	// 记录浏览历史
	if userID != nil {
		database.DB.Exec(
			"INSERT INTO view_histories (user_id, post_id) VALUES (?, ?)",
			userID, id,
		)
	}

	var post models.Post
	var imageURL, attachmentURL, attachmentType sql.NullString
	err := database.DB.QueryRow(
		`SELECT id, board_id, user_id, title, content, type, publisher, publish_time, coins, favorites, likes, 
		image_url, attachment_url, attachment_type, comment_count, view_count, last_reply_time, created_at, updated_at 
		FROM posts WHERE id = ?`,
		id,
	).Scan(
		&post.ID, &post.BoardID, &post.UserID, &post.Title, &post.Content, &post.Type, &post.Publisher,
		&post.PublishTime, &post.Coins, &post.Favorites, &post.Likes,
		&imageURL, &attachmentURL, &attachmentType, &post.CommentCount, &post.ViewCount, &post.LastReplyTime,
		&post.CreatedAt, &post.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "帖子不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询帖子失败: " + err.Error(),
		})
		return
	}

	if imageURL.Valid {
		post.ImageURL = imageURL.String
	}
	if attachmentURL.Valid {
		post.AttachmentURL = attachmentURL.String
	}
	if attachmentType.Valid {
		post.AttachmentType = attachmentType.String
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取帖子详情成功",
		Data:    post,
	})
}

// UpdatePost 更新帖子
func UpdatePost(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查帖子是否存在，并验证是否为作者本人
	var postUserID int64
	err := database.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", id).Scan(&postUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "帖子不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询帖子失败: " + err.Error(),
		})
		return
	}

	// 验证权限：只有作者本人才能编辑
	if postUserID != userID.(int64) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权编辑此帖子，只能编辑自己的帖子",
		})
		return
	}

	// 验证并设置帖子类型
	if req.Type == "" {
		req.Type = "text"
	}
	if req.Type != "text" && req.Type != "markdown" {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "帖子类型只能是 'text' 或 'markdown'",
		})
		return
	}

	_, err = database.DB.Exec(
		"UPDATE posts SET title = ?, content = ?, type = ?, image_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		req.Title, req.Content, req.Type, req.ImageURL, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新帖子失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "更新帖子成功",
	})
}

// DeletePost 删除帖子
func DeletePost(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 检查帖子是否存在，并验证是否为作者本人
	var postUserID int64
	err := database.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", id).Scan(&postUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "帖子不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询帖子失败: " + err.Error(),
		})
		return
	}

	// 验证权限：只有作者本人才能删除
	if postUserID != userID.(int64) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权删除此帖子，只能删除自己的帖子",
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除帖子失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 删除帖子相关的所有数据
	// 1. 删除帖子的所有评论点赞记录
	_, err = tx.Exec("DELETE FROM comment_likes WHERE comment_id IN (SELECT id FROM comments WHERE post_id = ?)", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除评论点赞记录失败: " + err.Error(),
		})
		return
	}

	// 2. 删除帖子的所有评论
	_, err = tx.Exec("DELETE FROM comments WHERE post_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除评论失败: " + err.Error(),
		})
		return
	}

	// 3. 删除帖子点赞记录
	_, err = tx.Exec("DELETE FROM post_likes WHERE post_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除帖子点赞记录失败: " + err.Error(),
		})
		return
	}

	// 4. 删除收藏记录
	_, err = tx.Exec("DELETE FROM favorite_items WHERE post_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除收藏记录失败: " + err.Error(),
		})
		return
	}

	// 5. 删除浏览历史记录
	_, err = tx.Exec("DELETE FROM view_histories WHERE post_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除浏览历史失败: " + err.Error(),
		})
		return
	}

	// 6. 最后删除帖子本身
	_, err = tx.Exec("DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除帖子失败: " + err.Error(),
		})
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除帖子失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "删除帖子成功",
	})
}

// LikePost 点赞/取消点赞帖子（切换功能）
func LikePost(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 检查是否已点赞
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM post_likes WHERE user_id = ? AND post_id = ?",
		userID, id,
	).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询点赞状态失败: " + err.Error(),
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "操作失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	var message string
	var isLiked bool

	if count > 0 {
		// 已点赞，执行取消点赞
		_, err = tx.Exec(
			"DELETE FROM post_likes WHERE user_id = ? AND post_id = ?",
			userID, id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "取消点赞失败: " + err.Error(),
			})
			return
		}

		// 减少帖子点赞数
		_, err = tx.Exec("UPDATE posts SET likes = likes - 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND likes > 0", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "取消点赞失败: " + err.Error(),
			})
			return
		}

		message = "取消点赞成功"
		isLiked = false
	} else {
		// 未点赞，执行点赞
		_, err = tx.Exec(
			"INSERT INTO post_likes (user_id, post_id) VALUES (?, ?)",
			userID, id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "点赞失败: " + err.Error(),
			})
			return
		}

		// 增加帖子点赞数
		_, err = tx.Exec("UPDATE posts SET likes = likes + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "点赞失败: " + err.Error(),
			})
			return
		}

		message = "点赞成功"
		isLiked = true
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "操作失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的点赞数
	var likes int
	database.DB.QueryRow("SELECT likes FROM posts WHERE id = ?", id).Scan(&likes)

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: message,
		Data: gin.H{
			"likes":    likes,
			"is_liked": isLiked,
		},
	})
}

// UnlikePost 取消点赞帖子（兼容性API，建议使用LikePost）
func UnlikePost(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 检查是否已点赞
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM post_likes WHERE user_id = ? AND post_id = ?",
		userID, id,
	).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询点赞状态失败: " + err.Error(),
		})
		return
	}

	if count == 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "未点赞该帖子",
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "取消点赞失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 删除点赞记录
	_, err = tx.Exec(
		"DELETE FROM post_likes WHERE user_id = ? AND post_id = ?",
		userID, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "取消点赞失败: " + err.Error(),
		})
		return
	}

	// 更新帖子点赞数
	_, err = tx.Exec("UPDATE posts SET likes = likes - 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND likes > 0", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "取消点赞失败: " + err.Error(),
		})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "取消点赞失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的点赞数
	var likes int
	database.DB.QueryRow("SELECT likes FROM posts WHERE id = ?", id).Scan(&likes)

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "取消点赞成功",
		Data: gin.H{
			"likes":    likes,
			"is_liked": false,
		},
	})
}

// CoinPost 投币帖子
func CoinPost(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 可以从请求体中获取投币数量
	var req struct {
		Amount int `json:"amount" binding:"required,min=1,max=10"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Amount = 1 // 默认投1个币
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "投币失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 检查帖子是否存在，并获取帖子作者ID
	var postUserID int64
	err = tx.QueryRow("SELECT user_id FROM posts WHERE id = ?", id).Scan(&postUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "帖子不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询帖子失败: " + err.Error(),
		})
		return
	}

	// 检查投币者硬币是否足够
	var userCoins int
	err = tx.QueryRow("SELECT coins FROM users WHERE id = ?", userID).Scan(&userCoins)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询硬币失败: " + err.Error(),
		})
		return
	}

	if userCoins < req.Amount {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "硬币不足",
		})
		return
	}

	// 如果不是给自己投币，则进行硬币转移
	if postUserID != userID.(int64) {
		// 扣除投币者硬币
		_, err = tx.Exec("UPDATE users SET coins = coins - ? WHERE id = ?", req.Amount, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "扣除硬币失败: " + err.Error(),
			})
			return
		}

		// 给帖子作者增加硬币
		_, err = tx.Exec("UPDATE users SET coins = coins + ? WHERE id = ?", req.Amount, postUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "增加作者硬币失败: " + err.Error(),
			})
			return
		}
	}

	// 更新帖子投币数
	_, err = tx.Exec("UPDATE posts SET coins = coins + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", req.Amount, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "投币失败: " + err.Error(),
		})
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "投币失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的投币数
	var coins int
	database.DB.QueryRow("SELECT coins FROM posts WHERE id = ?", id).Scan(&coins)

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "投币成功",
		Data: gin.H{
			"coins":      coins,
			"user_coins": userCoins - req.Amount,
		},
	})
}

// GetPostStats 获取帖子统计信息
func GetPostStats(c *gin.Context) {
	id := c.Param("id")

	var stats struct {
		Likes        int `json:"likes"`
		Favorites    int `json:"favorites"`
		Coins        int `json:"coins"`
		CommentCount int `json:"comment_count"`
		ViewCount    int `json:"view_count"`
	}

	err := database.DB.QueryRow(
		"SELECT likes, favorites, coins, comment_count, view_count FROM posts WHERE id = ?",
		id,
	).Scan(&stats.Likes, &stats.Favorites, &stats.Coins, &stats.CommentCount, &stats.ViewCount)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "帖子不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询统计信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取统计信息成功",
		Data:    stats,
	})
}

// GetMyPosts 获取我发布的帖子
func GetMyPosts(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	// 获取查询参数
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	
	boardID := c.Query("board_id") // 可选的板块筛选
	sort := c.DefaultQuery("sort", "time") // 排序方式：time(时间), likes(点赞), comments(评论)
	
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE p.user_id = ?"
	args := []interface{}{userID}
	
	if boardID != "" {
		whereClause += " AND p.board_id = ?"
		args = append(args, boardID)
	}

	// 排序条件
	var orderClause string
	switch sort {
	case "likes":
		orderClause = "ORDER BY p.likes DESC, p.publish_time DESC"
	case "comments":
		orderClause = "ORDER BY p.comment_count DESC, p.publish_time DESC"
	case "time":
		fallthrough
	default:
		orderClause = "ORDER BY p.publish_time DESC"
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM posts p %s", whereClause)
	var total int
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询帖子总数失败: " + err.Error(),
		})
		return
	}

	// 查询帖子列表
	query := fmt.Sprintf(`
		SELECT p.id, p.board_id, p.title, p.content, p.type, p.publisher, p.publish_time, 
			p.coins, p.favorites, p.likes, p.image_url, p.comment_count, p.view_count,
			b.name as board_name
		FROM posts p 
		LEFT JOIN boards b ON p.board_id = b.id 
		%s %s 
		LIMIT ? OFFSET ?`, whereClause, orderClause)
	
	queryArgs := append(args, pageSize, offset)
	rows, err := database.DB.Query(query, queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询帖子列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var posts []gin.H
	for rows.Next() {
		var post struct {
			ID           int64
			BoardID      int64
			Title        string
			Content      string
			Type         string
			Publisher    string
			PublishTime  time.Time
			Coins        int
			Favorites    int
			Likes        int
			ImageURL     sql.NullString
			CommentCount int
			ViewCount    int
			BoardName    sql.NullString
		}

		err := rows.Scan(
			&post.ID, &post.BoardID, &post.Title, &post.Content, &post.Type,
			&post.Publisher, &post.PublishTime, &post.Coins, &post.Favorites,
			&post.Likes, &post.ImageURL, &post.CommentCount, &post.ViewCount,
			&post.BoardName,
		)
		if err != nil {
			continue
		}

		postData := gin.H{
			"id":            post.ID,
			"board_id":      post.BoardID,
			"board_name":    post.BoardName.String,
			"title":         post.Title,
			"content":       post.Content,
			"type":          post.Type,
			"publisher":     post.Publisher,
			"publish_time":  post.PublishTime.Format("2006-01-02 15:04:05"),
			"publish_time_ts": post.PublishTime.Unix(),
			"coins":         post.Coins,
			"favorites":     post.Favorites,
			"likes":         post.Likes,
			"comment_count": post.CommentCount,
			"view_count":    post.ViewCount,
		}

		if post.ImageURL.Valid {
			postData["image_url"] = post.ImageURL.String
		}

		posts = append(posts, postData)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取我的帖子成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     posts,
		},
	})
}
