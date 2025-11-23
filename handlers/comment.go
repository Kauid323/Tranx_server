package handlers

import (
	"TaruApp/database"
	"TaruApp/models"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateComment 创建评论
func CreateComment(c *gin.Context) {
	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户信息
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	// 获取用户头像
	var avatar string
	database.DB.QueryRow("SELECT COALESCE(avatar, '') FROM users WHERE id = ?", userID).Scan(&avatar)

	// 检查帖子是否存在，并获取帖子作者ID
	var postUserID int64
	err := database.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", req.PostID).Scan(&postUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "帖子不存在",
		})
		return
	}

	// 如果是楼中楼回复，检查父评论是否存在
	if req.ParentID != nil {
		var parentExists int
		err = database.DB.QueryRow("SELECT COUNT(*) FROM comments WHERE id = ? AND post_id = ?", *req.ParentID, req.PostID).Scan(&parentExists)
		if err != nil || parentExists == 0 {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "父评论不存在",
			})
			return
		}
	}

	// 判断是否为楼主
	isAuthor := userID.(int64) == postUserID

	// 获取当前楼层号（只有顶级评论才有楼层号，楼中楼回复楼层号为0）
	var floor int
	if req.ParentID == nil {
		database.DB.QueryRow("SELECT COALESCE(MAX(floor), 0) + 1 FROM comments WHERE post_id = ? AND parent_id IS NULL", req.PostID).Scan(&floor)
	} else {
		floor = 0 // 楼中楼回复没有楼层号
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建评论失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 插入评论
	now := time.Now()
	result, err := tx.Exec(
		`INSERT INTO comments (post_id, user_id, parent_id, content, publisher, publish_time, is_author, floor) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.PostID, userID, req.ParentID, req.Content, username, now, isAuthor, floor,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建评论失败: " + err.Error(),
		})
		return
	}

	// 如果是楼中楼回复，更新父评论的回复数
	if req.ParentID != nil {
		_, err = tx.Exec("UPDATE comments SET reply_count = reply_count + 1 WHERE id = ?", *req.ParentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "更新父评论回复数失败: " + err.Error(),
			})
			return
		}
	}

	// 更新帖子评论数
	_, err = tx.Exec("UPDATE posts SET comment_count = comment_count + 1, last_reply_time = ? WHERE id = ?", now, req.PostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新帖子评论数失败: " + err.Error(),
		})
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建评论失败: " + err.Error(),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建评论失败: " + err.Error(),
		})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "创建评论成功",
		Data: gin.H{
			"id":        id,
			"floor":     floor,
			"parent_id": req.ParentID,
		},
	})
}

// GetComments 获取评论列表（支持多种排序）
func GetComments(c *gin.Context) {
	var query models.GetCommentsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 默认分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 50
	}
	if query.PageSize > 200 {
		query.PageSize = 200
	}

	// 构建查询语句 - 联合用户表获取头像
	baseQuery := `SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.publisher, 
		c.publish_time, c.likes, c.coins, c.is_author, c.floor, c.reply_count, 
		c.created_at, c.updated_at, COALESCE(u.avatar, '') as avatar 
		FROM comments c 
		LEFT JOIN users u ON c.user_id = u.id 
		WHERE c.post_id = ? AND c.parent_id IS NULL`
	countQuery := "SELECT COUNT(*) FROM comments WHERE post_id = ? AND parent_id IS NULL"
	orderClause := ""

	// 排序逻辑（只对顶级评论排序）
	switch query.Sort {
	case "default":
		orderClause = " ORDER BY c.floor ASC" // 默认：按楼层正序
	case "likes":
		orderClause = " ORDER BY c.likes DESC, c.floor ASC" // 点赞最高
	case "author":
		orderClause = " ORDER BY c.is_author DESC, c.floor ASC" // 楼主发布（楼主的评论在前）
	case "desc":
		orderClause = " ORDER BY c.floor DESC" // 倒序：按楼层倒序
	default:
		orderClause = " ORDER BY c.floor ASC" // 默认正序
	}

	// 分页
	offset := (query.Page - 1) * query.PageSize
	limitClause := " LIMIT ? OFFSET ?"

	// 查询总数
	var total int
	err := database.DB.QueryRow(countQuery, query.PostID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询评论总数失败: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID（如果已登录）
	var currentUserID int64
	if userID, exists := c.Get("user_id"); exists {
		currentUserID = userID.(int64)
	}

	// 查询评论列表
	rows, err := database.DB.Query(
		baseQuery+orderClause+limitClause,
		query.PostID, query.PageSize, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询评论列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		var parentID sql.NullInt64
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &parentID, &comment.Content, &comment.Publisher,
			&comment.PublishTime, &comment.Likes, &comment.Coins, &comment.IsAuthor, &comment.Floor, &comment.ReplyCount,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.Avatar,
		)
		if err != nil {
			continue
		}

		if parentID.Valid {
			comment.ParentID = &parentID.Int64
		}

		// 判断是否是当前用户的评论
		if currentUserID > 0 {
			comment.IsMyComment = (comment.UserID == currentUserID)
			
			// 查询当前用户是否点赞了该评论
			var likeCount int
			database.DB.QueryRow(
				"SELECT COUNT(*) FROM comment_likes WHERE user_id = ? AND comment_id = ?",
				currentUserID, comment.ID,
			).Scan(&likeCount)
			comment.IsLiked = likeCount > 0
		}

		comments = append(comments, comment)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取评论列表成功",
		Data: models.PageData{
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
			List:     comments,
		},
	})
}

// GetCommentReplies 获取评论的子回复列表
func GetCommentReplies(c *gin.Context) {
	commentID := c.Param("id")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page := 1
	pageSize := 20
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	offset := (page - 1) * pageSize

	// 查询子回复总数
	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM comments WHERE parent_id = ?", commentID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询子回复总数失败: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID（如果已登录）
	var currentUserID int64
	if userID, exists := c.Get("user_id"); exists {
		currentUserID = userID.(int64)
	}

	// 查询子回复列表
	rows, err := database.DB.Query(`
		SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.publisher, 
			c.publish_time, c.likes, c.coins, c.is_author, c.floor, c.reply_count,
			c.created_at, c.updated_at, COALESCE(u.avatar, '') as avatar 
		FROM comments c 
		LEFT JOIN users u ON c.user_id = u.id 
		WHERE c.parent_id = ? 
		ORDER BY c.publish_time ASC 
		LIMIT ? OFFSET ?
	`, commentID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询子回复列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var replies []models.Comment
	for rows.Next() {
		var reply models.Comment
		var parentID sql.NullInt64
		err := rows.Scan(
			&reply.ID, &reply.PostID, &reply.UserID, &parentID, &reply.Content, &reply.Publisher,
			&reply.PublishTime, &reply.Likes, &reply.Coins, &reply.IsAuthor, &reply.Floor, &reply.ReplyCount,
			&reply.CreatedAt, &reply.UpdatedAt, &reply.Avatar,
		)
		if err != nil {
			continue
		}

		if parentID.Valid {
			reply.ParentID = &parentID.Int64
		}

		// 判断是否是当前用户的评论
		if currentUserID > 0 {
			reply.IsMyComment = (reply.UserID == currentUserID)
			
			// 查询当前用户是否点赞了该评论
			var likeCount int
			database.DB.QueryRow(
				"SELECT COUNT(*) FROM comment_likes WHERE user_id = ? AND comment_id = ?",
				currentUserID, reply.ID,
			).Scan(&likeCount)
			reply.IsLiked = likeCount > 0
		}

		replies = append(replies, reply)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取子回复列表成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     replies,
		},
	})
}

// UpdateComment 更新评论
func UpdateComment(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查评论是否存在，并验证是否为作者本人
	var commentUserID int64
	err := database.DB.QueryRow("SELECT user_id FROM comments WHERE id = ?", id).Scan(&commentUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "评论不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询评论失败: " + err.Error(),
		})
		return
	}

	// 验证权限：只有作者本人才能编辑
	if commentUserID != userID.(int64) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权编辑此评论，只能编辑自己的评论",
		})
		return
	}

	_, err = database.DB.Exec(
		"UPDATE comments SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		req.Content, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新评论失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "更新评论成功",
	})
}

// DeleteComment 删除评论
func DeleteComment(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 获取评论信息，验证是否为作者本人
	var commentUserID, postID int64
	var parentID sql.NullInt64
	err := database.DB.QueryRow(
		"SELECT user_id, post_id, parent_id FROM comments WHERE id = ?",
		id,
	).Scan(&commentUserID, &postID, &parentID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "评论不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询评论失败: " + err.Error(),
		})
		return
	}

	// 验证权限：只有作者本人才能删除
	if commentUserID != userID.(int64) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权删除此评论，只能删除自己的评论",
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除评论失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 1. 删除该评论的所有子回复的点赞记录
	_, err = tx.Exec("DELETE FROM comment_likes WHERE comment_id IN (SELECT id FROM comments WHERE parent_id = ?)", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除子回复点赞记录失败: " + err.Error(),
		})
		return
	}

	// 2. 删除该评论的所有子回复
	subCommentsResult, err := tx.Exec("DELETE FROM comments WHERE parent_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除子回复失败: " + err.Error(),
		})
		return
	}
	subCommentsCount, _ := subCommentsResult.RowsAffected()

	// 3. 删除该评论的点赞记录
	_, err = tx.Exec("DELETE FROM comment_likes WHERE comment_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除评论点赞记录失败: " + err.Error(),
		})
		return
	}

	// 4. 删除评论本身
	_, err = tx.Exec("DELETE FROM comments WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除评论失败: " + err.Error(),
		})
		return
	}

	// 5. 如果是楼中楼回复，更新父评论的回复数
	if parentID.Valid {
		_, err = tx.Exec("UPDATE comments SET reply_count = reply_count - 1 WHERE id = ? AND reply_count > 0", parentID.Int64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "更新父评论回复数失败: " + err.Error(),
			})
			return
		}
	}

	// 6. 更新帖子的评论数（包括删除的子回复数量）
	totalDeletedComments := 1 + subCommentsCount // 主评论 + 子回复数量
	_, err = tx.Exec(
		"UPDATE posts SET comment_count = comment_count - ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND comment_count >= ?",
		totalDeletedComments, postID, totalDeletedComments,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新帖子评论数失败: " + err.Error(),
		})
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除评论失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "删除评论成功",
		Data: gin.H{
			"deleted_replies": subCommentsCount,
		},
	})
}

// LikeComment 点赞/取消点赞评论（切换功能）
func LikeComment(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 检查是否已点赞
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM comment_likes WHERE user_id = ? AND comment_id = ?",
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
			"DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?",
			userID, id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "取消点赞失败: " + err.Error(),
			})
			return
		}

		// 减少评论点赞数
		_, err = tx.Exec("UPDATE comments SET likes = likes - 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND likes > 0", id)
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
			"INSERT INTO comment_likes (user_id, comment_id) VALUES (?, ?)",
			userID, id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "点赞失败: " + err.Error(),
			})
			return
		}

		// 增加评论点赞数
		_, err = tx.Exec("UPDATE comments SET likes = likes + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
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
	database.DB.QueryRow("SELECT likes FROM comments WHERE id = ?", id).Scan(&likes)

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: message,
		Data: gin.H{
			"likes":    likes,
			"is_liked": isLiked,
		},
	})
}

// CoinComment 投币评论
func CoinComment(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 获取投币数量
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

	// 检查评论是否存在，并获取评论作者ID
	var commentUserID int64
	err = tx.QueryRow("SELECT user_id FROM comments WHERE id = ?", id).Scan(&commentUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "评论不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询评论失败: " + err.Error(),
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
	if commentUserID != userID.(int64) {
		// 扣除投币者硬币
		_, err = tx.Exec("UPDATE users SET coins = coins - ? WHERE id = ?", req.Amount, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "扣除硬币失败: " + err.Error(),
			})
			return
		}

		// 给评论作者增加硬币
		_, err = tx.Exec("UPDATE users SET coins = coins + ? WHERE id = ?", req.Amount, commentUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "增加作者硬币失败: " + err.Error(),
			})
			return
		}
	}

	// 更新评论投币数
	_, err = tx.Exec("UPDATE comments SET coins = coins + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", req.Amount, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "投币失败: " + err.Error(),
		})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "投币失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的投币数
	var coins int
	database.DB.QueryRow("SELECT coins FROM comments WHERE id = ?", id).Scan(&coins)

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "投币成功",
		Data: gin.H{
			"coins":      coins,
			"user_coins": userCoins - req.Amount,
		},
	})
}
