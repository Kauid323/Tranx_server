package config

import (
	"log"
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	ServerPort   string
	DatabasePath string
	LogLevel     string
	MaxPageSize  int
	EnableCORS   bool
}

var AppConfig *Config

// InitConfig 初始化配置
func InitConfig() {
	AppConfig = &Config{
		ServerPort:   getEnv("SERVER_PORT", "4999"),
		DatabasePath: getEnv("DATABASE_PATH", "./taruapp.db"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		MaxPageSize:  getEnvAsInt("MAX_PAGE_SIZE", 100),
		EnableCORS:   getEnvAsBool("ENABLE_CORS", true),
	}

	log.Println("配置加载完成:")
	log.Printf("  服务器端口: %s", AppConfig.ServerPort)
	log.Printf("  数据库路径: %s", AppConfig.DatabasePath)
	log.Printf("  日志级别: %s", AppConfig.LogLevel)
	log.Printf("  最大分页大小: %d", AppConfig.MaxPageSize)
	log.Printf("  启用CORS: %v", AppConfig.EnableCORS)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt 获取整数类型的环境变量
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("警告: 无法解析环境变量 %s, 使用默认值 %d", key, defaultValue)
		return defaultValue
	}
	
	return value
}

// getEnvAsBool 获取布尔类型的环境变量
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("警告: 无法解析环境变量 %s, 使用默认值 %v", key, defaultValue)
		return defaultValue
	}
	
	return value
}

