package judge

import (
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/judge/manager"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/judge/types"
)

// Submit 提交评测任务
func Submit(task *types.JudgeTask) error {
	return manager.SendToJudgeQueue(task)
}
