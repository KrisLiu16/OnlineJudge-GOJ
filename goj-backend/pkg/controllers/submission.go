package controllers

import (
	"encoding/json"
	"goj/pkg/config"
	"goj/pkg/judge/manager"
	"goj/pkg/judge/types"
	"goj/pkg/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"log"

	"github.com/gin-gonic/gin"
)

// 提交代码请求
type SubmitRequest struct {
	ProblemID string `json:"problemId" binding:"required"`
	Language  string `json:"language" binding:"required"`
	Code      string `json:"code" binding:"required"`
	ContestID string `json:"contestId"`
}

// CreateSubmission 处理代码提交
func CreateSubmission(c *gin.Context) {
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid submission request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    nil,
		})
		return
	}

	log.Printf("[Submission] New submission received - ProblemID: %s, Language: %s, CodeLength: %d",
		req.ProblemID, req.Language, len(req.Code))

	// 获取用户ID
	userID := c.GetUint("userID")
	log.Printf("[Submission] User ID: %d", userID)

	// 检查题目是否存在
	var problem models.Problem
	if err := config.DB.First(&problem, "id = ?", req.ProblemID).Error; err != nil {
		log.Printf("[Submission] Problem not found: %s", req.ProblemID)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "题目不存在",
			"data":    nil,
		})
		return
	}

	// 打印题目信息
	log.Printf("\033[31m[Submit] Problem limits from database - Problem: %s, Time: %d ms, Memory: %d MB\033[0m",
		problem.ID, problem.TimeLimit, problem.MemoryLimit)

	// 检查语言是否持
	if !isLanguageSupported(req.Language, problem.Languages) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的编程语言",
			"data":    nil,
		})
		return
	}

	// 添加数据库连接检查
	if err := config.DB.Raw("SELECT 1").Error; err != nil {
		log.Printf("[Submission] Database connection error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库连接错误",
			"data":    nil,
		})
		return
	}

	// 创建提交记录
	submission := models.Submission{
		UserID:     userID,
		ProblemID:  req.ProblemID,
		ContestID:  req.ContestID,
		Language:   req.Language,
		Code:       req.Code,
		Status:     types.StatusPending,
		SubmitTime: time.Now(),
	}

	// 保存提交记录
	if err := config.DB.Create(&submission).Error; err != nil {
		log.Printf("[Submission] Failed to save submission: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建提交记录失败",
			"data":    nil,
		})
		return
	}

	log.Printf("[Submission] Successfully saved submission with ID: %d", submission.ID)

	// 创建评测任务
	task := &types.JudgeTask{
		ID:          submission.ID,
		ProblemID:   req.ProblemID,
		Language:    req.Language,
		Code:        req.Code,
		UserID:      userID,
		TimeLimit:   problem.TimeLimit,
		MemoryLimit: problem.MemoryLimit,
		UseSPJ:      problem.UseSPJ,
	}

	// 打印任务信息
	log.Printf("\033[31m[Submit] Created judge task - ID: %d, Time: %d ms, Memory: %d MB\033[0m",
		task.ID, task.TimeLimit, task.MemoryLimit)

	// 发送到评测队列
	if err := manager.SendToJudgeQueue(task); err != nil {
		// 更新提交状态为系统错误
		config.DB.Model(&submission).Update("status", types.StatusSystemError)

		log.Printf("[Submission] Failed to send task to judge queue: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交失败",
			"data":    nil,
		})
		return
	}

	log.Printf("[Submission] Successfully sent task to judge queue")

	// 返回提交ID
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "提交成功",
		"data": gin.H{
			"id": submission.ID,
		},
	})
}

// 检查语言是否支持
func isLanguageSupported(lang, supportedLangs string) bool {
	langs := strings.Split(supportedLangs, ",")
	for _, l := range langs {
		if l == lang {
			return true
		}
	}
	return false
}

// 辅助函数：获取分页参数
func getPaginationParams(c *gin.Context) (page int, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 确保页码和每页数量在合理范围内
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return
}

// GetSubmissions 获取提交列表
func GetSubmissions(c *gin.Context) {
	// 只从查询参数获取所有筛选条件
	problemId := c.Query("problemId")
	username := c.Query("username")
	contestId := c.Query("contestId")
	status := c.Query("status")

	// 构建查询
	query := config.DB.Model(&models.Submission{}).
		Select("submissions.*, problems.title as problem_title, users.username, users.avatar").
		Joins("LEFT JOIN problems ON submissions.problem_id = problems.id").
		Joins("LEFT JOIN users ON submissions.user_id = users.id")

	// 添加所有筛选条件
	if problemId != "" {
		query = query.Where("submissions.problem_id = ?", problemId)
	}
	if username != "" {
		query = query.Where("users.username = ?", username)
	}
	if contestId != "" {
		query = query.Where("submissions.contest_id = ?", contestId)
	}
	if status != "" {
		query = query.Where("submissions.status = ?", status)
	}

	// 获取提交列表 - 限制最新500条记录
	var submissions []struct {
		models.Submission
		ProblemTitle string `json:"problemTitle"`
		Username     string `json:"username"`
		Avatar       string `json:"userAvatar"`
	}

	if err := query.Order("submissions.created_at DESC").
		Limit(500).
		Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取提交记录失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"submissions": submissions,
			"total":       len(submissions),
		},
	})
}

// GetSubmissionDetail 获取提交详情
func GetSubmissionDetail(c *gin.Context) {
	var submission struct {
		models.Submission
		Username     string `json:"username"`
		UserAvatar   string `json:"userAvatar"`
		ProblemTitle string `json:"problemTitle"`
	}

	// 获取用户信息
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	var userID uint
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
	}

	// 先查询提交记录
	if err := config.DB.Model(&models.Submission{}).
		Select("submissions.*, users.username, users.avatar as user_avatar, problems.title as problem_title").
		Joins("LEFT JOIN users ON submissions.user_id = users.id").
		Joins("LEFT JOIN problems ON submissions.problem_id = problems.id").
		Where("submissions.id = ?", c.Param("ID")).
		First(&submission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "提交记录不存在,或无权限查看",
		})
		return
	}

	// 权限检查：
	// 1. 管理员可以看所有提交
	// 2. 用户可以看自己的提交
	// 3. 用户可以看非管理员题目的其他用户提交
	if role != "admin" {
		if submission.Role == "admin" && submission.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无权限查看该提交",
			})
			return
		}
	}

	// 构造返回数据
	response := gin.H{
		"id":              submission.ID,
		"userId":          submission.UserID,
		"username":        submission.Username,
		"userAvatar":      submission.UserAvatar,
		"problemId":       submission.ProblemID,
		"problemTitle":    submission.ProblemTitle,
		"language":        submission.Language,
		"code":            submission.Code,
		"status":          submission.Status,
		"timeUsed":        submission.TimeUsed,
		"memoryUsed":      submission.MemoryUsed,
		"errorInfo":       submission.ErrorInfo,
		"submitTime":      submission.SubmitTime,
		"judgeTime":       submission.JudgeTime,
		"testcasesStatus": submission.TestcasesStatus, // 直接使用，因为已经是[]string类型
		"testcasesInfo":   submission.TestcasesInfo,   // 直接使用，因为已经是[]string类型
		"testCaseResults": []types.TestCaseResult{},   // 初始化为空数组
	}

	// 只需要解析详细的测试点结果
	if submission.TestCaseResults != "" {
		var results []types.TestCaseResult
		if err := json.Unmarshal([]byte(submission.TestCaseResults), &results); err == nil {
			response["testCaseResults"] = results
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": response,
	})
}
