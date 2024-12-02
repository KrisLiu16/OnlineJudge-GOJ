package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// LanguageConfig 语言配置结构体
type LanguageConfig struct {
	Defaults  DefaultConfig         `yaml:"defaults"`  // 默认配置
	Languages map[string]LangConfig `yaml:"languages"` // 语言名到配置的映射
}

// DefaultConfig 默认配置结构体
type DefaultConfig struct {
	Env     []string  `yaml:"env"`     // 默认环境变量
	Compile CmdConfig `yaml:"compile"` // 默认编译配置
	Run     CmdConfig `yaml:"run"`     // 默认运行配置
}

// LangConfig 单个语言的配置
type LangConfig struct {
	Name     string     `yaml:"name"`     // 语言名称
	Filename string     `yaml:"filename"` // 源代码文件名
	Env      []string   `yaml:"env"`      // 环境变量
	Compile  *CmdConfig `yaml:"compile"`  // 编译配置,解释型语言为nil
	Run      CmdConfig  `yaml:"run"`      // 运行配置
}

// CmdConfig 命令配置结构体
type CmdConfig struct {
	Command      []string `yaml:"command"`       // 命令及参数
	CompiledName string   `yaml:"compiled_name"` // 编译后的文件名
	CPULimit     int64    `yaml:"cpu_limit"`     // CPU时间限制(ns)
	MemoryLimit  int64    `yaml:"memory_limit"`  // 内存限制(bytes)
	ProcLimit    int      `yaml:"proc_limit"`    // 进程数限制
	StdoutMax    int64    `yaml:"stdout_max"`    // 标准输出限制
	StderrMax    int64    `yaml:"stderr_max"`    // 标准错误限制
	StackLimit   int64    `yaml:"stack_limit"`   // 栈空间限制
	LimitAmplify int      `yaml:"limit_amplify"` // 时间和内存限制的放大倍数
}

var Language LanguageConfig

// InitLanguageConfig 初始化语言配置
func InitLanguageConfig() error {
	// 获取配置文件路径
	configPath := filepath.Join("pkg", "judge", "config", "language.yaml")
	if os.Getenv("JUDGE_LANGUAGE_CONFIG") != "" {
		configPath = os.Getenv("JUDGE_LANGUAGE_CONFIG")
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// 解析配置
	return yaml.Unmarshal(data, &Language)
}
