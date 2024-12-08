package types

// JudgeStatus 评测状态常量
const (
	StatusPending             = "Pending"
	StatusCompiling           = "Compiling"
	StatusRunning             = "Running"
	StatusAccepted            = "Accepted"
	StatusWrongAnswer         = "Wrong Answer"
	StatusTimeLimitExceeded   = "Time Limit Exceeded"
	StatusMemoryLimitExceeded = "Memory Limit Exceeded"
	StatusRuntimeError        = "Runtime Error"
	StatusCompileError        = "Compile Error"
	StatusSystemError         = "System Error"
	StatusOutputLimitExceeded = "Output Limit Exceeded"
	StatusFileError           = "File Error"
	StatusNonzeroExit         = "Nonzero Exit Status"
	StatusSignalled           = "Signalled"
	StatusInternalError       = "Internal Error"
	StatusPresentationError   = "Presentation Error"
)

// JudgeConfig 评测配置 可能 没用到 但是不敢删
type JudgeConfig struct {
	TimeLimit   int  `json:"timeLimit"`   // 时间限制(ms)
	MemoryLimit int  `json:"memoryLimit"` // 内存限制(MB)
	UseSPJ      bool `json:"useSPJ"`      // 是否使用特殊评测
}

// TestCaseResult 单个测试点的结果
type TestCaseResult struct {
	Status     string `json:"status"`     // 状态
	TimeUsed   int    `json:"timeUsed"`   // 运行时间(ms)
	MemoryUsed int    `json:"memoryUsed"` // 内存使用(KB)
	ErrorInfo  string `json:"errorInfo"`  // 错误信息
}

// JudgeResult 评测结果
type JudgeResult struct {
	ID              uint             `json:"id"`
	UserID          uint             `json:"userId"`
	ProblemID       string           `json:"problemId"`
	Status          string           `json:"status"`
	TimeUsed        int              `json:"timeUsed"`   // 最大运行时间
	MemoryUsed      int              `json:"memoryUsed"` // 最大内存使用
	ErrorInfo       string           `json:"errorInfo"`
	TestcasesStatus []string         `json:"testcasesStatus"` // 兼容旧版
	TestCasesInfo   []string         `json:"testCasesInfo"`   // 兼容旧版
	TestCaseResults []TestCaseResult `json:"testCaseResults"` // 新增：详细的测试点结果
}

// JudgeTask 评测任务
type JudgeTask struct {
	ID          uint        // 提交ID
	ProblemID   string      // 题目ID
	ContestID   string      // 比赛ID
	UserID      uint        // 用户ID
	Language    string      // 编程语言
	Code        string      // 源代码
	TimeLimit   int64       // 时间限制(ms)
	MemoryLimit int64       // 内存限制(MB)
	Config      JudgeConfig // 评测配置
	UseSPJ      bool        // 是否使用特殊评测
}

// TestCase 测试用例
type TestCase struct {
	Name   string // 测试用例名称
	Input  string // 输入数据
	Output string // 期望输出
}
