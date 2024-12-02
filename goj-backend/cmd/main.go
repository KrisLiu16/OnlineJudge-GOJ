package main

import (
	"goj/pkg/config"
	"goj/pkg/judge"
	"goj/pkg/judge/handler"
	"goj/pkg/rank"
	"goj/pkg/routes"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置 Gin 为 Release 模式，关闭默认日志
	gin.SetMode(gin.ReleaseMode)

	// 如果需要完全禁用 Gin 的日志输出
	gin.DefaultWriter = io.Discard

	// 初始化配置
	config.Init()
	config.InitRedis()
	config.InitJudgeConfig()

	// 初始化 WebSocket 管理器
	handler.InitWebSocketManager()

	// 初始化评测系统
	if err := judge.Init(); err != nil {
		log.Fatalf("Failed to initialize judge system: %v", err)
	}

	// 初始化排行榜更新任务
	rank.InitRankUpdateTask()

	// 初始化数据库
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 执行数据库迁移
	if err := config.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 创建 Gin 实例
	r := gin.Default()

	// 设置路由
	routes.SetupRoutes(r)
	routes.SetupAdminRoutes(r)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
