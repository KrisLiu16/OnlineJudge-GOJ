package controllers

import (
	"goj/pkg/config"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ClearCache 手动清除所有缓存
func ClearCache(c *gin.Context) {
	ctx := c.Request.Context()

	if err := config.CleanupCache(ctx); err != nil {
		log.Printf("Error clearing cache: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "清除缓存失败",
			"error":   err.Error(),
		})
		return
	}

	// 添加成功日志
	log.Println("Cache cleared successfully")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "缓存已清除",
	})
}
