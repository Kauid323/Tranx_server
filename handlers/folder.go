package handlers

import (
	"TaruApp/database"
	"TaruApp/models"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateFavoriteFolder 创建收藏夹
func CreateFavoriteFolder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req models.CreateFavoriteFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查收藏夹名称是否重复
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM favorite_folders WHERE user_id = ? AND name = ?",
		userID, req.Name,
	).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "收藏夹名称已存在",
		})
		return
	}

	// 创建收藏夹
	result, err := database.DB.Exec(
		"INSERT INTO favorite_folders (user_id, name, description, is_public) VALUES (?, ?, ?, ?)",
		userID, req.Name, req.Description, req.IsPublic,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建收藏夹失败: " + err.Error(),
		})
		return
	}

	folderID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "创建收藏夹成功",
		Data: gin.H{
			"folder_id": folderID,
		},
	})
}

// GetMyFavoriteFolders 获取我的收藏夹列表
func GetMyFavoriteFolders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	rows, err := database.DB.Query(`
		SELECT id, user_id, name, description, is_public, item_count, created_at, updated_at
		FROM favorite_folders
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询收藏夹失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var folders []models.FavoriteFolder
	for rows.Next() {
		var folder models.FavoriteFolder
		if err := rows.Scan(&folder.ID, &folder.UserID, &folder.Name, &folder.Description, &folder.IsPublic, &folder.ItemCount, &folder.CreatedAt, &folder.UpdatedAt); err != nil {
			continue
		}
		folders = append(folders, folder)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取收藏夹列表成功",
		Data:    folders,
	})
}

// GetUserFavoriteFolders 获取指定用户的公开收藏夹列表
func GetUserFavoriteFolders(c *gin.Context) {
	userID := c.Param("id")
	currentUserID, _ := c.Get("user_id")

	// 如果是查看自己的，显示所有收藏夹；否则只显示公开的
	query := `
		SELECT id, user_id, name, description, is_public, item_count, created_at, updated_at
		FROM favorite_folders
		WHERE user_id = ?`
	
	currentUserIDStr := strconv.FormatInt(currentUserID.(int64), 10)
	if currentUserIDStr != userID {
		query += " AND is_public = 1"
	}
	
	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询收藏夹失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var folders []models.FavoriteFolder
	for rows.Next() {
		var folder models.FavoriteFolder
		if err := rows.Scan(&folder.ID, &folder.UserID, &folder.Name, &folder.Description, &folder.IsPublic, &folder.ItemCount, &folder.CreatedAt, &folder.UpdatedAt); err != nil {
			continue
		}
		folders = append(folders, folder)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取收藏夹列表成功",
		Data:    folders,
	})
}

// UpdateFavoriteFolder 更新收藏夹
func UpdateFavoriteFolder(c *gin.Context) {
	folderID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req models.UpdateFavoriteFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查收藏夹是否属于当前用户
	var ownerID int64
	err := database.DB.QueryRow("SELECT user_id FROM favorite_folders WHERE id = ?", folderID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "收藏夹不存在",
		})
		return
	}
	if ownerID != userID.(int64) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权操作此收藏夹",
		})
		return
	}

	// 构建更新语句
	query := "UPDATE favorite_folders SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}

	if req.Name != "" {
		query += ", name = ?"
		args = append(args, req.Name)
	}
	if req.Description != "" {
		query += ", description = ?"
		args = append(args, req.Description)
	}
	if req.IsPublic != nil {
		query += ", is_public = ?"
		args = append(args, *req.IsPublic)
	}

	query += " WHERE id = ?"
	args = append(args, folderID)

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新收藏夹失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "更新收藏夹成功",
	})
}

// DeleteFavoriteFolder 删除收藏夹
func DeleteFavoriteFolder(c *gin.Context) {
	folderID := c.Param("id")
	userID, _ := c.Get("user_id")

	// 检查收藏夹是否属于当前用户
	var ownerID int64
	err := database.DB.QueryRow("SELECT user_id FROM favorite_folders WHERE id = ?", folderID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "收藏夹不存在",
		})
		return
	}
	if ownerID != userID.(int64) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权操作此收藏夹",
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除收藏夹失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 删除收藏夹中的所有项目
	_, err = tx.Exec("DELETE FROM favorite_items WHERE folder_id = ?", folderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除收藏夹失败: " + err.Error(),
		})
		return
	}

	// 删除收藏夹
	_, err = tx.Exec("DELETE FROM favorite_folders WHERE id = ?", folderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除收藏夹失败: " + err.Error(),
		})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除收藏夹失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "删除收藏夹成功",
	})
}

// AddPostToFolder 添加帖子到收藏夹
func AddPostToFolder(c *gin.Context) {
	folderID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req models.AddToFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	postID := req.PostID

	// 检查帖子是否存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ?", postID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "帖子不存在",
		})
		return
	}

	// 检查收藏夹是否属于当前用户
	var ownerID int64
	err = database.DB.QueryRow("SELECT user_id FROM favorite_folders WHERE id = ?", folderID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "收藏夹不存在",
		})
		return
	}
	if ownerID != userID.(int64) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权操作此收藏夹",
		})
		return
	}

	// 检查是否已收藏
	err = database.DB.QueryRow(
		"SELECT COUNT(*) FROM favorite_items WHERE folder_id = ? AND post_id = ?",
		folderID, postID,
	).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "该帖子已在此收藏夹中",
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "添加收藏失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 添加到收藏夹
	_, err = tx.Exec(
		"INSERT INTO favorite_items (folder_id, post_id) VALUES (?, ?)",
		folderID, postID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "添加收藏失败: " + err.Error(),
		})
		return
	}

	// 更新收藏夹项目数量
	_, err = tx.Exec(
		"UPDATE favorite_folders SET item_count = item_count + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		folderID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新收藏数失败: " + err.Error(),
		})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "添加收藏失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "添加收藏成功",
	})
}

// RemovePostFromFolder 从收藏夹移除帖子
func RemovePostFromFolder(c *gin.Context) {
	folderID := c.Param("id")
	postID := c.Param("post_id")
	userID, _ := c.Get("user_id")

	// 检查收藏夹是否属于当前用户
	var ownerID int64
	err := database.DB.QueryRow("SELECT user_id FROM favorite_folders WHERE id = ?", folderID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "收藏夹不存在",
		})
		return
	}
	if ownerID != userID.(int64) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权操作此收藏夹",
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "移除收藏失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 删除收藏项
	result, err := tx.Exec(
		"DELETE FROM favorite_items WHERE folder_id = ? AND post_id = ?",
		folderID, postID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "移除收藏失败: " + err.Error(),
		})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "该帖子不在此收藏夹中",
		})
		return
	}

	// 更新收藏夹项目数量
	_, err = tx.Exec(
		"UPDATE favorite_folders SET item_count = item_count - 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND item_count > 0",
		folderID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新收藏数失败: " + err.Error(),
		})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "移除收藏失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "移除收藏成功",
	})
}

// GetFolderPosts 获取收藏夹中的帖子列表
func GetFolderPosts(c *gin.Context) {
	folderID := c.Param("id")
	currentUserID, _ := c.Get("user_id")

	// 检查收藏夹是否存在及是否有权访问
	var folder models.FavoriteFolder
	err := database.DB.QueryRow(
		"SELECT id, user_id, name, description, is_public, item_count FROM favorite_folders WHERE id = ?",
		folderID,
	).Scan(&folder.ID, &folder.UserID, &folder.Name, &folder.Description, &folder.IsPublic, &folder.ItemCount)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "收藏夹不存在",
		})
		return
	}

	// 如果不是公开的，且不是收藏夹主人，则无权访问
	if !folder.IsPublic && folder.UserID != currentUserID.(int64) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "无权访问此收藏夹",
		})
		return
	}

	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 查询收藏夹中的帖子
	rows, err := database.DB.Query(`
		SELECT p.id, p.board_id, p.user_id, p.title, p.content, p.type, p.publisher, p.publish_time, 
		       p.coins, p.favorites, p.likes, p.image_url, p.attachment_url, p.attachment_type,
		       p.comment_count, p.view_count, p.last_reply_time, p.created_at, p.updated_at
		FROM favorite_items fi
		JOIN posts p ON fi.post_id = p.id
		WHERE fi.folder_id = ?
		ORDER BY fi.created_at DESC
		LIMIT ? OFFSET ?
	`, folderID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询收藏帖子失败: " + err.Error(),
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
		Message: "获取收藏夹帖子成功",
		Data: gin.H{
			"folder": folder,
			"posts": models.PageData{
				Total:    folder.ItemCount,
				Page:     page,
				PageSize: pageSize,
				List:     posts,
			},
		},
	})
}

// GetViewHistory 获取浏览历史
func GetViewHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 查询浏览历史总数
	var total int
	err := database.DB.QueryRow(
		"SELECT COUNT(DISTINCT post_id) FROM view_histories WHERE user_id = ?",
		userID,
	).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}

	// 查询浏览历史（按最后浏览时间排序，去重）
	rows, err := database.DB.Query(`
		SELECT p.id, p.board_id, p.user_id, p.title, p.content, p.publisher, p.publish_time, 
		       p.coins, p.favorites, p.likes, p.image_url, p.attachment_url, p.attachment_type,
		       p.comment_count, p.view_count, p.last_reply_time, p.created_at, p.updated_at,
		       vh.viewed_at
		FROM (
			SELECT post_id, MAX(viewed_at) as viewed_at
			FROM view_histories
			WHERE user_id = ?
			GROUP BY post_id
		) vh
		JOIN posts p ON vh.post_id = p.id
		ORDER BY vh.viewed_at DESC
		LIMIT ? OFFSET ?
	`, userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询浏览历史失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var post models.Post
		var imageURL, attachmentURL, attachmentType sql.NullString
		var viewedAt string
		err := rows.Scan(
			&post.ID, &post.BoardID, &post.UserID, &post.Title, &post.Content, &post.Publisher,
			&post.PublishTime, &post.Coins, &post.Favorites, &post.Likes,
			&imageURL, &attachmentURL, &attachmentType, &post.CommentCount, &post.ViewCount, &post.LastReplyTime,
			&post.CreatedAt, &post.UpdatedAt, &viewedAt,
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

		history = append(history, map[string]interface{}{
			"post":      post,
			"viewed_at": viewedAt,
		})
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取浏览历史成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     history,
		},
	})
}


