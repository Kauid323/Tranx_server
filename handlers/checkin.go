package handlers

import (
	"TaruApp/database"
	"TaruApp/models"
	"TaruApp/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CheckIn 每日签到
func CheckIn(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// 获取今天的日期
	today := time.Now().Format("2006-01-02")

	// 检查今天是否已签到
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM check_ins WHERE user_id = ? AND check_date = ?",
		userID, today,
	).Scan(&count)

	if err == nil && count > 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "今天已经签到过了",
		})
		return
	}

	// 签到奖励
	rewardCoins := 50
	rewardExp := 25
	now := time.Now()

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "签到失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 记录签到
	_, err = tx.Exec(
		"INSERT INTO check_ins (user_id, check_date, check_time, reward) VALUES (?, ?, ?, ?)",
		userID, today, now, rewardCoins,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "签到失败: " + err.Error(),
		})
		return
	}

	// 增加用户硬币和经验值
	_, err = tx.Exec(
		"UPDATE users SET coins = coins + ?, exp = exp + ? WHERE id = ?",
		rewardCoins, rewardExp, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "增加硬币和经验失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的用户信息
	var user models.User
	err = tx.QueryRow(
		"SELECT coins, exp FROM users WHERE id = ?",
		userID,
	).Scan(&user.Coins, &user.Exp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询用户信息失败: " + err.Error(),
		})
		return
	}

	// 计算新的用户等级
	newUserLevel := utils.CalculateUserLevel(user.Exp)

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
			Message: "签到失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "签到成功",
		Data: gin.H{
			"reward_coins": rewardCoins,
			"reward_exp":   rewardExp,
			"total_coins":  user.Coins,
			"total_exp":    user.Exp,
			"user_level":   newUserLevel,
			"check_time":   now,
		},
	})
}

// GetCheckInStatus 获取签到状态
func GetCheckInStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// 获取今天的日期
	today := time.Now().Format("2006-01-02")

	// 检查今天是否已签到
	var checkIn models.CheckIn
	err := database.DB.QueryRow(
		"SELECT id, user_id, check_date, check_time, reward FROM check_ins WHERE user_id = ? AND check_date = ?",
		userID, today,
	).Scan(&checkIn.ID, &checkIn.UserID, &checkIn.CheckDate, &checkIn.CheckTime, &checkIn.Reward)

	if err != nil {
		// 今天未签到
		c.JSON(http.StatusOK, models.Response{
			Code:    200,
			Message: "今天未签到",
			Data: gin.H{
				"checked_in": false,
				"can_check":  true,
			},
		})
		return
	}

	// 今天已签到
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "今天已签到",
		Data: gin.H{
			"checked_in": true,
			"can_check":  false,
			"check_time": checkIn.CheckTime,
			"reward":     checkIn.Reward,
		},
	})
}

// GetCheckInRank 获取今日签到排行榜
func GetCheckInRank(c *gin.Context) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "100")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	today := time.Now().Format("2006-01-02")

	// 查询今日签到总数
	var total int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM check_ins WHERE check_date = ?",
		today,
	).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}

	// 查询今日签到排行榜（按签到时间排序，最早的排第一）
	rows, err := database.DB.Query(`
		SELECT c.user_id, u.username, u.avatar, c.check_time
		FROM check_ins c
		JOIN users u ON c.user_id = u.id
		WHERE c.check_date = ?
		ORDER BY c.check_time ASC
		LIMIT ? OFFSET ?
	`, today, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询排行榜失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var rankList []models.CheckInRankItem
	rank := offset + 1
	for rows.Next() {
		var item models.CheckInRankItem
		if err := rows.Scan(&item.UserID, &item.Username, &item.Avatar, &item.CheckTime); err != nil {
			continue
		}
		item.Rank = rank
		rankList = append(rankList, item)
		rank++
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取签到排行榜成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     rankList,
		},
	})
}

// GetCheckInHistory 获取用户签到历史
func GetCheckInHistory(c *gin.Context) {
	userID := c.Param("id")

	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "30")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	// 查询签到总数
	var total int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM check_ins WHERE user_id = ?",
		userID,
	).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}

	// 查询签到历史
	rows, err := database.DB.Query(`
		SELECT id, user_id, check_date, check_time, reward, created_at
		FROM check_ins
		WHERE user_id = ?
		ORDER BY check_date DESC
		LIMIT ? OFFSET ?
	`, userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询签到历史失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var history []models.CheckIn
	for rows.Next() {
		var checkIn models.CheckIn
		if err := rows.Scan(&checkIn.ID, &checkIn.UserID, &checkIn.CheckDate, &checkIn.CheckTime, &checkIn.Reward, &checkIn.CreatedAt); err != nil {
			continue
		}
		history = append(history, checkIn)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取签到历史成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     history,
		},
	})
}
