package config

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"syscall"
)

// JudgeConfig 评测配置
type JudgeConfig struct {
	JudgeAddr     string // 评测机地址
	Concurrency   int    // 评测并发数
	MemoryLimitMB int    // 每个评测任务的内存限制(MB)
}

var Judge JudgeConfig

// 获取系统内存大小(GB)
func getSystemMemory() float64 {
	var si syscall.Sysinfo_t
	err := syscall.Sysinfo(&si)
	if err != nil {
		log.Printf("[Config] Failed to get system info: %v, using default memory size", err)
		return 2
	}

	totalRAM := float64(si.Totalram) * float64(si.Unit) / (1024 * 1024 * 1024)
	return totalRAM
}

// 计算最优并发数
func calculateConcurrency(memoryLimitMB int) int {
	// 获取系统总内存(GB)
	totalMemoryGB := getSystemMemory()

	// 获取CPU核心数
	cpuCores := runtime.NumCPU()

	// 从环境变量获取预留配置
	reserveMemoryGB := 0.0
	if reserve := os.Getenv("JUDGE_RESERVE_MEMORY"); reserve != "" {
		if val, err := strconv.ParseFloat(reserve, 64); err == nil {
			reserveMemoryGB = val
		}
	}

	reserveCPU := 0
	if reserve := os.Getenv("JUDGE_RESERVE_CPU"); reserve != "" {
		if val, err := strconv.Atoi(reserve); err == nil {
			reserveCPU = val
		}
	}

	// 计算可用资源
	availableMemoryGB := totalMemoryGB - reserveMemoryGB
	availableCPU := cpuCores - reserveCPU

	// 根据内存计算可能的并发数
	memoryConcurrency := int(availableMemoryGB * 1024 / float64(memoryLimitMB))

	// 取内存和CPU核心数的较小值作为并发数
	concurrency := min(memoryConcurrency, availableCPU)

	// 确保至少有1个并发
	if concurrency < 1 {
		concurrency = 1
	}

	log.Printf("[Config] System info: Memory: %.2fGB (reserved: %.2fGB), CPU cores: %d (reserved: %d)",
		totalMemoryGB, reserveMemoryGB, cpuCores, reserveCPU)
	log.Printf("[Config] Calculated concurrency: %d (memory based: %d, CPU based: %d)",
		concurrency, memoryConcurrency, availableCPU)

	return concurrency
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// InitJudgeConfig 初始化评测配置
func InitJudgeConfig() {
	// 从环境变量获取评测机地址
	Judge.JudgeAddr = os.Getenv("JUDGE_ADDR")
	if Judge.JudgeAddr == "" {
		Judge.JudgeAddr = "http://goj-judge:5050"
	}

	// 从环境变量获取每个评测任务的内存限制(MB)
	Judge.MemoryLimitMB = 1024 // 默认1GB
	if limit := os.Getenv("JUDGE_TASK_MEMORY_LIMIT"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			Judge.MemoryLimitMB = val
		}
	}

	// 计算最优并发数
	Judge.Concurrency = calculateConcurrency(Judge.MemoryLimitMB)
}
