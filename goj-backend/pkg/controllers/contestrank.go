package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"goj/pkg/config"
	"goj/pkg/models"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// StringArray 用于JSON序列化和反序列化字符串数组
type StringArray = models.StringArray

// ContestRankData 比赛排名数据结构
type ContestRankData struct {
	UserID     uint              `json:"userId"`
	Username   string            `json:"username"`
	Avatar     string            `json:"avatar"`
	Bio        string            `json:"bio"`      // 添加 Bio 字段
	Problems   map[string]string `json:"problems"`   // 题目状态 map[problemId]status
	Scores     map[string]int    `json:"scores"`     // IOI模式得分 map[problemId]score
	Attempts   map[string]int    `json:"attempts"`   // 尝试次数 map[problemId]attempts
	Solved     int               `json:"solved"`     // ACM模式解题数
	Penalty    int               `json:"penalty"`    // ACM模式罚时
	TotalScore int               `json:"totalScore"` // IOI模式总分
}

// 添加计算总分的函数
func calculateTotalScore(scores map[string]int) int {
	total := 0
	for _, score := range scores {
		total += score
	}
	return total
}

// 修改缓存时间为1分钟
const (
	ContestRankCacheKey    = "contest_rank:%s:%s" // 格式: contest_rank:contestId:rankType
	ContestRankCachePeriod = 60 * time.Second     // 缓存时间改为1分钟
)

// 修改提交记录查询结构体
type ContestSubmission struct {
	ID              uint        `json:"id"`
	UserID          uint        `json:"userId"`
	Username        string      `json:"username"`
	Avatar          string      `json:"avatar"`
	Bio             string      `json:"bio"`
	ProblemID       string      `json:"problemId"`
	Status          string      `json:"status"`
	SubmitTime      time.Time   `json:"submitTime"`
	TimeUsed        int         `json:"timeUsed"`
	MemoryUsed      int         `json:"memoryUsed"`
	TestcasesStatus StringArray `json:"testcasesStatus"`
	TestcasesInfo   StringArray `json:"testcasesInfo"`
}

// GetContestRank 获取比赛排名
func GetContestRank(c *gin.Context) {
	contestID := c.Param("id")
	rankType := c.DefaultQuery("type", "acm")

	// 构造缓存键
	cacheKey := fmt.Sprintf(ContestRankCacheKey, contestID, rankType)

	// 尝试从缓存获取数据
	if cachedData, err := config.RDB.Get(c, cacheKey).Result(); err == nil {
		var rankData gin.H
		if err := json.Unmarshal([]byte(cachedData), &rankData); err == nil {
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": rankData})
			return
		}
	}

	// 如果缓存不存在或已过期，则重新计算排名
	var contest models.Contest
	if err := config.DB.First(&contest, contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "比赛不存在"})
		return
	}

	// 获取比赛的提交ID列表
	var contestStatus struct {
		SubmissionIDs json.RawMessage `gorm:"column:submission_ids"`
	}
	if err := config.DB.Table("contest_submission_status").
		Where("contest_id = ?", contestID).
		First(&contestStatus).Error; err != nil {
		// 如果没有提交记录，返回空的排名数据并缓存
		rankData := gin.H{
			"problems":  strings.Split(contest.Problems, ","),
			"ranks":     []*ContestRankData{},
			"penalty":   contest.PenaltyTime,
			"type":      rankType,
			"startTime": contest.StartTime,
		}
		
		// 缓存空数据
		if jsonData, err := json.Marshal(rankData); err == nil {
			config.RDB.Set(c, cacheKey, jsonData, ContestRankCachePeriod)
		}
		
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": rankData})
		return
	}

	// 解析提交ID列表
	var submissionIDs []uint
	if err := json.Unmarshal(contestStatus.SubmissionIDs, &submissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解析提交记录失败"})
		return
	}

	// 按顺序获取所有提交记录（包含测试点信息）
	var submissions []ContestSubmission
	if err := config.DB.Model(&models.Submission{}).
		Select(`
			submissions.*,
			users.username,
			users.avatar,
			users.bio
		`).
		Joins("LEFT JOIN users ON users.id = submissions.user_id").
		Where("submissions.id IN ?", submissionIDs).
		Order(fmt.Sprintf("FIELD(submissions.id, %s)", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(submissionIDs)), ","), "[]"))).
		Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取提交详情失败"})
		return
	}

	// 在获取提交记录后添加更详细的调试信息
	for _, sub := range submissions {
		fmt.Printf("\n=== Submission Details ===\n")
		fmt.Printf("ID: %d\n", sub.ID)
		fmt.Printf("UserID: %d\n", sub.UserID)
		fmt.Printf("ProblemID: %s\n", sub.ProblemID)
		fmt.Printf("Status: %s\n", sub.Status)
		fmt.Printf("TestcasesStatus type: %T\n", sub.TestcasesStatus)
		fmt.Printf("TestcasesStatus: %+v\n", sub.TestcasesStatus)
		fmt.Printf("TestcasesInfo: %+v\n", sub.TestcasesInfo)
		fmt.Printf("Raw TestcasesStatus length: %d\n", len(sub.TestcasesStatus))
		fmt.Printf("========================\n")
	}

	// 计算排名
	rankMap := make(map[uint]*ContestRankData)
	for _, sub := range submissions {
		userData := initUserRankData(rankMap, &sub)
		if rankType == "acm" {
			// ACM模式：使用比赛开始时间计算罚时
			handleACMSubmission(userData, &sub, contest.PenaltyTime, contest.StartTime)
		} else {
			// IOI模式：使用测试点信息计算分数
			handleIOISubmission(userData, &sub)
		}
	}

	// 获取排序后的排名列表
	ranks := getRankedList(rankMap, rankType)

	// 构造返回数据
	rankData := gin.H{
		"problems":  strings.Split(contest.Problems, ","),
		"ranks":     ranks,
		"penalty":   contest.PenaltyTime,
		"type":      rankType,
		"startTime": contest.StartTime,
	}

	// 将计算结果存入缓存
	if jsonData, err := json.Marshal(rankData); err == nil {
		config.RDB.Set(c, cacheKey, jsonData, ContestRankCachePeriod)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": rankData})
}

// 初始化用户排名数据
func initUserRankData(rankMap map[uint]*ContestRankData, sub *ContestSubmission) *ContestRankData {
	if _, exists := rankMap[sub.UserID]; !exists {
		rankMap[sub.UserID] = &ContestRankData{
			UserID:     sub.UserID,
			Username:   sub.Username,
			Avatar:     sub.Avatar,
			Bio:        sub.Bio,           // 添加 Bio 信息
			Problems:   make(map[string]string),
			Scores:     make(map[string]int),
			Attempts:   make(map[string]int),
			Solved:     0,
			Penalty:    0,
			TotalScore: 0,
		}
	}
	return rankMap[sub.UserID]
}

// 修改ACM提交处理函数，使用比赛开始时间计算罚时
func handleACMSubmission(userData *ContestRankData, sub *ContestSubmission, penaltyTime int, contestStart time.Time) {
	if userData.Problems[sub.ProblemID] != "Accepted" {
		userData.Attempts[sub.ProblemID]++
		if sub.Status == "Accepted" {
			userData.Problems[sub.ProblemID] = "Accepted"
			userData.Solved++
			// 使用比赛开始时间计算罚时（分钟）
			timePenalty := int(sub.SubmitTime.Sub(contestStart).Minutes())
			userData.Penalty += timePenalty + (userData.Attempts[sub.ProblemID]-1)*penaltyTime
		}
	}
}

// 修改IOI提交处理函数，直接查询数据库获取测试点信息
func handleIOISubmission(userData *ContestRankData, sub *ContestSubmission) {
	fmt.Printf("\n=== Processing IOI Submission ===\n")
	fmt.Printf("Processing submission for user %s (ID: %d)\n", userData.Username, userData.UserID)
	fmt.Printf("Problem ID: %s\n", sub.ProblemID)
	
	// 直接从数据库查询完整的提交信息
	var submission models.Submission
	if err := config.DB.First(&submission, sub.ID).Error; err != nil {
		fmt.Printf("Error fetching submission details: %v\n", err)
		return
	}
	
	currentScore := 0
	if len(submission.TestcasesStatus) > 0 {
		fmt.Printf("TestcasesStatus found with length: %d\n", len(submission.TestcasesStatus))
		fmt.Printf("TestcasesStatus content: %v\n", submission.TestcasesStatus)
		
		// 计算通过的测试点数量
		passedCount := 0
		for i, status := range submission.TestcasesStatus {
			fmt.Printf("Testcase %d status: %s\n", i+1, status)
			if status == "Accepted" {
				passedCount++
			}
		}
		
		// 按通过测试点数量比例计算分数
		currentScore = (passedCount * 100) / len(submission.TestcasesStatus)
		fmt.Printf("Passed testcases: %d/%d\n", passedCount, len(submission.TestcasesStatus))
		fmt.Printf("Calculated score: %d\n", currentScore)
	} else {
		fmt.Printf("No testcases status found for this submission\n")
	}

	// 检查现有分数
	fmt.Printf("Current score for problem %s: %d\n", sub.ProblemID, userData.Scores[sub.ProblemID])
	
	// 更新最高分
	if currentScore > userData.Scores[sub.ProblemID] {
		fmt.Printf("Updating score from %d to %d\n", userData.Scores[sub.ProblemID], currentScore)
		userData.Scores[sub.ProblemID] = currentScore
		userData.TotalScore = calculateTotalScore(userData.Scores)
		fmt.Printf("New total score: %d\n", userData.TotalScore)
	} else {
		fmt.Printf("Score not updated (current: %d, new: %d)\n", userData.Scores[sub.ProblemID], currentScore)
	}
	
	fmt.Printf("=== End Processing ===\n\n")
}

// 获取排序后的排名列表
func getRankedList(rankMap map[uint]*ContestRankData, rankType string) []*ContestRankData {
	fmt.Printf("\n=== Generating Rank List ===\n")
	fmt.Printf("Total participants: %d\n", len(rankMap))
	fmt.Printf("Rank type: %s\n", rankType)
	
	ranks := make([]*ContestRankData, 0, len(rankMap))
	for userID, rank := range rankMap {
		fmt.Printf("\nUser %s (ID: %d):\n", rank.Username, userID)
		fmt.Printf("Scores: %v\n", rank.Scores)
		fmt.Printf("Total Score: %d\n", rank.TotalScore)
		ranks = append(ranks, rank)
	}

	// 排序
	sort.Slice(ranks, func(i, j int) bool {
		if rankType == "acm" {
			if ranks[i].Solved != ranks[j].Solved {
				return ranks[i].Solved > ranks[j].Solved
			}
			return ranks[i].Penalty < ranks[j].Penalty
		}
		return ranks[i].TotalScore > ranks[j].TotalScore
	})

	fmt.Printf("\nFinal Ranking:\n")
	for i, rank := range ranks {
		fmt.Printf("%d. %s - Score: %d\n", i+1, rank.Username, rank.TotalScore)
	}
	fmt.Printf("=== End Ranking ===\n\n")

	return ranks
}

// GetContest 获取单个比赛详情
func GetContest(c *gin.Context) {
	contestID := c.Param("id")

	var contest models.Contest
	if err := config.DB.First(&contest, "id = ?", contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "比赛不存在",
			"data":    nil,
		})
		return
	}

	// 更新比赛状态
	now := time.Now()
	if now.Before(contest.StartTime) {
		contest.Status = "not_started"
	} else if now.After(contest.EndTime) {
		contest.Status = "ended"
	} else {
		contest.Status = "running"
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    contest,
	})
}

// UpdateContestRating 更新比赛后的用户rating
func UpdateContestRating(c *gin.Context) {
	// 验证管理员权限
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限执行此操作",
		})
		return
	}

	contestID := c.Param("id")
	rankType := c.DefaultQuery("type", "acm")

	// 1. 获取比赛信息
	var contest models.Contest
	if err := config.DB.First(&contest, contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "比赛不存在",
		})
		return
	}

	// 2. 获取比赛排名
	// 先获取提交记录
	var contestStatus struct {
		SubmissionIDs json.RawMessage `gorm:"column:submission_ids"`
	}
	if err := config.DB.Table("contest_submission_status").
		Where("contest_id = ?", contestID).
		First(&contestStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取比赛提交记录失败"})
		return
	}

	var submissionIDs []uint
	if err := json.Unmarshal(contestStatus.SubmissionIDs, &submissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解析提交记录失败"})
		return
	}

	// 获取所有提交记录
	var submissions []ContestSubmission
	if err := config.DB.Table("submissions").
		Select(`
			submissions.id,
			submissions.user_id,
			submissions.problem_id,
			submissions.status,
			submissions.submit_time,
			users.username,
			users.rating
		`).
		Joins("LEFT JOIN users ON users.id = submissions.user_id").
		Where("submissions.id IN ?", submissionIDs).
		Order(fmt.Sprintf("FIELD(submissions.id, %s)", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(submissionIDs)), ","), "[]"))).
		Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取提交详情失败"})
		return
	}

	// 3. 计算用户排名
	rankMap := make(map[uint]*ContestRankData)
	for _, sub := range submissions {
		userData := initUserRankData(rankMap, &sub)
		if rankType == "acm" {
			handleACMSubmission(userData, &sub, contest.PenaltyTime, contest.StartTime)
		} else {
			handleIOISubmission(userData, &sub)
		}
	}

	ranks := getRankedList(rankMap, rankType)

	// 4. 构造用户排名数据
	type UserRank struct {
		UserID   uint   `json:"userId"`
		Username string `json:"username"`
		Rating   int64  `json:"rating"`
		Rank     int    `json:"rank"`
		Solved   int    `json:"solved"`
	}

	var users []UserRank
	for i, rank := range ranks {
		users = append(users, UserRank{
			UserID:   rank.UserID,
			Username: rank.Username,
			Rating:   0, // 将在后面查询
			Rank:     i + 1,
			Solved:   rank.Solved,
		})
	}

	// 5. 获取用户当前rating
	for i := range users {
		var user struct {
			Rating int64
		}
		if err := config.DB.Table("users").
			Select("rating").
			Where("id = ?", users[i].UserID).
			First(&user).Error; err != nil {
			continue
		}
		users[i].Rating = user.Rating
	}

	// 6. 更新rating
	tx := config.DB.Begin()
	for i := range users {
		// 计算种子分（根据现有rating）
		seed := 1.0
		for j := range users {
			if i != j {
				seed += 1.0 / (1.0 + math.Pow(10.0, float64(users[j].Rating-users[i].Rating)/400.0))
			}
		}

		// 计算实际排名
		rank := float64(users[i].Rank)
		expectedRank := seed

		// 计算rating变化
		ratingChange := int64(math.Round((expectedRank - rank) * 400.0 / float64(len(users))))

		// 限制rating变化幅度
		if ratingChange > 400 {
			ratingChange = 400
		} else if ratingChange < -400 {
			ratingChange = -400
		}

		// 更新用户rating
		newRating := users[i].Rating + ratingChange
		if newRating < 1 {
			newRating = 1
		}

		// 记录rating变化历史
		ratingHistory := models.RatingHistory{
			UserID:    users[i].UserID,
			ContestID: contestID,
			OldRating: users[i].Rating,
			NewRating: newRating,
			Rank:      users[i].Rank,
			UpdatedAt: time.Now(),
		}

		if err := tx.Create(&ratingHistory).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "保存rating历史失败",
			})
			return
		}

		// 更新用户rating
		if err := tx.Model(&models.User{}).
			Where("id = ?", users[i].UserID).
			Update("rating", newRating).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新用户rating失败",
			})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "rating更新成功",
	})
}

// UpdateContestPenalty 更新比赛罚时
func UpdateContestPenalty(c *gin.Context) {
	// 验证管理员权限
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限执行此操作",
		})
		return
	}

	contestID := c.Param("id")
	var req struct {
		PenaltyTime int `json:"penaltyTime"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 更新罚时
	err := config.DB.Model(&models.Contest{}).
		Where("id = ?", contestID).
		Update("penalty_time", req.PenaltyTime).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新罚时失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// 添加新的导出函数
func ExportContestRank(c *gin.Context) {
	contestID := c.Param("id")
	rankType := c.DefaultQuery("type", "acm")

	// 获取比赛信息
	var contest models.Contest
	if err := config.DB.First(&contest, contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "比赛不存在"})
		return
	}

	// 获取排名数据（复用现有逻辑）
	var contestStatus struct {
		SubmissionIDs json.RawMessage `gorm:"column:submission_ids"`
	}
	if err := config.DB.Table("contest_submission_status").
		Where("contest_id = ?", contestID).
		First(&contestStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取比赛提交记录失败"})
		return
	}

	var submissionIDs []uint
	if err := json.Unmarshal(contestStatus.SubmissionIDs, &submissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解析提交记录失败"})
		return
	}

	var submissions []ContestSubmission
	if err := config.DB.Table("submissions").
		Select(`
			submissions.id,
			submissions.user_id,
			submissions.problem_id,
			submissions.status,
			submissions.submit_time,
			submissions.time_used,
			submissions.memory_used,
			submissions.testcases_status,
			submissions.testcases_info,
			users.username,
			users.avatar,
			users.bio
		`).
		Joins("LEFT JOIN users ON users.id = submissions.user_id").
		Where("submissions.id IN ?", submissionIDs).
		Order(fmt.Sprintf("FIELD(submissions.id, %s)", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(submissionIDs)), ","), "[]"))).
		Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取提交详情失败"})
		return
	}

	// 计算排名
	rankMap := make(map[uint]*ContestRankData)
	for _, sub := range submissions {
		userData := initUserRankData(rankMap, &sub)
		if rankType == "acm" {
			handleACMSubmission(userData, &sub, contest.PenaltyTime, contest.StartTime)
		} else {
			handleIOISubmission(userData, &sub)
		}
	}

	// 获取排序后的排名列表
	ranks := getRankedList(rankMap, rankType)

	// 生成Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 创建工作表
	sheetName := "比赛排名"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// 设置表头
	headers := []string{"排名", "用户名", "个人简介"}  // 添加"个人简介"列
	if rankType == "acm" {
		headers = append(headers, "解题数", "罚时")
	} else {
		headers = append(headers, "总分")
	}
	// 添加题目列
	problems := strings.Split(contest.Problems, ",")
	for _, problem := range problems {
		headers = append(headers, fmt.Sprintf("题目%s", problem))
	}

	// 写入表头
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

	// 写入排名数据
	for i, rank := range ranks {
		row := i + 2
		// 写入基本信息
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), rank.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), rank.Bio)  // 添加 Bio 信息

		if rankType == "acm" {
			// ACM模式下的列偏移量需要+1，因为多了bio列
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), rank.Solved)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("%d:%02d", rank.Penalty/60, rank.Penalty%60))
			// 写入每题状态
			for j, problemId := range problems {
				status := "-"
				if rank.Problems[problemId] == "Accepted" {
					status = fmt.Sprintf("AC(%d)", rank.Attempts[problemId])
				} else if rank.Attempts[problemId] > 0 {
					status = fmt.Sprintf("-%d", rank.Attempts[problemId])
				}
				f.SetCellValue(sheetName, fmt.Sprintf("%c%d", 'F'+j, row), status)  // 列号从F开始
			}
		} else {
			// IOI模式下的列偏移量也需要+1
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), rank.TotalScore)
			// 写入每题分数
			for j, problemId := range problems {
				score := rank.Scores[problemId]
				f.SetCellValue(sheetName, fmt.Sprintf("%c%d", 'E'+j, row), score)  // 列号从E开始
			}
		}
	}

	// 调整列宽
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}

	// 调整bio列的宽度，因为可能内容较长
	f.SetColWidth(sheetName, "C", "C", 30)  // 设置bio列宽度为30

	// 生成文件名
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("contest_%s_rank_%s.xlsx", contestID, timestamp)
	filepath := fmt.Sprintf("temp/%s", filename)

	// 确保temp目录存在
	if err := os.MkdirAll("temp", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建临时目录失败",
		})
		return
	}

	// 保存Excel文件
	if err := f.SaveAs(filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成Excel文件失败",
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

// 修改清除缓存的函数
func ClearContestRankCache(contestID string) {
	// 同时清除 ACM 和 IOI 模式的缓存
	cacheKeyACM := fmt.Sprintf(ContestRankCacheKey, contestID, "acm")
	cacheKeyIOI := fmt.Sprintf(ContestRankCacheKey, contestID, "ioi")
	
	config.RDB.Del(context.Background(), cacheKeyACM)
	config.RDB.Del(context.Background(), cacheKeyIOI)
}
