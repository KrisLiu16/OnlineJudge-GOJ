package controllers

import (
	"goj/pkg/config"
	"goj/pkg/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProblemSubmissions 获取题目的提交记录
func GetProblemSubmissions(c *gin.Context) {
	problemID := c.Param("problemId")
	page, pageSize := getPaginationParams(c)

	var submissions []struct {
		models.Submission
		Username   string `json:"username"`
		UserAvatar string `json:"userAvatar"`
	}
	var total int64

	query := config.DB.Model(&models.Submission{}).
		Select("submissions.*, users.username, users.avatar as user_avatar").
		Joins("LEFT JOIN users ON submissions.user_id = users.id").
		Where("problem_id = ?", problemID)
	// 获取用户ID（如果已登录）
	var userID uint
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
		log.Printf("Debug - User is logged in with userID: %d", userID)
	} else {
		log.Printf("Debug - No user logged in")
	}
	// 获取用户角色
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
			"data":    nil,
		})
		return
	}
	// 如果不是管理员，过滤掉管理员的提交
	if role != "admin" {
		query = query.Where("users.role != ? AND users.id != ?", "admin", userID)
	}
	// 获取总数
	query.Count(&total)

	// 获取分页数据
	if err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("submit_time DESC").
		Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取提交记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"submissions": submissions,
			"total":       total,
		},
	})
}

// GetUserSubmissions 获取用户的提交记录
func GetUserSubmissions(c *gin.Context) {
	username := c.Param("username")
	page, pageSize := getPaginationParams(c)

	var submissions []struct {
		models.Submission
		Username   string `json:"username"`
		UserAvatar string `json:"userAvatar"`
	}
	var total int64

	query := config.DB.Model(&models.Submission{}).
		Select("submissions.*, users.username, users.avatar as user_avatar").
		Joins("LEFT JOIN users ON submissions.user_id = users.id").
		Where("users.username = ?", username)

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
			"data":    nil,
		})
		return
	}
	// 获取用户ID（如果已登录）
	var userID uint
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
		log.Printf("Debug - User is logged in with userID: %d", userID)
	} else {
		log.Printf("Debug - No user logged in")
	}
	// 如果不是管理员，过滤掉管理员的提交
	if role != "admin" {
		query = query.Where("users.role != ? AND users.id != ?", "admin", userID)
	}
	// 获取总数
	query.Count(&total)

	if err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("submit_time DESC").
		Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取提交记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"submissions": submissions,
			"total":       total,
		},
	})
}

// GetContestSubmissions 获取比赛的提交记录
func GetContestSubmissions(c *gin.Context) {
	contestID := c.Param("contestId")
	page, pageSize := getPaginationParams(c)

	var submissions []struct {
		models.Submission
		Username   string `json:"username"`
		UserAvatar string `json:"userAvatar"`
	}
	var total int64

	query := config.DB.Model(&models.Submission{}).
		Select("submissions.*, users.username, users.avatar as user_avatar").
		Joins("LEFT JOIN users ON submissions.user_id = users.id").
		Where("contest_id = ?", contestID)

	// 获取用户角色
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
			"data":    nil,
		})
		return
	}
	// 获取用户ID（如果已登录）
	var userID uint
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
		log.Printf("Debug - User is logged in with userID: %d", userID)
	} else {
		log.Printf("Debug - No user logged in")
	}
	// 如果不是管理员，过滤掉管理员的提交
	if role != "admin" {
		query = query.Where("users.role != ? AND users.id != ?", "admin", userID)
	}
	// 获取总数
	query.Count(&total)

	if err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("submit_time DESC").
		Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取提交记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"submissions": submissions,
			"total":       total,
		},
	})
}
