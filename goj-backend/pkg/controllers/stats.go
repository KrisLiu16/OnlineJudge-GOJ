package controllers

import (
	"context"
	"goj/pkg/config"
	"goj/pkg/models"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SystemStats 系统状态统计
type SystemStats struct {
	// Go运行时信息
	GoRoutines  int    `json:"goRoutines"`  // Go协程数量
	HeapObjects uint64 `json:"heapObjects"` // 堆对象数
	HeapAlloc   uint64 `json:"heapAlloc"`   // 堆内存使用
	StackInUse  uint64 `json:"stackInUse"`  // 栈内存使用

	// Redis状态
	RedisKeyCount         int64   `json:"redisKeyCount"`         // Redis键总数
	RedisConnectedClients int64   `json:"redisConnectedClients"` // Redis连接数
	RedisHitRate          float64 `json:"redisHitRate"`          // 缓存命中率
	RedisExpiredKeys      int64   `json:"redisExpiredKeys"`      // 过期键数量

	// 应用统计
	UptimeSeconds   int64   `json:"uptimeSeconds"`   // 运行时间(秒)
	RequestsPerMin  float64 `json:"requestsPerMin"`  // 每分钟请求数
	ErrorRate       float64 `json:"errorRate"`       // 错误率
	AverageResponse float64 `json:"averageResponse"` // 平均响应时间(ms)
}

// GetSystemStats 获取系统状态
func GetSystemStats(ctx context.Context) (*SystemStats, error) {
	// 尝试从缓存获取
	cacheKey := "stats:system"
	var stats SystemStats

	err := config.GetCache(ctx, cacheKey, &stats)
	if err == nil {
		return &stats, nil
	}

	// 缓存未命中，获取新数据
	stats = SystemStats{}

	// 获取Go运行时信息
	stats.GoRoutines = runtime.NumGoroutine()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	stats.HeapObjects = memStats.HeapObjects
	stats.HeapAlloc = memStats.HeapAlloc
	stats.StackInUse = memStats.StackInuse

	// 获取Redis状态
	if redisInfo, err := config.RDB.Info(ctx, "keyspace", "clients", "stats").Result(); err == nil {
		stats.RedisKeyCount = parseRedisKeyCount(redisInfo)
		stats.RedisConnectedClients = parseRedisClients(redisInfo)
		stats.RedisHitRate = parseRedisHitRate(redisInfo)
		stats.RedisExpiredKeys = parseRedisExpiredKeys(redisInfo)
	}

	// 获取应用统计
	stats.UptimeSeconds = int64(time.Since(startTime).Seconds())
	stats.RequestsPerMin = calculateRequestsPerMinute()
	stats.ErrorRate = calculateErrorRate()
	stats.AverageResponse = calculateAverageResponse()

	// 将结果存入缓存，设置较短的过期时间（30秒）
	_ = config.SetCache(ctx, cacheKey, stats, 30*time.Second)

	return &stats, nil
}

// Redis相关辅助函数
func parseRedisKeyCount(info string) int64 {
	var total int64
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "db") {
			fields := strings.Split(line, ",")
			for _, field := range fields {
				if strings.HasPrefix(field, "keys=") {
					if count, err := strconv.ParseInt(strings.TrimPrefix(field, "keys="), 10, 64); err == nil {
						total += count
					}
				}
			}
		}
	}
	return total
}

func parseRedisClients(info string) int64 {
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "connected_clients:") {
			fields := strings.Split(line, ":")
			if len(fields) == 2 {
				if value, err := strconv.ParseInt(strings.TrimSpace(fields[1]), 10, 64); err == nil {
					return value
				}
			}
		}
	}
	return 0
}

func parseRedisHitRate(info string) float64 {
	var hits, misses int64
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "keyspace_hits:") {
			hits, _ = strconv.ParseInt(strings.TrimPrefix(line, "keyspace_hits:"), 10, 64)
		}
		if strings.HasPrefix(line, "keyspace_misses:") {
			misses, _ = strconv.ParseInt(strings.TrimPrefix(line, "keyspace_misses:"), 10, 64)
		}
	}
	if hits+misses == 0 {
		return 0
	}
	return float64(hits) * 100 / float64(hits+misses)
}

func parseRedisExpiredKeys(info string) int64 {
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "expired_keys:") {
			if value, err := strconv.ParseInt(strings.TrimPrefix(line, "expired_keys:"), 10, 64); err == nil {
				return value
			}
		}
	}
	return 0
}

// 请求统计相关
var (
	startTime    = time.Now()
	requestStats = struct {
		sync.RWMutex
		data map[string]*RequestStat
	}{
		data: make(map[string]*RequestStat),
	}
)

type RequestStat struct {
	Count      int64
	Errors     int64
	Duration   time.Duration
	LastUpdate time.Time
}

func calculateRequestsPerMinute() float64 {
	requestStats.RLock()
	defer requestStats.RUnlock()

	var total int64
	now := time.Now()
	for _, stat := range requestStats.data {
		if now.Sub(stat.LastUpdate) <= time.Minute {
			total += stat.Count
		}
	}
	return float64(total)
}

func calculateErrorRate() float64 {
	requestStats.RLock()
	defer requestStats.RUnlock()

	var totalRequests, totalErrors int64
	now := time.Now()
	for _, stat := range requestStats.data {
		if now.Sub(stat.LastUpdate) <= time.Minute {
			totalRequests += stat.Count
			totalErrors += stat.Errors
		}
	}
	if totalRequests == 0 {
		return 0
	}
	return float64(totalErrors) * 100 / float64(totalRequests)
}

func calculateAverageResponse() float64 {
	requestStats.RLock()
	defer requestStats.RUnlock()

	var totalDuration time.Duration
	var totalRequests int64
	now := time.Now()
	for _, stat := range requestStats.data {
		if now.Sub(stat.LastUpdate) <= time.Minute {
			totalDuration += stat.Duration
			totalRequests += stat.Count
		}
	}
	if totalRequests == 0 {
		return 0
	}
	return float64(totalDuration.Milliseconds()) / float64(totalRequests)
}

// GetUserCount 获取用户总数
func GetUserCount(ctx context.Context) (int64, error) {
	// 尝试从缓存获取
	var count int64
	cacheKey := "stats:user_count"

	err := config.GetCache(ctx, cacheKey, &count)
	if err == nil {
		return count, nil
	}

	// 缓存未命中，从数据库查询
	if err := config.DB.Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, err
	}

	// 设置缓存，有效期1小时
	_ = config.SetCache(ctx, cacheKey, count, time.Hour)

	return count, nil
}

// GetProblemCount 获取题目总数
func GetProblemCount(ctx context.Context) (int64, error) {
	// 尝试从缓存获取
	var count int64
	cacheKey := "stats:problem_count"

	err := config.GetCache(ctx, cacheKey, &count)
	if err == nil {
		return count, nil
	}

	// 缓存未命中，从数据库查询
	if err := config.DB.Model(&models.Problem{}).Count(&count).Error; err != nil {
		return 0, err
	}

	// 设置缓存，有效期1小时
	_ = config.SetCache(ctx, cacheKey, count, time.Hour)

	return count, nil
}

// GetTodaySubmissionCount 获取今日提交数
func GetTodaySubmissionCount(ctx context.Context) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	cacheKey := "stats:submission_count:" + today

	// 尝试从缓存获取
	err := config.GetCache(ctx, cacheKey, &count)
	if err == nil {
		return count, nil
	}

	// 缓存未命中，从数据库查询
	if err := config.DB.Model(&models.Submission{}).
		Where("DATE(created_at) = ?", today).
		Count(&count).Error; err != nil {
		return 0, err
	}

	// 设置缓存，有效期5分钟
	_ = config.SetCache(ctx, cacheKey, count, 5*time.Minute)

	return count, nil
}
