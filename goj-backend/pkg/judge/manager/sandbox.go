package manager

// SandboxCmd 沙箱命令配置
type SandboxCmd struct {
	Args          []string               `json:"args"`          // 程序命令行参数
	Env           []string               `json:"env"`           // 程序环境变量
	Files         []interface{}          `json:"files"`         // 文件配置
	CpuLimit      int64                  `json:"cpuLimit"`      // CPU时间限制(ns)
	MemoryLimit   int64                  `json:"memoryLimit"`   // 内存限制(byte)
	ProcLimit     int                    `json:"procLimit"`     // 进程数限制
	CopyIn        map[string]interface{} `json:"copyIn"`        // 输入文件
	CopyOut       []string               `json:"copyOut"`       // 输出文件
	CopyOutCached []string               `json:"copyOutCached"` // 缓存的输出文件
}

// SandboxRequest 评测请求
type SandboxRequest struct {
	Cmd []SandboxCmd `json:"cmd"` // 沙箱命令列表
}

// SandboxResponse 评测响应
type SandboxResponse struct {
	Status     string            `json:"status"`     // 运行状态
	ExitStatus int               `json:"exitStatus"` // 退出状态码
	Time       int64             `json:"time"`       // 运行时间(ns)
	Memory     int64             `json:"memory"`     // 内存使用(byte)
	Files      map[string]string `json:"files"`      // 输出文件内容
	FileIds    map[string]string `json:"fileIds"`    // 缓存文件ID
}
