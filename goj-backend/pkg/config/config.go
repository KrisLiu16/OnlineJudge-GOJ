package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 初始化数据库连接
	InitDB()

	// 初始化语言配置
	if err := InitLanguageConfig(); err != nil {
		log.Fatalf("Failed to load language config: %v", err)
	}
}

var (
	JWTSecret = []byte(getEnvOrDefault("JWT_SECRET", "your-secret-key"))
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type Config struct {
	LogLevel string `yaml:"log_level"` // debug, info, warn, error
}
