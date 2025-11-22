package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// RC4Key RC4加密的密钥
const RC4Key = "TaruApp2025SecretKey"

// GenerateToken 生成随机token
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// RC4Encrypt RC4加密
func RC4Encrypt(plaintext string) (string, error) {
	cipher, err := rc4.NewCipher([]byte(RC4Key))
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plaintext))
	cipher.XORKeyStream(ciphertext, []byte(plaintext))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RC4Decrypt RC4解密
func RC4Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	cipher, err := rc4.NewCipher([]byte(RC4Key))
	if err != nil {
		return "", err
	}

	plaintext := make([]byte, len(data))
	cipher.XORKeyStream(plaintext, data)

	return string(plaintext), nil
}

// HashPassword 使用bcrypt加密密码
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword 验证密码
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// CalculateUserLevel 根据经验值计算用户等级
// 等级公式: Lv = floor(sqrt(exp / 100)) + 1
// Lv1: 0-99 exp
// Lv2: 100-399 exp
// Lv3: 400-899 exp
// Lv4: 900-1599 exp
// ...
func CalculateUserLevel(exp int) int {
	if exp < 0 {
		exp = 0
	}
	// 使用平方根公式计算等级
	level := int(float64(exp) / 100.0)
	sqrtLevel := 1
	for sqrtLevel*sqrtLevel <= level {
		sqrtLevel++
	}
	return sqrtLevel
}

// GetExpForNextLevel 获取升级到下一级所需的总经验值
func GetExpForNextLevel(currentLevel int) int {
	if currentLevel < 1 {
		currentLevel = 1
	}
	nextLevel := currentLevel + 1
	return (nextLevel - 1) * (nextLevel - 1) * 100
}

// GetExpProgress 获取当前等级的经验进度
func GetExpProgress(exp int, currentLevel int) (currentLevelExp, nextLevelExp, progress int) {
	currentLevelExp = (currentLevel - 1) * (currentLevel - 1) * 100
	nextLevelExp = GetExpForNextLevel(currentLevel)
	expInCurrentLevel := exp - currentLevelExp
	expNeeded := nextLevelExp - currentLevelExp
	if expNeeded > 0 {
		progress = (expInCurrentLevel * 100) / expNeeded
	}
	return currentLevelExp, nextLevelExp, progress
}

// GenerateID 生成唯一ID（基于时间戳和MD5）
func GenerateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%s-%d", prefix, timestamp)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// TimeSince 计算时间差（友好显示）
func TimeSince(t time.Time) string {
	duration := time.Since(t)

	if duration.Hours() < 1 {
		minutes := int(duration.Minutes())
		if minutes < 1 {
			return "刚刚"
		}
		return fmt.Sprintf("%d分钟前", minutes)
	}

	if duration.Hours() < 24 {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d小时前", hours)
	}

	days := int(duration.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%d天前", days)
	}

	if days < 365 {
		months := days / 30
		return fmt.Sprintf("%d个月前", months)
	}

	years := days / 365
	return fmt.Sprintf("%d年前", years)
}

// ValidateString 验证字符串长度
func ValidateString(str string, minLen, maxLen int) bool {
	length := len([]rune(str))
	return length >= minLen && length <= maxLen
}

// Contains 检查切片是否包含指定元素
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Pagination 计算分页参数
func Pagination(page, pageSize, total int) (offset, totalPages int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset = (page - 1) * pageSize
	totalPages = (total + pageSize - 1) / pageSize

	return offset, totalPages
}

// TruncateString 截断字符串
func TruncateString(str string, maxLen int) string {
	runes := []rune(str)
	if len(runes) <= maxLen {
		return str
	}
	return string(runes[:maxLen]) + "..."
}
