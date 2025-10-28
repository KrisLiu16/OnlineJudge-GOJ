package routes

import (
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/controllers"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(router *gin.Engine) {
	admin := router.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/stats", controllers.GetAdminStats)
		// 系统还原
		admin.POST("/system/restore", controllers.RestoreSystem)
		// 题目管理
		admin.GET("/problems", controllers.GetAdminProblems)
		admin.DELETE("/problems/:id", controllers.DeleteProblem)
		admin.GET("/problems/:id", controllers.GetProblemDetail)
		admin.PUT("/problems/:id", controllers.UpdateProblem)
		admin.GET("/problems/:id/spj", controllers.GetProblemSPJCode)

		// 添加清除缓存的路由
		admin.POST("/cache/clear", controllers.ClearCache)

		// 题目数据管理
		admin.GET("/problem/data/:id", controllers.GetProblemDataFiles)
		admin.POST("/problem/data/upload", controllers.UploadProblemData)
		admin.GET("/problem/data/:id/:filename", controllers.GetProblemDataFile)
		admin.GET("/problem/data/:id/:filename/download", controllers.DownloadProblemDataFile)
		admin.GET("/problem/data/:id/download-all", controllers.DownloadAllProblemData)
		admin.DELETE("/problem/data/:id/:filename", controllers.DeleteProblemDataFile)
		admin.DELETE("/problem/data/:id/batch", controllers.BatchDeleteProblemData)

		// 用户管理
		admin.GET("/users", controllers.GetUsers)
		admin.PUT("/users/:id/role", controllers.SetUserRole)
		admin.DELETE("/users/:id", controllers.DeleteUser)
		admin.POST("/users/batch-create", controllers.BatchCreateUsers)

		// 比赛管理
		admin.POST("/contest/:id/update-rating", controllers.UpdateContestRating)
		admin.POST("/contests/:contestId/open-submissions", controllers.OpenContestSubmissions)

		// 题目导入导出
		admin.POST("/problems/import", controllers.ImportProblems)
		admin.POST("/problems/:id/export", controllers.ExportProblem)
		admin.POST("/problems/export-batch", controllers.ExportBatchProblems)
		admin.POST("/problems/export-all", controllers.ExportAllProblems)

		// 网站设置
		admin.GET("/website/settings", controllers.GetWebsiteSettings)
		admin.POST("/website/settings", controllers.UpdateWebsiteSettings)
	}
}
