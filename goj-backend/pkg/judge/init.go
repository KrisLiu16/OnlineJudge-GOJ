package judge

import (
	"goj/pkg/config"
	"goj/pkg/judge/manager"
)

// Init 初始化评测系统
func Init() error {
	// 创建评测管理器
	judgeManager := manager.NewJudgeManager(
		config.Judge.JudgeAddr,
		config.Judge.Concurrency,
	)

	// 启动评测管理器
	judgeManager.Start()

	return nil
}
