package routes

import (
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/controllers"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupDiscussRoutes(router *gin.Engine) {
	discuss := router.Group("/api/discussions")
	{
		discuss.GET("", controllers.GetDiscussions)
		discuss.GET("/:id", controllers.GetDiscussion)
		discuss.GET("/:id/comments", controllers.GetComments)
		discuss.GET("/:id/interactions", middleware.AuthMiddleware(), controllers.GetUserInteractions)

		// 需要登录的路由
		auth := discuss.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			auth.POST("", controllers.CreateDiscussion)
			auth.PUT("/:id", controllers.UpdateDiscussion)
			auth.DELETE("/:id", controllers.DeleteDiscussion)
			auth.POST("/:id/comments", controllers.CreateComment)
			auth.POST("/:id/like", controllers.ToggleLike)
			auth.DELETE("/:id/like", controllers.ToggleLike)
			auth.POST("/:id/star", controllers.ToggleStar)
			auth.DELETE("/:id/star", controllers.ToggleStar)
		}
	}
}
