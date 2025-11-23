package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"TaruApp/database"
	"TaruApp/models"

	"github.com/gin-gonic/gin"
)

// 定义应用渠道选项
var appChannels = []models.AppChannel{
	{Value: "official", Label: "官方版"},
	{Value: "international", Label: "国际版"},
	{Value: "test", Label: "测试版"},
	{Value: "custom", Label: "定制版"},
}

// 定义广告级别选项
var appAdLevels = []models.AppAdLevel{
	{Value: "none", Label: "无广告"},
	{Value: "few", Label: "少量广告"},
	{Value: "many", Label: "超多广告"},
	{Value: "adware", Label: "广告软件"},
}

// 定义付费类型选项
var appPaymentTypes = []models.AppPaymentType{
	{Value: "free", Label: "免费"},
	{Value: "iap", Label: "内购"},
	{Value: "few_iap", Label: "少量内购"},
	{Value: "paid", Label: "不给钱不让用"},
}

// 定义运营方式选项
var appOperationTypes = []models.AppOperationType{
	{Value: "team", Label: "团队开发"},
	{Value: "indie", Label: "独立开发"},
	{Value: "opensource", Label: "开源软件"},
}

// GetAppChannels 获取应用渠道列表
func GetAppChannels(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取应用渠道成功",
		Data:    appChannels,
	})
}

// GetAppAdLevels 获取广告级别列表
func GetAppAdLevels(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取广告级别成功",
		Data:    appAdLevels,
	})
}

// GetAppPaymentTypes 获取付费类型列表
func GetAppPaymentTypes(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取付费类型成功",
		Data:    appPaymentTypes,
	})
}

// GetAppOperationTypes 获取运营方式列表
func GetAppOperationTypes(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取运营方式成功",
		Data:    appOperationTypes,
	})
}

// UploadApp 上传应用
func UploadApp(c *gin.Context) {
	var req models.UploadAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	// 验证分类是否存在
	if !validateCategory(req.MainCategory, req.SubCategory) {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "应用分类不存在",
		})
		return
	}

	// 将截图数组转为JSON字符串
	screenshotsJSON, _ := json.Marshal(req.Screenshots)

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "上传失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 插入上传任务
	result, err := tx.Exec(
		`INSERT INTO app_upload_tasks (
			user_id, package_name, name, icon_url, version, version_code, size,
			channel, main_category, sub_category, screenshots, description,
			share_desc, update_content, developer_name, ad_level, payment_type,
			operation_type, download_url, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, req.PackageName, req.Name, req.IconURL, req.Version,
		req.VersionCode, req.Size, req.Channel, req.MainCategory,
		req.SubCategory, string(screenshotsJSON), req.Description,
		req.ShareDesc, req.UpdateContent, req.DeveloperName,
		req.AdLevel, req.PaymentType, req.OperationType,
		req.DownloadURL, "pending",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "上传失败: " + err.Error(),
		})
		return
	}

	taskID, _ := result.LastInsertId()

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "上传失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "应用上传成功，等待审核",
		Data: gin.H{
			"task_id":     taskID,
			"status":      "pending",
			"uploader":    username.(string),
			"upload_time": time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// GetMyUploadTasks 获取我的上传任务（审核情况）
func GetMyUploadTasks(c *gin.Context) {
	userID, _ := c.Get("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM app_upload_tasks WHERE user_id = ?",
		userID,
	).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}

	// 查询列表
	rows, err := database.DB.Query(
		`SELECT id, package_name, name, icon_url, version, status, reject_reason, created_at
		FROM app_upload_tasks
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		userID, pageSize, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	tasks := []gin.H{}
	for rows.Next() {
		var task struct {
			ID           int64
			PackageName  string
			Name         string
			IconURL      string
			Version      string
			Status       string
			RejectReason sql.NullString
			CreatedAt    time.Time
		}
		if err := rows.Scan(&task.ID, &task.PackageName, &task.Name, &task.IconURL,
			&task.Version, &task.Status, &task.RejectReason, &task.CreatedAt); err != nil {
			continue
		}

		// 转换状态显示
		statusLabel := "待审核"
		if task.Status == "rejected" {
			statusLabel = "被拒绝"
		} else if task.Status == "approved" {
			statusLabel = "已通过"
		}

		taskData := gin.H{
			"task_id":      task.ID,
			"package_name": task.PackageName,
			"name":         task.Name,
			"icon_url":     task.IconURL,
			"version":      task.Version,
			"status":       task.Status,
			"status_label": statusLabel,
			"upload_time":  task.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if task.RejectReason.Valid && task.Status == "rejected" {
			taskData["reject_reason"] = task.RejectReason.String
		}

		tasks = append(tasks, taskData)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取上传任务成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     tasks,
		},
	})
}

// GetPendingApps 获取待审核应用列表（需要审核权限）
func GetPendingApps(c *gin.Context) {
	// 检查用户权限
	userLevel, _ := c.Get("level")
	if userLevel.(int) < 80 {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "权限不足，需要审核权限",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM app_upload_tasks WHERE status = 'pending'",
	).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}

	// 查询列表
	rows, err := database.DB.Query(
		`SELECT t.id, t.package_name, t.name, t.icon_url, t.version, t.created_at, 
			t.user_id, u.username
		FROM app_upload_tasks t
		LEFT JOIN users u ON t.user_id = u.id
		WHERE t.status = 'pending'
		ORDER BY t.created_at ASC
		LIMIT ? OFFSET ?`,
		pageSize, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	tasks := []gin.H{}
	for rows.Next() {
		var task struct {
			ID          int64
			PackageName string
			Name        string
			IconURL     string
			Version     string
			CreatedAt   time.Time
			UserID      int64
			Username    sql.NullString
		}
		if err := rows.Scan(&task.ID, &task.PackageName, &task.Name, &task.IconURL,
			&task.Version, &task.CreatedAt, &task.UserID, &task.Username); err != nil {
			continue
		}

		uploaderName := "未知用户"
		if task.Username.Valid {
			uploaderName = task.Username.String
		}

		tasks = append(tasks, gin.H{
			"task_id":      task.ID,
			"package_name": task.PackageName,
			"name":         task.Name,
			"icon_url":     task.IconURL,
			"version":      task.Version,
			"upload_time":  task.CreatedAt.Format("2006-01-02 15:04:05"),
			"uploader":     uploaderName,
			"uploader_id":  task.UserID,
		})
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取待审核应用成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     tasks,
		},
	})
}

// GetAppUploadDetail 获取上传任务详情（用于审核查看详细信息）
func GetAppUploadDetail(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "任务ID无效",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userLevel, _ := c.Get("level")

	// 查询任务详情
	var task models.AppUploadTask
	var uploaderName sql.NullString
	err = database.DB.QueryRow(
		`SELECT t.*, u.username
		FROM app_upload_tasks t
		LEFT JOIN users u ON t.user_id = u.id
		WHERE t.id = ?`,
		taskID,
	).Scan(
		&task.ID, &task.UserID, &task.PackageName, &task.Name, &task.IconURL,
		&task.Version, &task.VersionCode, &task.Size, &task.Channel,
		&task.MainCategory, &task.SubCategory, &task.Screenshots,
		&task.Description, &task.ShareDesc, &task.UpdateContent,
		&task.DeveloperName, &task.AdLevel, &task.PaymentType,
		&task.OperationType, &task.DownloadURL, &task.Status,
		&task.RejectReason, &task.ReviewerID, &task.ReviewTime,
		&task.CreatedAt, &task.UpdatedAt, &uploaderName,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "任务不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}

	// 权限检查：只能查看自己的任务，或者有审核权限的用户可以查看所有任务
	if task.UserID != userID.(int64) && userLevel.(int) < 80 {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权查看此任务",
		})
		return
	}

	if uploaderName.Valid {
		task.UploaderName = uploaderName.String
	}

	// 解析截图JSON
	var screenshots []string
	if task.Screenshots != "" {
		json.Unmarshal([]byte(task.Screenshots), &screenshots)
	}

	// 构建响应数据
	responseData := gin.H{
		"task_id":        task.ID,
		"package_name":   task.PackageName,
		"name":           task.Name,
		"icon_url":       task.IconURL,
		"version":        task.Version,
		"version_code":   task.VersionCode,
		"size":           task.Size,
		"channel":        task.Channel,
		"main_category":  task.MainCategory,
		"sub_category":   task.SubCategory,
		"screenshots":    screenshots,
		"description":    task.Description,
		"share_desc":     task.ShareDesc,
		"update_content": task.UpdateContent,
		"developer_name": task.DeveloperName,
		"ad_level":       task.AdLevel,
		"payment_type":   task.PaymentType,
		"operation_type": task.OperationType,
		"download_url":   task.DownloadURL,
		"status":         task.Status,
		"uploader_name":  task.UploaderName,
		"upload_time":    task.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if task.Status == "rejected" && task.RejectReason != "" {
		responseData["reject_reason"] = task.RejectReason
	}

	if task.ReviewTime != nil {
		responseData["review_time"] = task.ReviewTime.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取任务详情成功",
		Data:    responseData,
	})
}

// ReviewApp 审核应用（需要审核权限）
func ReviewApp(c *gin.Context) {
	// 检查用户权限
	userLevel, _ := c.Get("level")
	if userLevel.(int) < 80 {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "权限不足，需要审核权限",
		})
		return
	}

	var req models.ReviewAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 如果拒绝但没有提供原因
	if req.Accept == 0 && req.RejectReason == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "拒绝时必须提供拒绝原因",
		})
		return
	}

	reviewerID, _ := c.Get("user_id")
	reviewTime := time.Now()

	// 查询任务信息
	var task models.AppUploadTask
	err := database.DB.QueryRow(
		`SELECT id, user_id, package_name, name, icon_url, version, version_code, size,
			channel, main_category, sub_category, screenshots, description, share_desc,
			update_content, developer_name, ad_level, payment_type, operation_type,
			download_url, status
		FROM app_upload_tasks WHERE id = ?`,
		req.TaskID,
	).Scan(
		&task.ID, &task.UserID, &task.PackageName, &task.Name, &task.IconURL,
		&task.Version, &task.VersionCode, &task.Size, &task.Channel,
		&task.MainCategory, &task.SubCategory, &task.Screenshots,
		&task.Description, &task.ShareDesc, &task.UpdateContent,
		&task.DeveloperName, &task.AdLevel, &task.PaymentType,
		&task.OperationType, &task.DownloadURL, &task.Status,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "任务不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询任务失败: " + err.Error(),
		})
		return
	}

	// 检查任务状态
	if task.Status != "pending" {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "该任务已经审核过了",
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "审核失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	if req.Accept == 1 {
		// 通过审核，需要创建或更新应用信息
		// 首先检查应用是否存在
		var appID int64
		err = tx.QueryRow(
			"SELECT id FROM apps WHERE package_name = ?",
			task.PackageName,
		).Scan(&appID)

		if err == sql.ErrNoRows {
			// 应用不存在，创建新应用
			result, err := tx.Exec(
				`INSERT INTO apps (package_name, name, icon_url, description, tags,
					main_category, sub_category, channel, share_desc, developer_name,
					ad_level, payment_type, operation_type, rating, rating_count, 
					total_coins, download_count) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0)`,
				task.PackageName, task.Name, task.IconURL, task.Description,
				"", task.MainCategory, task.SubCategory, task.Channel,
				task.ShareDesc, task.DeveloperName, task.AdLevel,
				task.PaymentType, task.OperationType,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.Response{
					Code:    500,
					Message: "创建应用失败: " + err.Error(),
				})
				return
			}
			appID, _ = result.LastInsertId()
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "查询应用失败: " + err.Error(),
			})
			return
		} else {
			// 应用存在，更新应用信息
			_, err = tx.Exec(
				`UPDATE apps SET name = ?, icon_url = ?, description = ?,
					main_category = ?, sub_category = ?, channel = ?, share_desc = ?,
					developer_name = ?, ad_level = ?, payment_type = ?, operation_type = ?,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = ?`,
				task.Name, task.IconURL, task.Description,
				task.MainCategory, task.SubCategory, task.Channel, task.ShareDesc,
				task.DeveloperName, task.AdLevel, task.PaymentType, task.OperationType, appID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.Response{
					Code:    500,
					Message: "更新应用信息失败: " + err.Error(),
				})
				return
			}
		}

		// 将之前的最新版本标记为非最新
		_, err = tx.Exec(
			"UPDATE app_versions SET is_latest = 0 WHERE app_id = ?",
			appID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "更新版本状态失败: " + err.Error(),
			})
			return
		}

		// 获取上传者用户名
		var uploaderName string
		tx.QueryRow("SELECT username FROM users WHERE id = ?", task.UserID).Scan(&uploaderName)

		// 创建新版本
		_, err = tx.Exec(
			`INSERT INTO app_versions (app_id, package_name, version, version_code,
				size, download_url, update_content, screenshots, uploader_id,
				uploader_name, is_latest)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
			appID, task.PackageName, task.Version, task.VersionCode,
			task.Size, task.DownloadURL, task.UpdateContent,
			task.Screenshots, task.UserID, uploaderName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "创建版本失败: " + err.Error(),
			})
			return
		}

		// 更新任务状态为已通过
		_, err = tx.Exec(
			`UPDATE app_upload_tasks 
			SET status = 'approved', reviewer_id = ?, review_time = ?, updated_at = ?
			WHERE id = ?`,
			reviewerID, reviewTime, reviewTime, req.TaskID,
		)
	} else {
		// 拒绝审核
		_, err = tx.Exec(
			`UPDATE app_upload_tasks 
			SET status = 'rejected', reject_reason = ?, reviewer_id = ?, review_time = ?, updated_at = ?
			WHERE id = ?`,
			req.RejectReason, reviewerID, reviewTime, reviewTime, req.TaskID,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新审核状态失败: " + err.Error(),
		})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "审核失败: " + err.Error(),
		})
		return
	}

	message := "应用审核通过"
	if req.Accept == 0 {
		message = "应用审核拒绝"
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: message,
		Data: gin.H{
			"task_id":     req.TaskID,
			"status":      map[int]string{0: "rejected", 1: "approved"}[req.Accept],
			"review_time": reviewTime.Format("2006-01-02 15:04:05"),
		},
	})
}

// validateCategory 验证分类是否存在
func validateCategory(mainCategory, subCategory string) bool {
	subCategories, exists := appCategories[mainCategory]
	if !exists {
		return false
	}

	for _, sub := range subCategories {
		if sub == subCategory {
			return true
		}
	}
	return false
}
