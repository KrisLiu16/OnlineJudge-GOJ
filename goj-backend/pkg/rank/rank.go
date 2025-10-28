package rank

import (
	"context"
	"encoding/json"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/config"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RANK_CACHE_KEY = "rank:list"
	CACHE_DURATION = time.Hour
)

type RankUser struct {
	Username    string  `json:"username"`
	Avatar      string  `json:"avatar"`
	SolvedCount int64   `json:"solvedCount"`
	Acceptance  float64 `json:"acceptance"`
	Score       int64   `json:"score"`
	Rank        int     `json:"rank"`
	Submissions int64   `json:"submissions"`
}

type RankResponse struct {
	Users          []RankUser `json:"users"`
	Total          int64      `json:"total"`
	LastUpdateTime string     `json:"lastUpdateTime"`
}

// 后台定时更新排行榜缓存
func InitRankUpdateTask() {
	go func() {
		for {
			updateRankCache()
			time.Sleep(CACHE_DURATION)
		}
	}()
}

// 更新排行榜缓存
func updateRankCache() {
	ctx := context.Background()
	log.Println("Updating rank cache...")

	// 从数据库获取所有用户数据
	var users []RankUser
	err := config.DB.Table("users").
		Select("username, avatar, accepted_problems as solved_count, " +
			"CASE WHEN submissions > 0 THEN CAST(accepted_problems AS FLOAT) / submissions ELSE 0 END as acceptance, " +
			"rating as score, submissions").
		Where("deleted_at IS NULL").
		Order("rating DESC").
		Find(&users).Error

	if err != nil {
		log.Printf("Failed to fetch users from database: %v", err)
		return
	}

	// 添加排名
	for i := range users {
		users[i].Rank = i + 1
	}

	response := RankResponse{
		Users:          users,
		Total:          int64(len(users)),
		LastUpdateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 序列化数据
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal rank data: %v", err)
		return
	}

	// 存入Redis
	err = config.RDB.Set(ctx, RANK_CACHE_KEY, data, CACHE_DURATION).Err()
	if err != nil {
		log.Printf("Failed to cache rank data: %v", err)
		return
	}

	log.Println("Rank cache updated successfully")
}

// 获取排行榜数据
func GetRankList(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	sortBy := c.DefaultQuery("sortBy", "score")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	search := c.Query("search")

	// 从Redis获取缓存数据
	data, err := config.RDB.Get(ctx, RANK_CACHE_KEY).Bytes()
	if err != nil {
		log.Printf("Failed to get rank cache: %v", err)
		updateRankCache() // 如果缓存不存在，立即更新
		data, err = config.RDB.Get(ctx, RANK_CACHE_KEY).Bytes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rank data"})
			return
		}
	}

	var fullResponse RankResponse
	if err := json.Unmarshal(data, &fullResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse rank data"})
		return
	}

	// 在内存中处理排序和筛选
	users := filterAndSortUsers(fullResponse.Users, sortBy, sortOrder, search)

	// 计算分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(users) {
		start = 0
		end = 0
	}
	if end > len(users) {
		end = len(users)
	}

	response := RankResponse{
		Users:          users[start:end],
		Total:          int64(len(users)),
		LastUpdateTime: fullResponse.LastUpdateTime,
	}

	c.JSON(http.StatusOK, response)
}

// 在内存中处理排序和筛选
func filterAndSortUsers(users []RankUser, sortBy, sortOrder, search string) []RankUser {
	// 如果有搜索条件，先筛选
	if search != "" {
		filtered := make([]RankUser, 0)
		for _, user := range users {
			if strings.Contains(strings.ToLower(user.Username), strings.ToLower(search)) {
				filtered = append(filtered, user)
			}
		}
		users = filtered
	}

	// 排序
	sort.Slice(users, func(i, j int) bool {
		var result bool
		switch sortBy {
		case "solvedCount":
			result = users[i].SolvedCount > users[j].SolvedCount
		case "submissions":
			result = users[i].Submissions > users[j].Submissions
		default: // score
			result = users[i].Score > users[j].Score
		}
		return sortOrder == "desc" == result
	})

	// 更新排名
	for i := range users {
		users[i].Rank = i + 1
	}

	return users
}
