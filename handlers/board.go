package handlers

import (
	"TaruApp/database"
	"TaruApp/models"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateBoard 创建板块
func CreateBoard(c *gin.Context) {
	var req models.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取创建者信息
	creatorID, _ := c.Get("user_id")
	creatorName, _ := c.Get("username")

	// 获取创建者头像
	var creatorAvatar string
	database.DB.QueryRow("SELECT COALESCE(avatar, '') FROM users WHERE id = ?", creatorID).Scan(&creatorAvatar)

	result, err := database.DB.Exec(
		"INSERT INTO boards (name, description, avatar_url, creator_id, creator_name, creator_avatar) VALUES (?, ?, ?, ?, ?, ?)",
		req.Name, req.Description, req.AvatarURL, creatorID, creatorName, creatorAvatar,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建板块失败: " + err.Error(),
		})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "创建板块成功",
		Data: gin.H{
			"id": id,
		},
	})
}

// GetAllBoards 获取所有板块
func GetAllBoards(c *gin.Context) {
	rows, err := database.DB.Query(
		`SELECT id, name, description, avatar_url, creator_id, creator_name, creator_avatar, 
		created_at, updated_at FROM boards ORDER BY created_at DESC`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询板块失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var boards []models.Board
	for rows.Next() {
		var board models.Board
		var creatorID sql.NullInt64
		var creatorName, creatorAvatar sql.NullString

		if err := rows.Scan(&board.ID, &board.Name, &board.Description, &board.AvatarURL,
			&creatorID, &creatorName, &creatorAvatar, &board.CreatedAt, &board.UpdatedAt); err != nil {
			continue
		}

		// 处理可能为空的创建者信息
		if creatorID.Valid {
			board.CreatorID = creatorID.Int64
		}
		if creatorName.Valid {
			board.CreatorName = creatorName.String
		}
		if creatorAvatar.Valid {
			board.CreatorAvatar = creatorAvatar.String
		}

		// 设置时间戳
		board.CreatedAtTs = board.CreatedAt.Unix()
		board.UpdatedAtTs = board.UpdatedAt.Unix()

		boards = append(boards, board)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取板块列表成功",
		Data:    boards,
	})
}

// GetBoardDetail 获取板块详情
func GetBoardDetail(c *gin.Context) {
	id := c.Param("id")

	var board models.Board
	var creatorID sql.NullInt64
	var creatorName, creatorAvatar sql.NullString

	err := database.DB.QueryRow(
		`SELECT id, name, description, avatar_url, creator_id, creator_name, creator_avatar, 
		created_at, updated_at FROM boards WHERE id = ?`,
		id,
	).Scan(&board.ID, &board.Name, &board.Description, &board.AvatarURL,
		&creatorID, &creatorName, &creatorAvatar, &board.CreatedAt, &board.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "板块不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询板块失败: " + err.Error(),
		})
		return
	}

	// 处理可能为空的创建者信息
	if creatorID.Valid {
		board.CreatorID = creatorID.Int64
	}
	if creatorName.Valid {
		board.CreatorName = creatorName.String
	}
	if creatorAvatar.Valid {
		board.CreatorAvatar = creatorAvatar.String
	}

	// 设置时间戳
	board.CreatedAtTs = board.CreatedAt.Unix()
	board.UpdatedAtTs = board.UpdatedAt.Unix()

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取板块详情成功",
		Data:    board,
	})
}

// UpdateBoard 更新板块
func UpdateBoard(c *gin.Context) {
	id := c.Param("id")
	var req models.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	_, err := database.DB.Exec(
		"UPDATE boards SET name = ?, description = ?, avatar_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		req.Name, req.Description, req.AvatarURL, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新板块失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "更新板块成功",
	})
}

// DeleteBoard 删除板块
func DeleteBoard(c *gin.Context) {
	id := c.Param("id")

	_, err := database.DB.Exec("DELETE FROM boards WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除板块失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "删除板块成功",
	})
}

// GetBoardStats 获取板块统计信息
func GetBoardStats(c *gin.Context) {
	id := c.Param("id")

	var postCount, totalViews, totalComments int
	err := database.DB.QueryRow(
		`SELECT 
			COUNT(*) as post_count,
			COALESCE(SUM(view_count), 0) as total_views,
			COALESCE(SUM(comment_count), 0) as total_comments
		FROM posts WHERE board_id = ?`,
		id,
	).Scan(&postCount, &totalViews, &totalComments)

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
		Data: gin.H{
			"post_count":     postCount,
			"total_views":    totalViews,
			"total_comments": totalComments,
		},
	})
}
