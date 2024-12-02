package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"goj/pkg/config"
	"goj/pkg/judge/types"
	"log"
)

const (
	JudgeQueueKey  = "judge:queue"  // Redis队列键
	ResultQueueKey = "judge:result" // 结果队列键
)

// SendToJudgeQueue 发送任务到评测队列
func SendToJudgeQueue(task *types.JudgeTask) error {
	log.Printf("\033[31m[Queue] Sending task to queue - ID: %d, Time: %d ms, Memory: %d MB\033[0m",
		task.ID, task.TimeLimit, task.MemoryLimit)

	jsonData, err := json.Marshal(task)
	if err != nil {
		// log.Printf("\033[31m[Queue] Failed to marshal task %d: %v\033[0m", task.ID, err)
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	ctx := context.Background()
	if err := config.RDB.LPush(ctx, JudgeQueueKey, jsonData).Err(); err != nil {
		// log.Printf("\033[31m[Queue] Failed to push task %d to queue: %v\033[0m", task.ID, err)
		return fmt.Errorf("failed to push task to queue: %v", err)
	}

	// log.Printf("\033[31m[Queue] Task %d pushed to queue with code length: %d\033[0m", task.ID, len(task.Code))
	return nil
}

// GetFromJudgeQueue 从评测队列获取任务
func GetFromJudgeQueue() (*types.JudgeTask, error) {
	ctx := context.Background()

	// 使用阻塞式获取
	result, err := config.RDB.BRPop(ctx, 0, JudgeQueueKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to pop from queue: %v", err)
	}

	var task types.JudgeTask
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %v", err)
	}

	// log.Printf("\033[31m[Queue] Got task from queue - ID: %d, Time: %d ms, Memory: %d MB\033[0m",
	// 	task.ID, task.TimeLimit, task.MemoryLimit)

	return &task, nil
}

// SendJudgeResult 发送评测结果
func SendJudgeResult(result *types.JudgeResult) error {
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	ctx := context.Background()
	if err := config.RDB.LPush(ctx, ResultQueueKey, jsonData).Err(); err != nil {
		return fmt.Errorf("failed to push result to queue: %v", err)
	}

	log.Printf("[Queue] Result for task %d pushed to result queue", result.ID)
	return nil
}
