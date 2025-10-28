package routes

import (
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/auth"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/controllers"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/judge/handler"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/middleware"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/rank"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// 添加工作目录日志
	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get working directory: %v", err)
	} else {
		log.Printf("Working directory: %s", pwd)
	}

	// 修改静态文件服务配置 - 将这部分移到最前面
	// 确保目录存在
	os.MkdirAll("public/images/avatars", 0755)

	// 使用绝对路径配置静态文件服务
	publicDir := filepath.Join(pwd, "public")
	r.Static("/images", publicDir+"/images")

	// 添加调试日志
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/images/") {
			log.Printf("Static file request: %s", c.Request.URL.Path)
			filePath := filepath.Join(publicDir, c.Request.URL.Path)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				log.Printf("File not found: %s", filePath)
			} else {
				log.Printf("File exists: %s", filePath)
			}
		}
		c.Next()
	})

	// 其他中间件
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORSMiddleware())

	// 将讨论路由的设置移到其他路由之前
	SetupDiscussRoutes(r)

	// API 路由组
	api := r.Group("/api")

	// 公开路由 - 只保留登录、注册和排名
	public := api.Group("")
	{
		public.POST("/auth/login", auth.Login)
		public.POST("/auth/register", auth.Register)
		public.GET("/rank", rank.GetRankList)
		public.GET("/judge/status", controllers.GetJudgeStatus)
	}

	// 需要认证的路由
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// 提交相关路由
		protected.GET("/submissions", controllers.GetSubmissions)         // 通用查询，支持所有查询参数
		protected.GET("/submission/:ID", controllers.GetSubmissionDetail) // 获取单个提交
		protected.POST("/submit", controllers.CreateSubmission)           // 创建提

		// 用户相关路由
		protected.GET("/user/profile", auth.GetProfile)
		protected.PUT("/user/profile", auth.UpdateProfile)
		protected.POST("/user/avatar", auth.UploadAvatar)
		protected.PUT("/user/password", auth.UpdatePassword)
		protected.GET("/users/profile/:username", auth.GetUserProfile)
		protected.GET("/users/:username/submissions", auth.GetUserSubmissions)
		protected.GET("/users/:username/solved-problems", auth.GetUserSolvedProblems)
		protected.GET("/users/:username/contests", auth.GetUserContests)

		// 题目相关路由
		protected.POST("/problems/add", controllers.AddProblem)
		protected.GET("/problems", controllers.GetProblems)
		protected.GET("/problems/:id", controllers.GetProblemDetail)

		// 比赛相关路由
		protected.GET("/contests", controllers.GetContests)
		protected.POST("/contests", middleware.AdminRequired(), controllers.CreateContest)
		protected.GET("/contests/problems/:contestId", controllers.GetContestProblems)
		protected.GET("/contests/problem/:contestId/:problemId", controllers.GetContestProblem)
		protected.PUT("/contests/:id", middleware.AdminRequired(), controllers.UpdateContest)
		protected.DELETE("/contests/:id", middleware.AdminRequired(), controllers.DeleteContest)
		protected.GET("/contests/:id/rank", controllers.GetContestRank)
		protected.GET("/contests/:id", controllers.GetContest)
		protected.GET("/contests/:id/rank/export", controllers.ExportContestRank)

		// WebSocket 路由
		protected.GET("/ws", func(c *gin.Context) {
			// 从认证中间件获取用户ID
			userID, exists := c.Get("userID")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "未授权访问",
				})
				return
			}

			// 类型转换
			uid, ok := userID.(uint)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "用户ID类型错误",
				})
				return
			}

			// 处理 WebSocket 连接
			handler.HandleWebSocket(c.Writer, c.Request, uid)
		})
	}

	// 公开的网站设置路由
	r.GET("/api/website/settings", controllers.GetPublicWebsiteSettings)

	// 移除 /help 路由的处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "页面不存在",
			"data":    nil,
		})
	})

	// 替换默认的日志中间件
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: io.Discard, // 完全禁用请求日志
	}))

	// 或者完全移除日志中间件
	// r.Use(gin.Recovery()) // 只保留 Recovery 中间件

	// 添加健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}
