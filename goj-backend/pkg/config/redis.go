package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	RDB    *redis.Client
	Logger *log.Logger // 用于记录Redis相关日志
)

const (
	ProblemListKey  = "problem:list"   // 题目列表的key
	CacheExpiration = 10 * time.Minute // 缓存过期时间

	// Redis 配置常量
	MaxMemoryPolicy = "allkeys-lru"    // 内存淘汰策略: LRU算法
	MaxMemoryLimit  = "512mb"          // 最大内存使用限制
	MaxKeySize      = 1024             // 键最大长度(字节)
	MaxValueSize    = 10 * 1024 * 1024 // 值最大长度(10MB)
)

// Redis初始化配置
func InitRedis() {
	// 创建Redis客户端
	RDB = redis.NewClient(&redis.Options{
		Addr: getEnvOrDefault("REDIS_ADDR", "goj-redis:6379"),

		// 连接池配置
		PoolSize:     10, // 连接池大小
		MinIdleConns: 5,  // 最小空闲连接数

		// 超时设置
		DialTimeout:  5 * time.Second, // 建立连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
		PoolTimeout:  4 * time.Second, // 从连接池获取连接的超时时间
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// 配置Redis服务器
	if err := configureRedisServer(ctx); err != nil {
		log.Printf("Warning: Failed to configure Redis server: %v", err)
	}

	// 启动定时清除缓存的协程
	go scheduleCacheCleanup()

	log.Println("Redis connected successfully")
}

// 配置Redis服务器参数
func configureRedisServer(ctx context.Context) error {
	// 设置内存限制和淘汰策略
	configs := map[string]string{
		"maxmemory":              MaxMemoryLimit,
		"maxmemory-policy":       MaxMemoryPolicy,
		"maxmemory-samples":      "5",  // LRU算法的样本数
		"notify-keyspace-events": "Ex", // 开启键过期事件通知
	}

	for param, value := range configs {
		if err := RDB.ConfigSet(ctx, param, value).Err(); err != nil {
			return fmt.Errorf("failed to set %s=%s: %v", param, value, err)
		}
	}

	return nil
}

// SetCache 增强的缓存设置函数
func SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 检查key长度
	if len(key) > MaxKeySize {
		return fmt.Errorf("key too long: max %d bytes", MaxKeySize)
	}

	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	// 检查值大小
	if len(data) > MaxValueSize {
		return fmt.Errorf("value too large: max %d bytes", MaxValueSize)
	}

	// 检查内存使用情况
	info, err := RDB.Info(ctx, "memory").Result()
	if err == nil {
		// 这里可以解析info字符串来获取内存使用情况
		// 如果内存使用过高，可以记录警告日志
		if isHighMemoryUsage(info) {
			log.Printf("Warning: Redis memory usage is high")
		}
	}

	// 设置缓存，使用Pipeline减少网络往返
	pipe := RDB.Pipeline()
	pipe.Set(ctx, key, data, expiration)

	// 可选：设置内存使用警告
	pipe.ConfigSet(ctx, "maxmemory-policy", MaxMemoryPolicy)

	_, err = pipe.Exec(ctx)
	return err
}

// GetCache 增强的缓存获取函数
func GetCache(ctx context.Context, key string, dest interface{}) error {
	// 检查key长度
	if len(key) > MaxKeySize {
		return fmt.Errorf("key too long: max %d bytes", MaxKeySize)
	}

	data, err := RDB.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	// 检查值大小
	if len(data) > MaxValueSize {
		return fmt.Errorf("value too large: max %d bytes", MaxValueSize)
	}

	return json.Unmarshal(data, dest)
}

// DeleteCache 增强的缓存删除函数
func DeleteCache(ctx context.Context, key string) error {
	return RDB.Del(ctx, key).Err()
}

// 检查Redis内存使用情况
func isHighMemoryUsage(_ string) bool {
	// 这里可以实现具体的内存使用检查逻辑
	// 例如：解析info字符串，检查used_memory_rss等指标
	// 如果内存使用超过某个阈值（比如80%）返回true
	return false
}

// 清理缓存
func CleanupCache(ctx context.Context) error {
	// 使用SCAN命令遍历所有键
	var cursor uint64
	var keys []string
	var err error

	// 收集所有键
	for {
		keys, cursor, err = RDB.Scan(ctx, cursor, "*", 0).Result()
		if err != nil {
			return fmt.Errorf("scan keys failed: %v", err)
		}

		// 如果有键，则删除它们
		if len(keys) > 0 {
			if err := RDB.Del(ctx, keys...).Err(); err != nil {
				log.Printf("Failed to delete keys: %v", err)
			} else {
				log.Printf("Successfully deleted %d keys", len(keys))
			}
		}

		// 如果cursor为0，说明遍历完成
		if cursor == 0 {
			break
		}
	}

	log.Println("Cache cleanup completed")
	return nil
}

// 定时清除只清除过期的键
func cleanupExpiredCache(ctx context.Context) error {
	iter := RDB.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// 检查键是否过期
		ttl, err := RDB.TTL(ctx, key).Result()
		if err != nil {
			continue
		}
		// 如果键已过期或接近过期，删除它
		if ttl < 0 || ttl < time.Minute {
			if err := RDB.Del(ctx, key).Err(); err != nil {
				log.Printf("Failed to delete expired key %s: %v", key, err)
			}
		}
	}
	return iter.Err()
}

// 修改定时清除函数，使用cleanupExpiredCache
func scheduleCacheCleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		if err := cleanupExpiredCache(ctx); err != nil {
			log.Printf("Failed to cleanup expired cache: %v", err)
		} else {
			log.Println("Expired cache cleanup completed successfully")
		}
	}
}

// InitRedis 初始化 Redis 客户端
