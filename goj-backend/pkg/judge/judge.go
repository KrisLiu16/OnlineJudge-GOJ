package judge

import (
	"goj/pkg/judge/manager"
	"goj/pkg/judge/types"
)

// Submit 提交评测任务
func Submit(task *types.JudgeTask) error {
	return manager.SendToJudgeQueue(task)
}
