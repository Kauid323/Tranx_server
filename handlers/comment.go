package handlers

import (
	"TaruApp/database"
	"TaruApp/models"
	"database/sql"
	"net/http"
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

	// 判断是否为楼主
	isAuthor := userID.(int64) == postUserID

	// 获取当前楼层号
	var floor int
	database.DB.QueryRow("SELECT COALESCE(MAX(floor), 0) + 1 FROM comments WHERE post_id = ?", req.PostID).Scan(&floor)

	// 插入评论
	now := time.Now()
	result, err := database.DB.Exec(
		`INSERT INTO comments (post_id, user_id, content, publisher, publish_time, is_author, floor) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.PostID, userID, req.Content, username, now, isAuthor, floor,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建评论失败: " + err.Error(),
		})
		return
	}

	// 更新帖子的评论数和最后回复时间
	database.DB.Exec(
		"UPDATE posts SET comment_count = comment_count + 1, last_reply_time = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		now, req.PostID,
	)

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "创建评论成功",
		Data: gin.H{
			"id":    id,
			"floor": floor,
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

	// 构建查询语句
	baseQuery := "SELECT id, post_id, user_id, content, publisher, publish_time, likes, is_author, floor, created_at, updated_at FROM comments WHERE post_id = ?"
	countQuery := "SELECT COUNT(*) FROM comments WHERE post_id = ?"
	orderClause := ""

	// 排序逻辑
	switch query.Sort {
	case "default":
		orderClause = " ORDER BY floor ASC" // 默认：按楼层正序
	case "likes":
		orderClause = " ORDER BY likes DESC, floor ASC" // 点赞最高
	case "author":
		orderClause = " ORDER BY is_author DESC, floor ASC" // 楼主发布（楼主的评论在前）
	case "desc":
		orderClause = " ORDER BY floor DESC" // 倒序：按楼层倒序
	default:
		orderClause = " ORDER BY floor ASC" // 默认正序
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
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.Publisher,
			&comment.PublishTime, &comment.Likes, &comment.IsAuthor, &comment.Floor,
			&comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			continue
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

// UpdateComment 更新评论
func UpdateComment(c *gin.Context) {
	id := c.Param("id")
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

	_, err := database.DB.Exec(
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

	// 获取评论所属的帖子ID
	var postID int64
	err := database.DB.QueryRow("SELECT post_id FROM comments WHERE id = ?", id).Scan(&postID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "评论不存在",
		})
		return
	}

	// 删除评论
	_, err = database.DB.Exec("DELETE FROM comments WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除评论失败: " + err.Error(),
		})
		return
	}

	// 更新帖子的评论数
	database.DB.Exec("UPDATE posts SET comment_count = comment_count - 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", postID)

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "删除评论成功",
	})
}

// LikeComment 点赞评论
func LikeComment(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 检查是否已点赞
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM comment_likes WHERE user_id = ? AND comment_id = ?",
		userID, id,
	).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "已经点赞过该评论",
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "点赞失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 记录点赞
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

	// 更新评论点赞数
	_, err = tx.Exec("UPDATE comments SET likes = likes + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "点赞失败: " + err.Error(),
		})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "点赞失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的点赞数
	var likes int
	database.DB.QueryRow("SELECT likes FROM comments WHERE id = ?", id).Scan(&likes)

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "点赞成功",
		Data: gin.H{
			"likes": likes,
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

	// 检查用户硬币是否足够
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

	// 扣除用户硬币
	_, err = tx.Exec("UPDATE users SET coins = coins - ? WHERE id = ?", req.Amount, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "扣除硬币失败: " + err.Error(),
		})
		return
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
