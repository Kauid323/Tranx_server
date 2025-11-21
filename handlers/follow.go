package handlers

import (
	"TaruApp/database"
	"TaruApp/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FollowUser 关注用户
func FollowUser(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID, _ := c.Get("user_id")

	// 检查目标用户是否存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", targetUserID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "用户不存在",
		})
		return
	}

	// 检查是否已关注
	err = database.DB.QueryRow(
		"SELECT COUNT(*) FROM follows WHERE user_id = ? AND followed_id = ?",
		currentUserID, targetUserID,
	).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "已经关注该用户",
		})
		return
	}

	// 创建关注关系
	_, err = database.DB.Exec(
		"INSERT INTO follows (user_id, followed_id) VALUES (?, ?)",
		currentUserID, targetUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "关注失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "关注成功",
	})
}

// UnfollowUser 取消关注用户
func UnfollowUser(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID, _ := c.Get("user_id")

	// 删除关注关系
	result, err := database.DB.Exec(
		"DELETE FROM follows WHERE user_id = ? AND followed_id = ?",
		currentUserID, targetUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "取消关注失败: " + err.Error(),
		})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "未关注该用户",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "取消关注成功",
	})
}

// GetFollowingList 获取关注列表
func GetFollowingList(c *gin.Context) {
	userID := c.Param("id")
	page := 1
	pageSize := 20

	if p, ok := c.GetQuery("page"); ok {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if ps, ok := c.GetQuery("page_size"); ok {
		if psInt, err := strconv.Atoi(ps); err == nil && psInt > 0 && psInt <= 100 {
			pageSize = psInt
		}
	}

	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM follows WHERE user_id = ?",
		userID,
	).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}

	// 查询关注列表
	rows, err := database.DB.Query(`
		SELECT u.id, u.username, u.avatar, u.level, u.coins, u.exp, u.user_level, u.created_at
		FROM follows f
		JOIN users u ON f.followed_id = u.id
		WHERE f.user_id = ?
		ORDER BY f.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询关注列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Avatar, &user.Level, &user.Coins, &user.Exp, &user.UserLevel, &user.CreatedAt); err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取关注列表成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     users,
		},
	})
}

// GetFollowerList 获取粉丝列表
func GetFollowerList(c *gin.Context) {
	userID := c.Param("id")
	page := 1
	pageSize := 20

	if p, ok := c.GetQuery("page"); ok {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if ps, ok := c.GetQuery("page_size"); ok {
		if psInt, err := strconv.Atoi(ps); err == nil && psInt > 0 && psInt <= 100 {
			pageSize = psInt
		}
	}

	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM follows WHERE followed_id = ?",
		userID,
	).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}

	// 查询粉丝列表
	rows, err := database.DB.Query(`
		SELECT u.id, u.username, u.avatar, u.level, u.coins, u.exp, u.user_level, u.created_at
		FROM follows f
		JOIN users u ON f.user_id = u.id
		WHERE f.followed_id = ?
		ORDER BY f.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询粉丝列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Avatar, &user.Level, &user.Coins, &user.Exp, &user.UserLevel, &user.CreatedAt); err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取粉丝列表成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     users,
		},
	})
}

// GetUserStats 获取用户统计信息（关注数、粉丝数等）
func GetUserStats(c *gin.Context) {
	userID := c.Param("id")
	currentUserID, _ := c.Get("user_id")

	var stats models.UserStats

	// 获取关注数
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM follows WHERE user_id = ?",
		userID,
	).Scan(&stats.FollowingCount)

	// 获取粉丝数
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM follows WHERE followed_id = ?",
		userID,
	).Scan(&stats.FollowerCount)

	// 检查当前用户是否关注了该用户
	if currentUserID != nil {
		var count int
		err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM follows WHERE user_id = ? AND followed_id = ?",
			currentUserID, userID,
		).Scan(&count)
		if err == nil && count > 0 {
			stats.IsFollowing = true
		}
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取用户统计成功",
		Data:    stats,
	})
}
