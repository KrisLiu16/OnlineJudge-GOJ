package controllers

import (
	"fmt"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/config"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 获取比赛列表的请求参数
type ContestListParams struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Search   string `form:"search"`
	Status   string `form:"status"`
}

// 添加比赛请求结构
type AddContestRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	StartTime   string   `json:"startTime" binding:"required"`
	EndTime     string   `json:"endTime" binding:"required"`
	Role        string   `json:"role" binding:"required"`
	Problems    []string `json:"problems" binding:"required"`
}

// GetContests 获取比赛列表
func GetContests(c *gin.Context) {
	var params ContestListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 检查权限并添加调试信息
	role := c.GetString("role")
	if role == "" {
		role = "user"
	}

	query := config.DB.Model(&models.Contest{})

	// 非管理员只能看到公开比赛
	if role != "admin" {
		query = query.Where("role = ?", "public")
	}

	// 搜索条件
	if params.Search != "" {
		query = query.Where("title LIKE ?", "%"+params.Search+"%")
	}

	// 状态筛选
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取比赛总数失败"})
		return
	}

	// 分页
	offset := (params.Page - 1) * params.PageSize
	var contests []models.Contest
	if err := query.Order("start_time DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&contests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取比赛列表失败"})
		return
	}

	// 更新比赛状态
	now := time.Now()
	for i := range contests {
		if now.Before(contests[i].StartTime) {
			contests[i].Status = "not_started"
		} else if now.After(contests[i].EndTime) {
			contests[i].Status = "ended"
		} else {
			contests[i].Status = "running"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"contests": contests,
			"total":    total,
			"debug": gin.H{
				"role":          role,
				"filterApplied": role != "admin",
			},
		},
	})
}

// CreateContest 创建比赛
func CreateContest(c *gin.Context) {
	var req AddContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    nil,
		})
		return
	}

	// 开启事务
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建比赛失败",
			"data":    nil,
		})
		return
	}

	// 获取新的比赛ID
	var seq struct {
		ID uint
	}

	// 创建序列表（如果不存在）
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS contest_seq (
			id INT NOT NULL AUTO_INCREMENT PRIMARY KEY
		) AUTO_INCREMENT = 10001
	`)

	// 插入新记录获取ID
	if err := tx.Exec("INSERT INTO contest_seq VALUES (NULL)").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成比赛ID失败",
			"data":    nil,
		})
		return
	}

	// 获取生成的ID
	if err := tx.Raw("SELECT LAST_INSERT_ID() as id").Scan(&seq).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取比赛ID失败",
			"data":    nil,
		})
		return
	}

	contestID := fmt.Sprintf("%05d", seq.ID)

	// 解析时间
	startTime, err := time.Parse("2006-01-02T15:04:05Z", req.StartTime)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始时间格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 将时间转换为UTC
	startTime = startTime.UTC()

	endTime, err := time.Parse("2006-01-02T15:04:05Z", req.EndTime)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 将时间转换为UTC
	endTime = endTime.UTC()

	// 创建比赛记录
	contest := models.Contest{
		ID:          contestID,
		Title:       req.Title,
		Description: req.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Role:        req.Role,
		Problems:    strings.Join(req.Problems, ","),
	}

	if err := tx.Create(&contest).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存比赛失败",
			"data":    nil,
		})
		return
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
		"message": "创建比赛成功",
		"data": gin.H{
			"id": contestID,
		},
	})
}

// UpdateContest 更新比赛
func UpdateContest(c *gin.Context) {
	contestID := c.Param("id")
	var req AddContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    nil,
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse("2006-01-02T15:04:05Z", req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始时间格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 将时间转换为UTC
	startTime = startTime.UTC()

	endTime, err := time.Parse("2006-01-02T15:04:05Z", req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 将时间转换为UTC
	endTime = endTime.UTC()

	// 更新比赛
	contest := models.Contest{
		Title:       req.Title,
		Description: req.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Role:        req.Role,
		Problems:    strings.Join(req.Problems, ","),
	}

	if err := config.DB.Model(&models.Contest{}).Where("id = ?", contestID).Updates(&contest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新比赛失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    nil,
	})
}

// DeleteContest 删除比赛
func DeleteContest(c *gin.Context) {
	contestID := c.Param("id")

	// 开启事务
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除比赛失",
			"data":    nil,
		})
		return
	}

	// 删除比赛记录
	if err := tx.Delete(&models.Contest{}, "id = ?", contestID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除比赛记录失败",
			"data":    nil,
		})
		return
	}

	// 删除比赛参与记录
	if err := tx.Delete(&models.ContestParticipant{}, "contest_id = ?", contestID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除参与者记录失败",
			"data":    nil,
		})
		return
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
		"message": "删除成功",
		"data":    nil,
	})
}
