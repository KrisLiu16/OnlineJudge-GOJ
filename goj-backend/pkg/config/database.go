package config

import (
	"fmt"
	"io"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	// 从环境变量获取数据库配置
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "123456"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "goj-mysql"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "goj"
	}

	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// 配置 GORM，禁用 SQL 日志
	config := &gorm.Config{
		Logger: logger.New(
			log.New(io.Discard, "", 0), // 使用 io.Discard 丢弃所有日志
			logger.Config{
				LogLevel: logger.Silent, // 设置为 Silent 级别
			},
		),
	}

	// 连接数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	// 获取底层的 *sql.DB 对象
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Database connected successfully")
	return nil
}
