package controllers

import (
	"encoding/json"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/config"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetPublicWebsiteSettings 获取公开的网站设置
func GetPublicWebsiteSettings(c *gin.Context) {
	// log.Printf("收到获取公开网站设置请求")

	var settings models.WebsiteSetting

	// 尝试从缓存获取
	settingsJSON, err := config.RDB.Get(c.Request.Context(), "website:settings").Result()
	if err == nil {
		// log.Printf("从缓存获取设置成功")
		c.Data(http.StatusOK, "application/json", []byte(settingsJSON))
		return
	}

	// 从数据库获取
	result := config.DB.First(&settings)
	if result.Error != nil {
		// log.Printf("数据库中未找到设置，使用默认值")
		settings = models.WebsiteSetting{
			Title:    "GO! Judge",
			Subtitle: "快速、智能的在线评测系统",
			About:    "GOJ是一个高性能在线评测的平台，致力于提供快速、稳定的评测服务。",
			Email:    "support@example.com",
			Github:   "https://github.com/yourusername",
			ICP:      "",
			ICPLink:  "",
			Feature1: "<div class=\"feature-icon\"><span class=\"icon-wrapper\">📚</span></div><h3>丰富的题库</h3><p>包含各种难度的编程题目，从入门到进阶</p>",
			Feature2: "<div class=\"feature-icon\"><span class=\"icon-wrapper\">🚀</span></div><h3>实时评测</h3><p>快速的代码执行和结果反馈</p>",
			Feature3: "<div class=\"feature-icon\"><span class=\"icon-wrapper\">👥</span></div><h3>社区讨论</h3><p>与其他同学交流学习心得</p>",
		}
	} else {
		// log.Printf("从数据库获取设置成功")
	}

	// 将设置缓存到 Redis，设置 24 小时过期
	response := gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    settings,
	}

	// 序列化响应数据
	jsonBytes, err := json.Marshal(response)
	if err == nil {
		config.RDB.Set(c.Request.Context(), "website:settings", string(jsonBytes), 24*time.Hour)
	}

	c.JSON(http.StatusOK, response)
}
