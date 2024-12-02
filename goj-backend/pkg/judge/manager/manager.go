package manager

import (
	"fmt"
	"goj/pkg/config"
	"goj/pkg/judge/handler"
	"goj/pkg/judge/types"
	"log"
	"time"
)

type JudgeManager struct {
	judgeAddr     string
	concurrency   int
	ws            *handler.WebSocketManager
	resultHandler *handler.ResultHandler
	semaphore     chan struct{}
	timeout       time.Duration   // 最长执行时间
	maxRetries    int             // 最大重试次数
	retryDelays   []time.Duration // 重试间隔
}

func NewJudgeManager(judgeAddr string, concurrency int) *JudgeManager {
	ws := handler.NewWebSocketManager()
	return &JudgeManager{
		judgeAddr:     judgeAddr,
		concurrency:   concurrency,
		ws:            ws,
		resultHandler: handler.NewResultHandler(ws, judgeAddr),
		semaphore:     make(chan struct{}, concurrency),
		timeout:       600 * time.Second,                                                    // 600秒
		maxRetries:    3,                                                                    // 3次重试
		retryDelays:   []time.Duration{3 * time.Second, 10 * time.Second, 60 * time.Second}, // 重试间隔
	}
}

func (m *JudgeManager) Start() {
	log.Printf("[Manager] Starting judge manager")
	go m.processQueue()
}

func (m *JudgeManager) processQueue() {
	for {
		task, err := GetFromJudgeQueue()
		if err != nil {
			time.Sleep(time.Second) // 获取失败时等待一秒
			continue
		}

		m.semaphore <- struct{}{}

		go func(task *types.JudgeTask) {
			defer func() {
				<-m.semaphore
				if r := recover(); r != nil {
					log.Printf("[Manager] Panic recovered in processQueue: %v", r)
				}
			}()

			var result *types.JudgeResult
			var err error

			// 首次执行评测
			done := make(chan bool, 1)
			go func() {
				result, err = m.executeJudge(task)
				done <- true
			}()

			select {
			case <-done:
				// 只有在发生系统错误时才进行重试
				if err != nil && (result == nil || result.Status == types.StatusSystemError) {
					// 进行重试
					for retry := 0; retry < m.maxRetries-1; retry++ {
						delay := m.retryDelays[retry]
						log.Printf("[Manager] Retry %d for task %d after %v due to system error", retry+1, task.ID, delay)
						time.Sleep(delay)

						done := make(chan bool, 1)
						go func() {
							result, err = m.executeJudge(task)
							done <- true
						}()

						select {
						case <-done:
							if err == nil {
								break // 成功执行，退出重试循环
							}
							log.Printf("[Manager] Retry attempt %d failed for task %d: %v", retry+1, task.ID, err)
						case <-time.After(m.timeout):
							err = fmt.Errorf("judge timeout after %v", m.timeout)
							log.Printf("[Manager] Task %d timeout during retry", task.ID)
						}
					}
				}
			case <-time.After(m.timeout):
				err = fmt.Errorf("judge timeout after %v", m.timeout)
				log.Printf("[Manager] Task %d timeout", task.ID)
			}

			// 如果所有重试都失败
			if err != nil {
				result = &types.JudgeResult{
					ID:        task.ID,
					Status:    types.StatusSystemError,
					ErrorInfo: err.Error(),
				}
			}

			// 处理结果
			if err := m.resultHandler.HandleResult(result); err != nil {
				log.Printf("[Manager] Failed to handle result: %v", err)
			}

			// 发送到结果队列
			if err := SendJudgeResult(result); err != nil {
				log.Printf("[Manager] Failed to send result: %v", err)
			}
		}(task)
	}
}

// executeJudge 执行评测
func (m *JudgeManager) executeJudge(task *types.JudgeTask) (*types.JudgeResult, error) {
	// 获取语言配置
	langConfig, ok := config.Language.Languages[task.Language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", task.Language)
	}

	// 使用统一的评测策略
	strategy := &LanguageStrategy{
		judgeAddr: m.judgeAddr,
		config:    &langConfig,
	}

	result, err := strategy.Judge(task)
	if err != nil {
		log.Printf("[Manager] Judge error for task %d: %v", task.ID, err)
	} else {
		log.Printf("[Manager] Judge completed for task %d with status: %s", task.ID, result.Status)
	}

	return result, err
}
