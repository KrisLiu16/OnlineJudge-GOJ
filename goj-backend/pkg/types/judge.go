package types

const (
	StatusPending             = "Pending"               // 等待评测
	StatusJudging             = "Judging"               // 评测中
	StatusAccepted            = "Accepted"              // 通过
	StatusWrongAnswer         = "Wrong Answer"          // 答案错误
	StatusTimeLimitExceeded   = "Time Limit Exceeded"   // 超时
	StatusMemoryLimitExceeded = "Memory Limit Exceeded" // 内存超限
	StatusRuntimeError        = "Runtime Error"         // 运行时错误
	StatusCompileError        = "Compile Error"         // 编译错误
	StatusSystemError         = "System Error"          // 系统错误
)
