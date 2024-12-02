package controllers

import (
	"goj/pkg/config"
	"goj/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPublicWebsiteSettings 获取公开的网站设置
func GetPublicWebsiteSettings(c *gin.Context) {
	var settings models.WebsiteSetting

	// 尝试从缓存获取
	settingsJSON, err := config.RDB.Get(c.Request.Context(), "website:settings").Result()
	if err == nil {
		c.Data(http.StatusOK, "application/json", []byte(settingsJSON))
		return
	}

	// 从数据库获取
	result := config.DB.First(&settings)
	if result.Error != nil {
		settings = models.WebsiteSetting{
			Title:    "GO! Judge",
			Subtitle: "快速、智能的在线评测系统",
			About:    "GOJ是一个高性能在线评测的平台，致力于提供快速、稳定的评测服务。",
			Email:    "support@example.com",
			Github:   "https://github.com/yourusername",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    settings,
	})
}
