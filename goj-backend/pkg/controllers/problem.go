package controllers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"goj/pkg/config"
	"goj/pkg/models"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 获取题目列表的请求参数
type ProblemListParams struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Search     string `form:"search"`
	Difficulty int    `form:"difficulty"`
	Language   string `form:"language"`
	SortBy     string `form:"sortBy,default=id"`
	SortOrder  string `form:"sortOrder,default=asc"`
}

// 添加题目请求结构
type AddProblemRequest struct {
	Title       string   `json:"title" binding:"required"`
	Content     string   `json:"content" binding:"required"`
	Difficulty  int      `json:"difficulty" binding:"required,min=1,max=5"`
	Source      string   `json:"source"`
	Tags        []string `json:"tags"`
	Role        string   `json:"role" binding:"required"`
	Languages   []string `json:"languages" binding:"required"`
	TimeLimit   int      `json:"timeLimit" binding:"required,min=100,max=10000"`
	MemoryLimit int      `json:"memoryLimit" binding:"required,min=16,max=1024"`
	UseSPJ      bool     `json:"useSPJ"`
	SPJCode     string   `json:"spjCode"`
}

// GetProblems 获取题目列表
func GetProblems(c *gin.Context) {
	var params ProblemListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 检查权限
	role, exists := c.Get("role")
	var userID uint
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
	}

	query := config.DB.Model(&models.Problem{})

	// 修改权限检查逻辑
	if !exists || role == "" || role.(string) != "admin" {
		query = query.Where("role = ?", "user")
	}

	// 搜索条件
	if params.Search != "" {
		// 构建搜索条件
		searchQuery := config.DB.Where("1 = 0") // 初始化一个永假条件

		// 搜索 ID、标题、来源和标签
		searchQuery = searchQuery.Or(
			"id LIKE ? OR title LIKE ? OR source LIKE ? OR tag LIKE ?",
			"%"+params.Search+"%",
			"%"+params.Search+"%",
			"%"+params.Search+"%",
			"%"+params.Search+"%",
		)

		// 搜索题目内容(需要读取 problem.json 文件)
		var problemIDs []string
		err := config.DB.Model(&models.Problem{}).Select("id").Find(&problemIDs).Error
		if err == nil {
			for _, pid := range problemIDs {
				// 读取题目内容文件
				problemPath := filepath.Join("data", "problems", pid, "problem.json")
				data, err := os.ReadFile(problemPath)
				if err != nil {
					continue
				}

				var fullProblem struct {
					Content string `json:"content"`
				}

				if err := json.Unmarshal(data, &fullProblem); err != nil {
					continue
				}

				// 如果内容包含搜索关键词,添加到结果中
				if strings.Contains(
					strings.ToLower(fullProblem.Content),
					strings.ToLower(params.Search),
				) {
					searchQuery = searchQuery.Or("id = ?", pid)
				}
			}
		}

		// 应用搜索条件
		query = query.Where(searchQuery)
	}

	// 难度筛选
	if params.Difficulty > 0 {
		query = query.Where("difficulty = ?", params.Difficulty)
	}

	// 语言筛选
	if params.Language != "" {
		query = query.Where("languages LIKE ?", "%"+params.Language+"%")
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取题目总数失败"})
		return
	}

	// 排序
	orderStr := ""
	switch params.SortBy {
	case "acceptedCount":
		orderStr = "accepted_count " + params.SortOrder
	case "submissionCount":
		orderStr = "submission_count " + params.SortOrder
	case "id":
		orderStr = "CAST(id AS UNSIGNED) " + params.SortOrder
	default:
		orderStr = "CAST(id AS UNSIGNED) " + params.SortOrder
	}
	query = query.Order(orderStr)

	// 分页
	offset := (params.Page - 1) * params.PageSize
	var problemsList []models.Problem
	if err := query.Limit(params.PageSize).Offset(offset).Find(&problemsList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取题目列表失败"})
		return
	}

	// 获取用户提交状态（仅对已登录用户）
	statusMap := make(map[string]string)
	if userID > 0 {
		var problemIDs []string
		for _, p := range problemsList {
			problemIDs = append(problemIDs, p.ID)
		}

		if len(problemIDs) > 0 {
			var statuses []models.UserProblemStatus
			if err := config.DB.Where("user_id = ? AND problem_id IN ?", userID, problemIDs).
				Find(&statuses).Error; err != nil {
				log.Printf("Error - Failed to query problem statuses: %v", err)
			} else {
				for _, status := range statuses {
					statusMap[status.ProblemID] = string(status.Status)
				}
			}
		}

		for _, p := range problemsList {
			if _, exists := statusMap[p.ID]; !exists {
				statusMap[p.ID] = "unattempted"
			}
		}
	}

	// 构建带状态的题目列表
	type ProblemWithStatus struct {
		models.Problem
		Status string `json:"status"`
	}

	var problemsWithStatus []ProblemWithStatus
	for _, p := range problemsList {
		status := statusMap[p.ID]
		if status == "" {
			status = "unattempted"
		}
		problemsWithStatus = append(problemsWithStatus, ProblemWithStatus{
			Problem: p,
			Status:  status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"problems": problemsWithStatus,
			"total":    total,
		},
	})
}

// 添加题目
func AddProblem(c *gin.Context) {
	var req AddProblemRequest
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
			"message": "创建题目失败",
			"data":    nil,
		})
		return
	}

	// 获取新的题目ID
	var seq struct {
		ID uint
	}

	// 创建序列表（如果不存在）
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS problem_seq (
			id INT NOT NULL AUTO_INCREMENT PRIMARY KEY
		) AUTO_INCREMENT = 10001
	`)

	// 插入新记录获取ID
	if err := tx.Exec("INSERT INTO problem_seq VALUES (NULL)").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成题目ID失败",
			"data":    nil,
		})
		return
	}

	// 获取生的ID
	if err := tx.Raw("SELECT LAST_INSERT_ID() as id").Scan(&seq).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取题目ID失败",
			"data":    nil,
		})
		return
	}

	problemID := fmt.Sprintf("%05d", seq.ID)

	// 创建题目记录
	problem := models.Problem{
		ID:          problemID,
		Title:       req.Title,
		Difficulty:  req.Difficulty,
		Role:        req.Role,
		Tag:         strings.Join(req.Tags, ","),
		Source:      req.Source,
		Languages:   strings.Join(req.Languages, ","),
		TimeLimit:   int64(req.TimeLimit),
		MemoryLimit: int64(req.MemoryLimit),
		UseSPJ:      req.UseSPJ,
	}

	if err := tx.Create(&problem).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存题目失败",
			"data":    nil,
		})
		return
	}

	// 创建题目目录
	problemDir := filepath.Join("data", "problems", problemID)
	if err := os.MkdirAll(problemDir, 0755); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建题目目录失败",
			"data":    nil,
		})
		return
	}

	// 保存完整题目信息到JSON文件
	fullProblem := struct {
		ID          string   `json:"id"`
		Title       string   `json:"title"`
		Content     string   `json:"content"`
		Tags        []string `json:"tags"`
		Languages   []string `json:"languages"`
		Source      string   `json:"source"`
		Role        string   `json:"role"`
		Difficulty  int      `json:"difficulty"`
		TimeLimit   int      `json:"timeLimit"`
		MemoryLimit int      `json:"memoryLimit"`
		UseSPJ      bool     `json:"useSPJ"`
	}{
		ID:          problemID,
		Title:       req.Title,
		Content:     req.Content,
		Tags:        req.Tags,
		Languages:   req.Languages,
		Source:      req.Source,
		Role:        req.Role,
		Difficulty:  req.Difficulty,
		UseSPJ:      req.UseSPJ,
		TimeLimit:   req.TimeLimit,
		MemoryLimit: req.MemoryLimit,
	}

	jsonData, err := json.MarshalIndent(fullProblem, "", "  ")
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "序列化题目数据失败",
			"data":    nil,
		})
		return
	}

	jsonPath := filepath.Join(problemDir, "problem.json")
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存题目文件失败",
			"data":    nil,
		})
		return
	}

	if req.UseSPJ {
		spjPath := filepath.Join(problemDir, "spj.cpp")
		if err := os.WriteFile(spjPath, []byte(req.SPJCode), 0644); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "保存SPJ代码失败",
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

	// 在成功添加题目后，清除题目列表相关的所有缓存
	if err := config.RDB.Del(c.Request.Context(), config.ProblemListKey+"*").Err(); err != nil {
		log.Printf("Failed to clear problem list cache: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建题目成功",
		"data": gin.H{
			"id": problemID,
		},
	})
}

// GetAdminProblems 获取题目管理列表
func GetAdminProblems(c *gin.Context) {
	var params struct {
		Page     int    `form:"page,default=1"`
		PageSize int    `form:"pageSize,default=20"`
		Search   string `form:"search"`
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    nil,
		})
		return
	}

	query := config.DB.Model(&models.Problem{})

	// 搜索条件
	if params.Search != "" {
		query = query.Where("title LIKE ? OR id LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取题目总数失败",
			"data":    nil,
		})
		return
	}

	// 分页查询
	var problems []models.Problem
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("CAST(id AS UNSIGNED)").
		Limit(params.PageSize).
		Offset(offset).
		Find(&problems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取题目列表失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"problems": problems,
			"total":    total,
		},
	})
}

// DeleteProblem 删除题目
func DeleteProblem(c *gin.Context) {
	problemID := c.Param("id")

	// 开启事务
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除题目失败",
			"data":    nil,
		})
		return
	}

	// 删除数据库记录
	if err := tx.Delete(&models.Problem{}, "id = ?", problemID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除题目记录失败",
			"data":    nil,
		})
		return
	}

	// 删除题目文件
	problemDir := filepath.Join("data", "problems", problemID)
	if err := os.RemoveAll(problemDir); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除题目文件失败",
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

	// 清除缓存
	if err := config.RDB.Del(c.Request.Context(), config.ProblemListKey+"*").Err(); err != nil {
		log.Printf("Failed to clear problem list cache: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}

// GetProblemDetail 获取题目详细信息
func GetProblemDetail(c *gin.Context) {
	problemID := c.Param("id")
	log.Printf("Debug - Getting problem detail for problemID: %s", problemID)

	// 获取用户角色和ID（如果已登录）
	var userID uint
	var userRole string
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
		if role, exists := c.Get("role"); exists {
			userRole = role.(string)
		}
		log.Printf("Debug - User is logged in with userID: %d, role: %s", userID, userRole)
	} else {
		log.Printf("Debug - No user logged in")
	}

	var problem models.Problem
	if err := config.DB.First(&problem, "id = ?", problemID).Error; err != nil {
		log.Printf("Error - Problem not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "题目不存在",
			"data":    nil,
		})
		return
	}
	log.Printf("Debug - Found problem: %s", problem.Title)

	// 检查权限
	if problem.Role == "admin" && userRole != "admin" {
		log.Printf("Debug - Access denied: problem requires admin role")
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "题目不存在",
			"data":    nil,
		})
		return
	}

	// 获取用户提交状态（仅对已登录用户）
	var status string = "unattempted"
	log.Printf("Debug - Initial status set to: %s", status)

	if userID > 0 {
		var problemStatus models.UserProblemStatus
		if err := config.DB.Where("user_id = ? AND problem_id = ?", userID, problemID).
			First(&problemStatus).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				log.Printf("Error - Failed to query problem status: %v", err)
			}
		} else {
			status = string(problemStatus.Status)
			log.Printf("Debug - Status found in database: %s", status)
		}
	}

	// 从文件读取完整题目信息
	problemPath := filepath.Join("data", "problems", problemID, "problem.json")
	log.Printf("Debug - Reading problem file from: %s", problemPath)

	data, err := os.ReadFile(problemPath)
	if err != nil {
		log.Printf("Error - Failed to read problem file: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "题目文件不存在",
			"data":    nil,
		})
		return
	}

	var fullProblem struct {
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}

	if err := json.Unmarshal(data, &fullProblem); err != nil {
		log.Printf("Error - Failed to unmarshal problem data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "读取题目数据失败",
			"data":    nil,
		})
		return
	}
	log.Printf("Debug - Successfully parsed problem data")

	// 组合返回数据
	response := gin.H{
		"id":              problem.ID,
		"title":           problem.Title,
		"content":         fullProblem.Content,
		"difficulty":      problem.Difficulty,
		"source":          problem.Source,
		"tags":            fullProblem.Tags,
		"role":            problem.Role,
		"languages":       strings.Split(problem.Languages, ","),
		"timeLimit":       problem.TimeLimit,
		"memoryLimit":     problem.MemoryLimit,
		"acceptedCount":   problem.AcceptedCount,
		"submissionCount": problem.SubmissionCount,
		"status":          status,
		"useSPJ":          problem.UseSPJ,
	}
	log.Printf("Debug - Final status in response: %s", status)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    response,
	})
}

// UpdateProblem 更新题目
func UpdateProblem(c *gin.Context) {
	problemID := c.Param("id")
	var req AddProblemRequest
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
			"message": "更新题目失败",
			"data":    nil,
		})
		return
	}

	// 更新数据库记录
	problem := models.Problem{
		Title:       req.Title,
		Difficulty:  req.Difficulty,
		Role:        req.Role,
		Tag:         strings.Join(req.Tags, ","),
		Source:      req.Source,
		Languages:   strings.Join(req.Languages, ","),
		TimeLimit:   int64(req.TimeLimit),
		MemoryLimit: int64(req.MemoryLimit),
		UseSPJ:      req.UseSPJ,
	}

	if err := tx.Model(&models.Problem{}).Where("id = ?", problemID).Updates(&problem).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新题目失败",
			"data":    nil,
		})
		return
	}

	// 更新JSON文件
	problemDir := filepath.Join("data", "problems", problemID)
	fullProblem := struct {
		ID          string   `json:"id"`
		Title       string   `json:"title"`
		Content     string   `json:"content"`
		Tags        []string `json:"tags"`
		Languages   []string `json:"languages"`
		Source      string   `json:"source"`
		Role        string   `json:"role"`
		Difficulty  int      `json:"difficulty"`
		TimeLimit   int      `json:"timeLimit"`
		MemoryLimit int      `json:"memoryLimit"`
		UseSPJ      bool     `json:"useSPJ"`
	}{
		ID:          problemID,
		Title:       req.Title,
		Content:     req.Content,
		Tags:        req.Tags,
		Languages:   req.Languages,
		Source:      req.Source,
		Role:        req.Role,
		Difficulty:  req.Difficulty,
		TimeLimit:   req.TimeLimit,
		MemoryLimit: req.MemoryLimit,
		UseSPJ:      req.UseSPJ,
	}

	jsonData, err := json.MarshalIndent(fullProblem, "", "  ")
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "序列化题目数据失败",
			"data":    nil,
		})
		return
	}

	jsonPath := filepath.Join(problemDir, "problem.json")
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存题目文件失败",
			"data":    nil,
		})
		return
	}

	if req.UseSPJ {
		spjPath := filepath.Join(problemDir, "spj.cpp")
		if err := os.WriteFile(spjPath, []byte(req.SPJCode), 0644); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "保存SPJ代码失败",
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

	// 清除缓存
	if err := config.RDB.Del(c.Request.Context(), config.ProblemListKey+"*").Err(); err != nil {
		log.Printf("Failed to clear problem list cache: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    nil,
	})
}

// GetProblemDataFiles 获取题目测试数据文件列表
func GetProblemDataFiles(c *gin.Context) {
	problemID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 获取目数据目录
	dataDir := filepath.Join("data", "problems", problemID, "data")
	// 如果不存在就创建目录
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建题目数据目录失败",
			})
			return
		}
		// 目录刚创建，返回空列表
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取成功",
			"data": gin.H{
				"files": []gin.H{},
				"total": 0,
			},
		})
		return
	}

	// 读取目录中的所有文件
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "读取题目数据目录失败",
		})
		return
	}

	// 过滤并获取文件信息
	var files []gin.H
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 只处理.in和.out文件
		name := entry.Name()
		if !strings.HasSuffix(name, ".in") && !strings.HasSuffix(name, ".out") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, gin.H{
			"name":         name,
			"size":         info.Size(),
			"modifiedTime": info.ModTime().Unix(),
		})
	}

	// 计算总数和分页
	total := len(files)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	// 返回分页后的结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"files": files[start:end],
			"total": total,
		},
	})
}

// UploadProblemData 上传题目测试数据
func UploadProblemData(c *gin.Context) {
	problemID := c.PostForm("problemId")
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("Failed to get upload file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "获取上传文件失败: " + err.Error(),
		})
		return
	}

	// 检查文件后缀
	if !strings.HasSuffix(file.Filename, ".in") && !strings.HasSuffix(file.Filename, ".out") {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能上传.in或.out后缀的文件",
		})
		return
	}

	// 确保目录存在
	dataDir := filepath.Join("data", "problems", problemID, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("Failed to create directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建目录失败: " + err.Error(),
		})
		return
	}

	// 保存文件
	dst := filepath.Join(dataDir, file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		log.Printf("Failed to save file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存文件失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "上传成功",
	})
}

// GetProblemDataFile 获取题目测试数据文件内容
func GetProblemDataFile(c *gin.Context) {
	problemID := c.Param("id")
	filename := c.Param("filename")

	// 读取文件内容
	filePath := filepath.Join("data", "problems", problemID, "data", filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"content": string(content),
		},
	})
}

// DownloadProblemDataFile 下载题目测试数据文件
func DownloadProblemDataFile(c *gin.Context) {
	problemID := c.Param("id")
	filename := c.Param("filename")

	filePath := filepath.Join("data", "problems", problemID, "data", filename)
	c.File(filePath)
}

// DownloadAllProblemData 下载题目所有测试数据
func DownloadAllProblemData(c *gin.Context) {
	problemID := c.Param("id")
	dataDir := filepath.Join("data", "problems", problemID, "data")

	// 创建临时zip文件
	tmpFile, err := os.CreateTemp("", "problem_data_*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建临时文件失败",
		})
		return
	}
	defer os.Remove(tmpFile.Name())

	// 创建zip写入器
	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close()

	// 遍历目录添加文件到zip
	err = filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只打包.in和.out文件
		if !strings.HasSuffix(info.Name(), ".in") && !strings.HasSuffix(info.Name(), ".out") {
			return nil
		}

		// 创建zip文件
		zipFile, err := zipWriter.Create(info.Name())
		if err != nil {
			return err
		}

		// 读取文件内容并写入zip
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = zipFile.Write(content)
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建zip文件失败",
		})
		return
	}

	// 关闭zip写入器
	zipWriter.Close()

	// 发送文件
	c.FileAttachment(tmpFile.Name(), fmt.Sprintf("problem_%s_data.zip", problemID))
}

// DeleteProblemDataFile 删除题目测试数据文件
func DeleteProblemDataFile(c *gin.Context) {
	problemID := c.Param("id")
	filename := c.Param("filename")

	filePath := filepath.Join("data", "problems", problemID, "data", filename)
	if err := os.Remove(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除文件失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// BatchDeleteProblemData 批量删除题目测试数据文件
func BatchDeleteProblemData(c *gin.Context) {
	problemID := c.Param("id")
	var req struct {
		Files []string `json:"files" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 删除文件
	for _, filename := range req.Files {
		filePath := filepath.Join("data", "problems", problemID, "data", filename)
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to delete file %s: %v", filename, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetContestProblem 获取比赛题目详情
func GetContestProblem(c *gin.Context) {
	contestID := c.Param("contestId")
	problemID := c.Param("problemId")

	// 检查比赛是否存在
	var contest models.Contest
	if err := config.DB.First(&contest, "id = ?", contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "比赛不存在",
			"data":    nil,
		})
		return
	}

	// 检查比赛状态
	now := time.Now()
	if now.Before(contest.StartTime) {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "比赛尚未开始",
			"data":    nil,
		})
		return
	}

	// 检查题目是否属于该比赛
	problemIDs := strings.Split(contest.Problems, ",")
	found := false
	for _, id := range problemIDs {
		if id == problemID {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "该题目不属于此比赛",
			"data":    nil,
		})
		return
	}

	// 从数据库获取基本信息
	var problem models.Problem
	if err := config.DB.Where("id = ?", problemID).First(&problem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "题目不存在",
			"data":    nil,
		})
		return
	}

	// 获取用户ID（如果已登录）
	var userID uint
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
	}

	// 获取用户提交状态（仅对已登录用户）
	var status string = "unattempted"
	log.Printf("Debug - Initial status set to: %s", status)

	if userID > 0 {
		var problemStatus models.UserProblemStatus
		if err := config.DB.Where("user_id = ? AND problem_id = ?", userID, problemID).
			First(&problemStatus).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				log.Printf("Error - Failed to query problem status: %v", err)
			}
		} else {
			status = string(problemStatus.Status)
			log.Printf("Debug - Status found in database: %s", status)
		}
	}

	// 从文件读取完整题目信息
	problemPath := filepath.Join("data", "problems", problemID, "problem.json")
	log.Printf("Debug - Reading problem file from: %s", problemPath)

	data, err := os.ReadFile(problemPath)
	if err != nil {
		log.Printf("Error - Failed to read problem file: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "题目文件不存在",
			"data":    nil,
		})
		return
	}

	var fullProblem struct {
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}

	if err := json.Unmarshal(data, &fullProblem); err != nil {
		log.Printf("Error - Failed to unmarshal problem data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "读取题目数据失败",
			"data":    nil,
		})
		return
	}
	log.Printf("Debug - Successfully parsed problem data")

	// 组合返回数据
	response := gin.H{
		"id":              problem.ID,
		"title":           problem.Title,
		"content":         fullProblem.Content,
		"difficulty":      problem.Difficulty,
		"source":          problem.Source,
		"tags":            fullProblem.Tags,
		"role":            problem.Role,
		"languages":       strings.Split(problem.Languages, ","),
		"timeLimit":       problem.TimeLimit,
		"memoryLimit":     problem.MemoryLimit,
		"acceptedCount":   problem.AcceptedCount,
		"submissionCount": problem.SubmissionCount,
		"status":          status,
	}
	log.Printf("Debug - Final status in response: %s", status)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    response,
	})

}

// GetContestProblems 获取比赛的所有题目
func GetContestProblems(c *gin.Context) {
	contestID := c.Param("contestId")

	// 获取用户ID（如果已登录）
	var userID uint
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
	}

	// 获取比赛信息
	var contest models.Contest
	if err := config.DB.First(&contest, "id = ?", contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "比赛不存在",
			"data":    nil,
		})
		return
	}

	// 获取所有题目
	problemIDs := strings.Split(contest.Problems, ",")
	var problems []models.Problem
	if err := config.DB.Where("id IN ?", problemIDs).Find(&problems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取题目列表失败",
			"data":    nil,
		})
		return
	}

	// 获取用户提交状态（仅对已登录用户）
	statusMap := make(map[string]string)
	if userID > 0 {
		log.Printf("Debug - Getting problem status for userID: %d", userID)

		// 获取所有题目ID
		var problemIDs []string
		for _, p := range problems {
			problemIDs = append(problemIDs, p.ID)
		}
		log.Printf("Debug - Collected problem IDs: %v", problemIDs)

		if len(problemIDs) > 0 {
			var statuses []models.UserProblemStatus
			if err := config.DB.Where("user_id = ? AND problem_id IN ?", userID, problemIDs).
				Find(&statuses).Error; err != nil {
				log.Printf("Error - Failed to query problem statuses: %v", err)
			} else {
				log.Printf("Debug - Found %d problem statuses", len(statuses))
				for _, status := range statuses {
					statusMap[status.ProblemID] = string(status.Status)
					log.Printf("Debug - Problem %s status: %s", status.ProblemID, status.Status)
				}
			}
		}

		// 未找到状态的题目设为 unattempted
		for _, p := range problems {
			if _, exists := statusMap[p.ID]; !exists {
				statusMap[p.ID] = "unattempted"
			}
		}
	}

	// 构建带状态的题目列表
	type ProblemWithStatus struct {
		models.Problem
		Status string `json:"status"`
	}

	var problemsWithStatus []ProblemWithStatus
	for _, p := range problems {
		status := statusMap[p.ID]
		if status == "" {
			status = "unattempted"
		}
		log.Printf("Debug - Final status for problem %s: %s", p.ID, status)
		problemsWithStatus = append(problemsWithStatus, ProblemWithStatus{
			Problem: p,
			Status:  status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"problems": problemsWithStatus,
			"total":    len(problems),
			"debug": gin.H{
				"userID": userID,
			},
		},
	})
}

// GetProblemSPJCode 获取题目的SPJ代码
func GetProblemSPJCode(c *gin.Context) {
	problemID := c.Param("id")

	// 检查题目是否存在且启用了SPJ
	var problem models.Problem
	if err := config.DB.First(&problem, "id = ?", problemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "题目不存在",
			"data":    nil,
		})
		return
	}

	if !problem.UseSPJ {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该题目未启用SPJ",
			"data":    nil,
		})
		return
	}

	// 读取SPJ代码文件
	spjPath := filepath.Join("data", "problems", problemID, "spj.cpp")
	spjCode, err := os.ReadFile(spjPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "SPJ代码文件不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    string(spjCode),
	})
}
