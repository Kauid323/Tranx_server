package handlers

import (
	"TaruApp/database"
	"TaruApp/models"
	"TaruApp/utils"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证用户名长度（3-20个字符）
	if len(req.Username) < 3 || len(req.Username) > 20 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "用户名长度必须在3-20个字符之间",
		})
		return
	}

	// 验证密码长度（至少8位）
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "密码长度至少为8位",
		})
		return
	}

	// 检查用户名是否已存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询用户失败: " + err.Error(),
		})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "用户名已存在",
		})
		return
	}

	// 密码哈希
	hashedPassword := utils.HashPassword(req.Password)

	// 创建用户
	result, err := database.DB.Exec(
		"INSERT INTO users (username, password, email, avatar, level) VALUES (?, ?, ?, ?, ?)",
		req.Username, hashedPassword, req.Email, req.Avatar, 0,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建用户失败: " + err.Error(),
		})
		return
	}

	userID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "注册成功",
		Data: gin.H{
			"user_id":  userID,
			"username": req.Username,
		},
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证用户
	hashedPassword := utils.HashPassword(req.Password)
	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, level, avatar, coins, exp, user_level, created_at FROM users WHERE username = ? AND password = ?",
		req.Username, hashedPassword,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Level, &user.Avatar, &user.Coins, &user.Exp, &user.UserLevel, &user.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "用户名或密码错误",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "登录失败: " + err.Error(),
		})
		return
	}

	// 生成token
	rawToken, err := utils.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "生成令牌失败: " + err.Error(),
		})
		return
	}

	// RC4加密token
	encryptedToken, err := utils.RC4Encrypt(rawToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "加密令牌失败: " + err.Error(),
		})
		return
	}

	// 保存token到数据库（有效期30天）
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = database.DB.Exec(
		"INSERT INTO tokens (user_id, token, expires_at) VALUES (?, ?, ?)",
		user.ID, encryptedToken, expiresAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "保存令牌失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "登录成功",
		Data: gin.H{
			"token":      encryptedToken,
			"user":       user,
			"expires_at": expiresAt,
		},
	})
}

// GetUserInfo 获取用户信息
func GetUserInfo(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, level, avatar, coins, exp, user_level, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Level, &user.Avatar, &user.Coins, &user.Exp, &user.UserLevel, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "用户不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询用户失败: " + err.Error(),
		})
		return
	}

	// 获取用户标签
	rows, err := database.DB.Query(
		"SELECT id, user_id, tag_name, tag_color, created_at FROM user_tags WHERE user_id = ?",
		userID,
	)

	var tags []models.UserTag
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tag models.UserTag
			if err := rows.Scan(&tag.ID, &tag.UserID, &tag.TagName, &tag.TagColor, &tag.CreatedAt); err == nil {
				tags = append(tags, tag)
			}
		}
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取用户信息成功",
		Data: gin.H{
			"user": user,
			"tags": tags,
		},
	})
}

// GetUserDetail 获取用户详情（包含统计信息、发布的帖子和收藏）
func GetUserDetail(c *gin.Context) {
	userID := c.Param("id")

	// 获取用户基本信息
	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, level, avatar, coins, exp, user_level, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Level, &user.Avatar, &user.Coins, &user.Exp, &user.UserLevel, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "用户不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询用户失败: " + err.Error(),
		})
		return
	}

	// 获取关注数
	var followingCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE user_id = ?", userID).Scan(&followingCount)

	// 获取粉丝数
	var followerCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE followed_id = ?", userID).Scan(&followerCount)

	// 获取发帖数
	var postCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ?", userID).Scan(&postCount)

	// 获取收藏数（统计所有收藏夹中的帖子总数）
	var favoriteCount int
	database.DB.QueryRow(`
		SELECT COUNT(DISTINCT post_id) 
		FROM favorite_items fi
		JOIN favorite_folders ff ON fi.folder_id = ff.id
		WHERE ff.user_id = ?
	`, userID).Scan(&favoriteCount)

	// 获取用户的收藏夹列表
	folderRows, err := database.DB.Query(`
		SELECT id, user_id, name, description, is_public, item_count, created_at, updated_at
		FROM favorite_folders
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 5
	`, userID)

	var folders []models.FavoriteFolder
	if err == nil {
		defer folderRows.Close()
		for folderRows.Next() {
			var folder models.FavoriteFolder
			if err := folderRows.Scan(&folder.ID, &folder.UserID, &folder.Name, &folder.Description, &folder.IsPublic, &folder.ItemCount, &folder.CreatedAt, &folder.UpdatedAt); err == nil {
				folders = append(folders, folder)
			}
		}
	}

	// 获取最近发布的帖子（最多10条）
	rows, err := database.DB.Query(`
		SELECT id, board_id, user_id, title, content, publisher, publish_time, 
		       coins, favorites, likes, image_url, attachment_url, attachment_type,
		       comment_count, view_count, last_reply_time, created_at, updated_at
		FROM posts
		WHERE user_id = ?
		ORDER BY publish_time DESC
		LIMIT 10
	`, userID)

	var posts []models.Post
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var post models.Post
			var imageURL, attachmentURL, attachmentType sql.NullString
			err := rows.Scan(
				&post.ID, &post.BoardID, &post.UserID, &post.Title, &post.Content, &post.Publisher,
				&post.PublishTime, &post.Coins, &post.Favorites, &post.Likes,
				&imageURL, &attachmentURL, &attachmentType, &post.CommentCount, &post.ViewCount, &post.LastReplyTime,
				&post.CreatedAt, &post.UpdatedAt,
			)
			if err == nil {
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
		}
	}

	// 获取最近收藏的帖子（最多10条）
	rows2, err := database.DB.Query(`
		SELECT p.id, p.board_id, p.user_id, p.title, p.content, p.publisher, p.publish_time, 
		       p.coins, p.favorites, p.likes, p.image_url, p.attachment_url, p.attachment_type,
		       p.comment_count, p.view_count, p.last_reply_time, p.created_at, p.updated_at
		FROM favorite_items fi
		JOIN posts p ON fi.post_id = p.id
		JOIN favorite_folders ff ON fi.folder_id = ff.id
		WHERE ff.user_id = ?
		ORDER BY fi.created_at DESC
		LIMIT 10
	`, userID)

	var favorites []models.Post
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var post models.Post
			var imageURL, attachmentURL, attachmentType sql.NullString
			err := rows2.Scan(
				&post.ID, &post.BoardID, &post.UserID, &post.Title, &post.Content, &post.Publisher,
				&post.PublishTime, &post.Coins, &post.Favorites, &post.Likes,
				&imageURL, &attachmentURL, &attachmentType, &post.CommentCount, &post.ViewCount, &post.LastReplyTime,
				&post.CreatedAt, &post.UpdatedAt,
			)
			if err == nil {
				if imageURL.Valid {
					post.ImageURL = imageURL.String
				}
				if attachmentURL.Valid {
					post.AttachmentURL = attachmentURL.String
				}
				if attachmentType.Valid {
					post.AttachmentType = attachmentType.String
				}
				favorites = append(favorites, post)
			}
		}
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取用户详情成功",
		Data: gin.H{
			"user":            user,
			"coins":           user.Coins,
			"following_count": followingCount,
			"follower_count":  followerCount,
			"post_count":      postCount,
			"favorite_count":  favoriteCount,
			"folders":         folders,
			"posts":           posts,
			"favorites":       favorites,
		},
	})
}

// GetCurrentUser 获取当前登录用户信息
func GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "未授权访问",
		})
		return
	}

	currentUser := user.(models.User)

	// 获取用户标签
	rows, err := database.DB.Query(
		"SELECT id, user_id, tag_name, tag_color, created_at FROM user_tags WHERE user_id = ?",
		currentUser.ID,
	)

	var tags []models.UserTag
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tag models.UserTag
			if err := rows.Scan(&tag.ID, &tag.UserID, &tag.TagName, &tag.TagColor, &tag.CreatedAt); err == nil {
				tags = append(tags, tag)
			}
		}
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取当前用户信息成功",
		Data: gin.H{
			"user": currentUser,
			"tags": tags,
		},
	})
}

// SetUserLevel 设置用户等级（管理员权限）
func SetUserLevel(c *gin.Context) {
	userID := c.Param("id")
	var req models.SetUserLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证等级值
	if req.Level != 0 && req.Level != 50 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "等级值无效，只能设置为0(普通用户)或50(管理员)",
		})
		return
	}

	// 更新用户等级
	_, err := database.DB.Exec(
		"UPDATE users SET level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		req.Level, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "设置用户等级失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "设置用户等级成功",
	})
}

// Logout 退出登录
func Logout(c *gin.Context) {
	token := c.GetHeader("Token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "未提供令牌",
		})
		return
	}

	// 删除token
	_, err := database.DB.Exec("DELETE FROM tokens WHERE token = ?", token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "退出登录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "退出登录成功",
	})
}

// CreateUserTag 创建用户标签（管理员权限）
func CreateUserTag(c *gin.Context) {
	var req models.CreateUserTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户是否存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", req.UserID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "用户不存在",
		})
		return
	}

	// 创建标签
	result, err := database.DB.Exec(
		"INSERT INTO user_tags (user_id, tag_name, tag_color) VALUES (?, ?, ?)",
		req.UserID, req.TagName, req.TagColor,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建标签失败: " + err.Error(),
		})
		return
	}

	tagID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "创建标签成功",
		Data: gin.H{
			"tag_id": tagID,
		},
	})
}

// DeleteUserTag 删除用户标签（管理员权限）
func DeleteUserTag(c *gin.Context) {
	tagID := c.Param("id")

	_, err := database.DB.Exec("DELETE FROM user_tags WHERE id = ?", tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除标签失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "删除标签成功",
	})
}

// GetUserTags 获取用户的所有标签
func GetUserTags(c *gin.Context) {
	userID := c.Param("id")

	rows, err := database.DB.Query(
		"SELECT id, user_id, tag_name, tag_color, created_at FROM user_tags WHERE user_id = ?",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询标签失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var tags []models.UserTag
	for rows.Next() {
		var tag models.UserTag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.TagName, &tag.TagColor, &tag.CreatedAt); err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取标签列表成功",
		Data:    tags,
	})
}

// GetAllUsers 获取所有用户列表
func GetAllUsers(c *gin.Context) {
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

	// 查询用户总数
	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
		})
		return
	}

	// 查询用户列表（按ID升序排列）
	rows, err := database.DB.Query(`
		SELECT id, username, email, level, avatar, coins, exp, user_level, created_at, updated_at
		FROM users
		ORDER BY id ASC
		LIMIT ? OFFSET ?
	`, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询用户列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Level, &user.Avatar, &user.Coins, &user.Exp, &user.UserLevel, &user.CreatedAt, &user.UpdatedAt); err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取用户列表成功",
		Data: models.PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     users,
		},
	})
}
