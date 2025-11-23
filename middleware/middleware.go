package middleware

import (
	"TaruApp/database"
	"TaruApp/models"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(startTime)

		// 获取状态码
		statusCode := c.Writer.Status()

		// 记录日志
		log.Printf("[%s] %s %s %d %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			statusCode,
			latency,
		)
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// AuthRequired Token认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		token := c.GetHeader("Token")
		if token == "" {
			c.JSON(401, models.Response{
				Code:    401,
				Message: "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 在数据库中验证token
		var tokenData models.Token
		var user models.User
		err := database.DB.QueryRow(`
			SELECT t.id, t.user_id, t.token, t.expires_at, t.created_at,
			       u.id, u.username, u.level, u.avatar, u.coins, u.exp, u.user_level
			FROM tokens t
			JOIN users u ON t.user_id = u.id
			WHERE t.token = ? AND t.expires_at > datetime('now')
		`, token).Scan(
			&tokenData.ID, &tokenData.UserID, &tokenData.Token,
			&tokenData.ExpiresAt, &tokenData.CreatedAt,
			&user.ID, &user.Username, &user.Level, &user.Avatar, &user.Coins, &user.Exp, &user.UserLevel,
		)

		if err != nil {
			c.JSON(401, models.Response{
				Code:    401,
				Message: "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("user_level", user.Level)
		c.Set("user", user)

		c.Next()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userLevel, exists := c.Get("user_level")
		if !exists {
			c.JSON(403, models.Response{
				Code:    403,
				Message: "未授权访问",
			})
			c.Abort()
			return
		}

		if userLevel.(int) < 50 {
			c.JSON(403, models.Response{
				Code:    403,
				Message: "需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ReviewerRequired 审核权限中间件
func ReviewerRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		levelValue, exists := c.Get("level")
		if !exists {
			c.JSON(403, models.Response{
				Code:    403,
				Message: "未授权访问",
			})
			c.Abort()
			return
		}

		// level >= 80 表示有审核权限
		if levelValue.(int) < 80 {
			c.JSON(403, models.Response{
				Code:    403,
				Message: "需要审核权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiter 简单的限流中间件
func RateLimiter() gin.HandlerFunc {
	// 这里可以实现更复杂的限流逻辑
	return func(c *gin.Context) {
		c.Next()
	}
}

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("错误: %v", err)
		}
	}
}
