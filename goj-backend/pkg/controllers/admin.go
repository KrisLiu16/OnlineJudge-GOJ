package controllers

import (
	"context"
	"encoding/json"
	"goj/pkg/config"
	"goj/pkg/models"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

// GetAdminStats 获取管理后台统计数据
func GetAdminStats(c *gin.Context) {
	ctx := c.Request.Context()

	var response struct {
		BasicStats struct {
			ProblemCount     int64 `json:"problemCount"`
			UserCount        int64 `json:"userCount"`
			TodaySubmissions int64 `json:"todaySubmissions"`
		} `json:"basicStats"`
		SystemStats *SystemStats `json:"systemStats"`
	}

	// 获取基础统计数据
	problemCount, err := GetProblemCount(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取题目统计失败",
			"data":    nil,
		})
		return
	}
	response.BasicStats.ProblemCount = problemCount

	userCount, err := GetUserCount(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户统计失败",
			"data":    nil,
		})
		return
	}
	response.BasicStats.UserCount = userCount

	todaySubmissions, err := GetTodaySubmissionCount(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取提交统计失败",
			"data":    nil,
		})
		return
	}
	response.BasicStats.TodaySubmissions = todaySubmissions

	// 获取系统状态
	systemStats, err := GetSystemStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取系统状态失败",
			"data":    nil,
		})
		return
	}
	response.SystemStats = systemStats

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    response,
	})
}

// RestoreSystem 系统还原
func RestoreSystem(c *gin.Context) {
	// 检查是否是超级管理员
	userRole := c.GetString("role")
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足",
			"data":    nil,
		})
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "开启事务失败",
			"data":    nil,
		})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取当前数据库中的所有表
	var tables []string
	err := tx.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		AND table_type = 'BASE TABLE'
	`).Scan(&tables).Error

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取表列表失败",
			"data":    nil,
		})
		return
	}

	// 先禁用外键约束
	if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "禁用外键约束失败",
			"data":    nil,
		})
		return
	}

	// 删除所有表
	for _, table := range tables {
		log.Printf("正在删除表: %s", table)
		if err := tx.Exec("DROP TABLE IF EXISTS `" + table + "`").Error; err != nil {
			log.Printf("删除表 %s 失败: %v", table, err)
		}
	}

	// 重新启用外键约束
	if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "启用外键约束失败",
			"data":    nil,
		})
		return
	}

	// 清空Redis缓存
	if err := config.RDB.FlushDB(c.Request.Context()).Err(); err != nil {
		log.Printf("清空Redis缓存失败: %v", err)
	}

	// 删除data文件夹
	if err := os.RemoveAll("data"); err != nil {
		log.Printf("删除data文件夹失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败",
			"data":    nil,
		})
		return
	}

	// 重新运行数据库迁移
	if err := config.AutoMigrate(); err != nil {
		log.Printf("重新创建表结构失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "系统已还原",
		"data":    nil,
	})
}

// GetUsers 获取用户列表
func GetUsers(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")

	offset := (page - 1) * pageSize

	// 构建查询
	query := config.DB.Model(&models.User{})

	// 添加搜索条件
	if search != "" {
		query = query.Where("username LIKE ?", "%"+search+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户总数失败",
			"data":    nil,
		})
		return
	}

	// 获取用户列表
	var users []models.User
	if err := query.Select("id, username, email, role, created_at").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户列表失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"users": users,
			"total": total,
		},
	})
}

// SetUserRole 设置用户角色
func SetUserRole(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Role string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    nil,
		})
		return
	}

	if err := config.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("role", req.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "设置用户角色失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "设置成功",
		"data":    nil,
	})
}

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// 检查是否为管理员
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
			"data":    nil,
		})
		return
	}

	if user.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "不能删除管理员用户",
			"data":    nil,
		})
		return
	}

	if err := config.DB.Delete(&models.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}
func GetJudgeStatus(c *gin.Context) {
	rdb := config.RDB
	ctx := context.Background()

	queueLen, _ := rdb.LLen(ctx, "judge:queue").Result()
	processingLen, _ := rdb.LLen(ctx, "judge:processing").Result()
	resultsLen, _ := rdb.LLen(ctx, "judge:results").Result()

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"queue_length":      queueLen,
			"processing_length": processingLen,
			"results_length":    resultsLen,
		},
	})
}

// OpenContestSubmissions 开放比赛提交记录
func OpenContestSubmissions(c *gin.Context) {
	contestId := c.Param("contestId")
	if contestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "比赛ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取比赛的提交记录ID列表
	var contestSubmission models.ContestSubmissionStatus
	if err := config.DB.Where("contest_id = ?", contestId).First(&contestSubmission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    401,
			"message": "未找到比赛提交记录",
			"data":    nil,
		})
		return
	}

	// 解析提交ID列表
	var submissionIDs []uint
	if err := json.Unmarshal([]byte(contestSubmission.SubmissionIDs), &submissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "解析提交记录失败",
			"data":    nil,
		})
		return
	}

	// 开启事务
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "开启事务失败",
			"data":    nil,
		})
		return
	}

	// 逐个更新提交记录
	for _, id := range submissionIDs {
		if err := tx.Model(&models.Submission{}).
			Where("id = ?", id).
			Update("role", "user").Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "开放提交记录失败",
				"data":    nil,
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "提交记录已开放",
		"data":    nil,
	})
}

// BatchCreateUsers 批量创建用户
func BatchCreateUsers(c *gin.Context) {
	// 请求参数结构
	var req struct {
		StartNumber int    `json:"startNumber" binding:"required"`
		Count       int    `json:"count" binding:"required,max=1000"` // 限制最大数量
		Prefix      string `json:"prefix"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    nil,
		})
		return
	}

	// 生成的用户信息列表
	type UserInfo struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	userInfos := make([]UserInfo, 0, req.Count)

	// 开启事务
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "开启事务失败",
			"data":    nil,
		})
		return
	}
	defer tx.Rollback()

	// 生成用户
	for i := 0; i < req.Count; i++ {
		number := req.StartNumber + i
		username := strconv.Itoa(number)
		if req.Prefix != "" {
			username = req.Prefix + username
		}

		// 生成随机密码
		password := generateRandomPassword(8)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "生成密码哈希失败",
				"data":    nil,
			})
			return
		}

		// 创建用户
		user := models.User{
			Username:     username,
			Email:        username + "@goj.user",
			PasswordHash: string(hashedPassword),
			Role:         "user",
		}

		if err := tx.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建用户失败",
				"data":    nil,
			})
			return
		}

		// 记录用户信息
		userInfos = append(userInfos, UserInfo{
			Username: username,
			Email:    user.Email,
			Password: password, // 记录原始密码
		})
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败",
			"data":    nil,
		})
		return
	}

	// 生成Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 创建一个工作表
	sheetName := "用户信息"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// 设置表头
	headers := []string{"用户名", "邮箱", "密码"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置表头样式
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#CCCCCC"},
			Pattern: 1,
		},
	})
	f.SetRowStyle(sheetName, 1, 1, style)

	// 写入用户数据
	for i, user := range userInfos {
		row := i + 2 // 从第2行开始写入数据
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), user.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), user.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), user.Password)
	}

	// 调整列宽
	f.SetColWidth(sheetName, "A", "C", 20)

	// 生成文件名
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("users_%s.xlsx", timestamp)
	filepath := fmt.Sprintf("temp/%s", filename)

	// 确保temp目录存在
	if err := os.MkdirAll("temp", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建临时目录失败",
			"data":    nil,
		})
		return
	}

	// 保存Excel文件
	if err := f.SaveAs(filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成Excel文件失败",
			"data":    nil,
		})
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件
	c.File(filepath)

	// 异步删除临时文件
	go func() {
		time.Sleep(time.Second * 5) // 等待5秒确保文件传输完成
		os.Remove(filepath)
	}()
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}

// GetWebsiteSettings 获取网站设置
func GetWebsiteSettings(c *gin.Context) {
	var settings models.WebsiteSetting
	result := config.DB.First(&settings)

	// 如果没有找到记录，返回默认值
	if result.Error != nil {
		settings = models.WebsiteSetting{
			Title:    "GO! Judge",
			Subtitle: "快速、智能的在线评测系统",
			About:    "GOJ是一个高性能在线评测的平台，致力于提供快速、稳定的评测服务。",
			Email:    "support@example.com",
			Github:   "https://github.com/yourusername",
			ICPLink:  "https://beian.miit.gov.cn/",
			ICP:      "",
			Feature1: `<div class="feature-icon"><span class="icon-wrapper">📚</span></div><h3>丰富的题库</h3><p>包含各种难度的编程题目，从入门到进阶</p>`,
			Feature2: `<div class="feature-icon"><span class="icon-wrapper">🚀</span></div><h3>实时评测</h3><p>快速的代码执行和结果反馈</p>`,
			Feature3: `<div class="feature-icon"><span class="icon-wrapper">👥</span></div><h3>社区讨论</h3><p>与其他同学交流学习心得</p>`,
		}
		// 创建默认记录
		config.DB.Create(&settings)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    settings,
	})
}

// UpdateWebsiteSettings 更新网站设置
func UpdateWebsiteSettings(c *gin.Context) {
	var settings models.WebsiteSetting
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    nil,
		})
		return
	}

	// 查找现有记录
	var existingSettings models.WebsiteSetting
	result := config.DB.First(&existingSettings)

	if result.Error != nil {
		// 如果不存在则创建
		if err := config.DB.Create(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建设置失败",
				"data":    nil,
			})
			return
		}
	} else {
		// 如果存在则更新
		if err := config.DB.Model(&existingSettings).Updates(settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新设置失败",
				"data":    nil,
			})
			return
		}
	}

	// 清除缓存
	config.RDB.Del(c.Request.Context(), "website:settings")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    nil,
	})
}
