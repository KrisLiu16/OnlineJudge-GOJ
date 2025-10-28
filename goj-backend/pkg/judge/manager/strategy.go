package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/config"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/judge/types"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// sendRequest 发送请求到评测机
func sendRequest(judgeAddr string, req types.SandboxRequest) ([]types.SandboxResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(judgeAddr+"/run", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var result []types.SandboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result, nil
}

// JudgeStrategy 评测策略接口
type JudgeStrategy interface {
	Judge(task *types.JudgeTask) (*types.JudgeResult, error)
}

// LanguageStrategy 统一的语言评测策略
type LanguageStrategy struct {
	judgeAddr string
	config    *config.LangConfig
}

// Judge 实现评测接口
func (s *LanguageStrategy) Judge(task *types.JudgeTask) (*types.JudgeResult, error) {
	// 如果需要特判,提前编译SPJ
	var spjCompileResult *struct{ fileId string }
	var err error
	if task.UseSPJ {
		log.Printf("[Judge] Compiling special judge for problem %s", task.ProblemID)
		spjCompileResult, err = s.compileSpj(task.ProblemID)
		if err != nil {
			return &types.JudgeResult{
				ID:        task.ID,
				UserID:    task.UserID, 
				ProblemID: task.ProblemID,
				Status:    types.StatusSystemError,
				ErrorInfo: fmt.Sprintf("[Special Judge Compile Error] %v", err),
			}, nil
		}
	}
	// 如果需要编译
	if s.config.Compile != nil {
		// 编译代码
		compileResult, err := s.compile(task)
		if err != nil {
			return &types.JudgeResult{
				ID:        task.ID,
				UserID:    task.UserID,
				ProblemID: task.ProblemID,
				Status:    types.StatusCompileError,
				ErrorInfo: err.Error(),
			}, nil
		}
		
	
		
		// 运行测试
		return s.runTests(task, compileResult.fileId, spjCompileResult)
	}

	// 解释型语言直接运行测试
	return s.runTests(task, "", spjCompileResult)
}

// compile 编译代码
func (s *LanguageStrategy) compile(task *types.JudgeTask) (*struct{ fileId string }, error) {
	// 构造编译请求
	req := types.SandboxRequest{
		Cmd: []types.SandboxCmd{
			{
				Args: s.config.Compile.Command,
				Env:  s.config.Env,
				Files: []interface{}{
					map[string]string{"content": ""},
					map[string]interface{}{
						"name": "stdout",
						"max":  s.config.Compile.StderrMax,
					},
					map[string]interface{}{
						"name": "stderr",
						"max":  s.config.Compile.StderrMax,
					},
				},
				CpuLimit:    s.config.Compile.CPULimit,
				MemoryLimit: s.config.Compile.MemoryLimit,
				ProcLimit:   s.config.Compile.ProcLimit,
				CopyIn: map[string]interface{}{
					s.config.Filename: map[string]string{
						"content": task.Code,
					},
				},
				CopyOut:       []string{"stdout", "stderr"},
				CopyOutCached: []string{s.config.Compile.CompiledName},
			},
		},
	}

	// 发送编译请求
	resp, err := sendRequest(s.judgeAddr, req)
	if err != nil {
		return nil, err
	}

	// 检查编译结果
	if resp[0].Status != "Accepted" {
		stderr := resp[0].Files["stderr"]
		return nil, fmt.Errorf("compile error: %s", stderr)
	}

	return &struct{ fileId string }{
		fileId: resp[0].FileIds[s.config.Compile.CompiledName],
	}, nil
}

// getTestCases 获取测试用例
func getTestCases(problemID string) ([]types.TestCase, error) {
	// 构造题目数据目录路径
	dataDir := filepath.Join("data", "problems", problemID, "data")

	// 读取目录下的所有文件
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %v", err)
	}

	var testcases []types.TestCase
	// 遍历文件，查找.in和.out文件对
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".in") {
			baseName := strings.TrimSuffix(file.Name(), ".in")
			outFile := baseName + ".out"

			// 读取输入文件
			input, err := os.ReadFile(filepath.Join(dataDir, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read input file %s: %v", file.Name(), err)
			}

			// 读取输出文件
			output, err := os.ReadFile(filepath.Join(dataDir, outFile))
			if err != nil {
				return nil, fmt.Errorf("failed to read output file %s: %v", outFile, err)
			}

			testcases = append(testcases, types.TestCase{
				Name:   baseName,
				Input:  string(input),
				Output: string(output),
			})
		}
	}

	if len(testcases) == 0 {
		return nil, fmt.Errorf("no test cases found for problem %s", problemID)
	}

	return testcases, nil
}

// runTests 运行测试用例
func (s *LanguageStrategy) runTests(task *types.JudgeTask, execFileId string, spjCompileResult *struct{ fileId string }) (*types.JudgeResult, error) {
	solution := &types.JudgeResult{
		ID:         task.ID,
		UserID:     task.UserID,
		ProblemID:  task.ProblemID,
		Status:     types.StatusRunning,
		TimeUsed:   0,
		MemoryUsed: 0,
	}

	// 获取测试用例
	testcases, err := getTestCases(task.ProblemID)
	if err != nil {
		return nil, err
	}

	// 初始化测试点结果
	testCaseResults := make([]types.TestCaseResult, 0)
	testcasesStatus := make([]string, 0)
	testCasesInfo := make([]string, 0)

	// 遍历测试点，运行用户程序
	ac := 0
	notAc := 0
	maxTime := 0
	maxMemory := 0

	for i, tc := range testcases {
		memoryLimitBytes := int64(task.MemoryLimit) * 1024 * 1024 * int64(s.config.Run.LimitAmplify)
		timeLimitNanos := int64(task.TimeLimit) * 1000000 * int64(s.config.Run.LimitAmplify)

		// 构造运行命令
		cmd := types.SandboxCmd{
			Args: s.config.Run.Command,
			Env:  s.config.Env,
			Files: []interface{}{
				map[string]string{"content": tc.Input},
				map[string]interface{}{
					"name": fmt.Sprintf("stdout%d", i),
					"max":  s.config.Run.StdoutMax,
				},
				map[string]interface{}{
					"name": fmt.Sprintf("stderr%d", i),
					"max":  s.config.Run.StderrMax,
				},
			},
			CpuLimit:    timeLimitNanos,
			MemoryLimit: memoryLimitBytes,
			ProcLimit:   s.config.Run.ProcLimit,
			CopyIn:      make(map[string]interface{}),
			CopyOut:     []string{fmt.Sprintf("stdout%d", i), fmt.Sprintf("stderr%d", i)},
		}

		// 如果使用 SPJ，则需要缓存用户输出
		if task.UseSPJ {
			cmd.CopyOutCached = []string{fmt.Sprintf("stdout%d", i)}
		}

		// 根据是否有编译文件设置不同的输入
		if execFileId != "" {
			cmd.CopyIn[s.config.Compile.CompiledName] = map[string]string{
				"fileId": execFileId,
			}
		} else {
			cmd.CopyIn[s.config.Filename] = map[string]string{
				"content": task.Code,
			}
		}

		// 发送请求
		resp, err := sendRequest(s.judgeAddr, types.SandboxRequest{Cmd: []types.SandboxCmd{cmd}})
		if err != nil {
			return nil, err
		}

		// 分析运行结果
		result := resp[0]
		var status string
		var errorInfo string

		if result.Status == "Accepted" {
			log.Printf("[Judge] Program execution status: Accepted")
			log.Printf("[Judge] UseSPJ flag: %v", task.UseSPJ)

			if task.UseSPJ {
				// 检查用户输出是否存在
				userOutputKey := fmt.Sprintf("stdout%d", i)
				userOutputId, ok := result.FileIds[userOutputKey]
				if !ok {
					log.Printf("[Judge] User output not found in FileIds: %+v", result.FileIds)
					return nil, fmt.Errorf("user output not found")
				}
				log.Printf("[Judge] User output fileId: %s", userOutputId)

				log.Printf("[Judge] Using special judge for problem %s", task.ProblemID)
				// 使用特判程序
				status, errorInfo = s.specialJudge(
						task.ProblemID,
						filepath.Join("data", "problems", task.ProblemID, "data", tc.Name+".in"),
						filepath.Join("data", "problems", task.ProblemID, "data", tc.Name+".out"),
						userOutputId,
						spjCompileResult,
					)
				log.Printf("[Judge] Special judge result: status=%s, error=%s", status, errorInfo)
			} else {
				// 普通文本比对
				userOutput, ok := result.Files[fmt.Sprintf("stdout%d", i)]
				if !ok {
					log.Printf("[Judge] User output not found in Files: %+v", result.Files)
					return nil, fmt.Errorf("user output not found")
				}
				log.Printf("[Judge] Using normal text comparison")
				status, errorInfo = s.diffJudge(tc.Output, userOutput)
			}
		} else {
			log.Printf("[Judge] Program execution failed with status: %s", result.Status)
			status = mapSandboxStatus(result.Status)
			errorInfo = fmt.Sprintf("[%s]\n%s\n", result.Status, result.Files[fmt.Sprintf("stderr%d", i)])
		}

		// 更新测试点结果
		timeUsed := int(result.Time / 1000000)  // ns to ms
		memoryUsed := int(result.Memory / 1024) // bytes to KB

		testCaseResults = append(testCaseResults, types.TestCaseResult{
			Status:     status,
			TimeUsed:   timeUsed,
			MemoryUsed: memoryUsed,
			ErrorInfo:  errorInfo,
		})
		testcasesStatus = append(testcasesStatus, status)
		testCasesInfo = append(testCasesInfo, fmt.Sprintf("Time: %dms Memory: %dKB", timeUsed, memoryUsed))

		// 更新统计信息
		if status == types.StatusAccepted {
			ac++
		} else {
			notAc++
			if notAc == 1 { // 首个错误作为整体结果
				solution.Status = status
				solution.ErrorInfo = fmt.Sprintf("[Test #%d]\n%s", i+1, errorInfo)
			}
		}

		maxTime = max(maxTime, timeUsed)
		maxMemory = max(maxMemory, memoryUsed)
	}

	// 更新最终结果
	if ac == 0 && notAc == 0 {
		solution.Status = types.StatusSystemError
		solution.ErrorInfo = "No test data available"
	} else {
		solution.TimeUsed = maxTime
		solution.MemoryUsed = maxMemory
		solution.TestcasesStatus = testcasesStatus
		solution.TestCasesInfo = testCasesInfo
		solution.TestCaseResults = testCaseResults
		if notAc == 0 { // 全部通过
			solution.Status = types.StatusAccepted
		}
	}

	return solution, nil
}

// specialJudge 特判程序评测
func (s *LanguageStrategy) specialJudge(problemID, stdInPath, stdOutPath, userOutFileId string, spjCompileResult *struct{ fileId string }) (string, string) {
	log.Printf("[Judge] SPJ paths: input=%s, output=%s", stdInPath, stdOutPath)
	log.Printf("[Judge] SPJ compile result: %+v", spjCompileResult)

	// 读取输入和答案文件内容
	stdIn, err := os.ReadFile(stdInPath)
	if err != nil {
		log.Printf("[Judge] Failed to read input file: %v", err)
		return types.StatusSystemError, fmt.Sprintf("Failed to read input file: %v", err)
	}

	stdOut, err := os.ReadFile(stdOutPath)
	if err != nil {
		log.Printf("[Judge] Failed to read answer file: %v", err)
		return types.StatusSystemError, fmt.Sprintf("Failed to read answer file: %v", err)
	}

	// 构造运行请求 - 硬编码 SPJ 运行配置
	req := types.SandboxRequest{
		Cmd: []types.SandboxCmd{
			{
				Args: []string{"./spj", "std.in", "std.out", "user.out"},
				Env:  []string{"PATH=/usr/bin:/bin"},
				Files: []interface{}{
					map[string]string{"content": ""},
					map[string]interface{}{
						"name": "stdout",
						"max":  10240,
					},
					map[string]interface{}{
						"name": "stderr",
						"max":  10240,
					},
				},
				CpuLimit:    10000000000,  // 10s
				MemoryLimit: 512 << 20,    // 512MB
				ProcLimit:   50,
				CopyIn: map[string]interface{}{
					"spj": map[string]string{
						"fileId": spjCompileResult.fileId,
					},
					"std.in": map[string]string{
						"content": string(stdIn),
					},
					"std.out": map[string]string{
						"content": string(stdOut),
					},
					"user.out": map[string]string{
						"fileId": userOutFileId,
					},
				},
				CopyOut: []string{"stdout", "stderr"},
			},
		},
	}

	log.Printf("[Judge] SPJ compile command: %v", req.Cmd[0].Args)
	log.Printf("[Judge] SPJ files: %+v", req.Cmd[0].CopyIn)

	// 发送请求
	resp, err := sendRequest(s.judgeAddr, req)
	if err != nil {
		return types.StatusSystemError, fmt.Sprintf("Failed to run SPJ: %v", err)
	}


	// 检查特判结果
	switch resp[0].ExitStatus {
	case 0:
		return types.StatusAccepted, ""
	default:
		return types.StatusWrongAnswer, ""
	}
}

// diffJudge 文本对比
func (s *LanguageStrategy) diffJudge(stdOut, userOut string) (string, string) {
	// 按分割
	stdLines := strings.Split(strings.TrimSpace(stdOut), "\n")
	userLines := strings.Split(strings.TrimSpace(userOut), "\n")

	// 检查行数
	if len(stdLines) != len(userLines) {
		return types.StatusWrongAnswer, ""
	}

	// 逐行检查
	for i := 0; i < len(stdLines); i++ {
		userLine := strings.TrimRight(userLines[i], "\r\n")
		answerLine := strings.TrimRight(stdLines[i], "\r\n")

		if userLine != answerLine {
			if strings.TrimSpace(userLine) != strings.TrimSpace(answerLine) {
				return types.StatusWrongAnswer, ""
			}
			// 内容相同但格式不同（空白字符不同）
			return types.StatusPresentationError, ""
		}
	}

	return types.StatusAccepted, ""
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(Too long to display)"
}

// mapSandboxStatus 映射沙箱状态到评测状态
func mapSandboxStatus(status string) string {
	switch status {
	case "Accepted":
		return types.StatusAccepted
	case "Memory Limit Exceeded":
		return types.StatusMemoryLimitExceeded
	case "Time Limit Exceeded":
		return types.StatusTimeLimitExceeded
	case "Output Limit Exceeded":
		return types.StatusOutputLimitExceeded
	case "Runtime Error":
		return types.StatusRuntimeError
	case "File Error":
		return types.StatusFileError
	case "Nonzero Exit Status":
		return types.StatusNonzeroExit
	case "Signalled":
		return types.StatusSignalled
	case "Internal Error":
		return types.StatusInternalError
	default:
		return types.StatusSystemError
	}
}

// compileSpj 函数用于编译特判程序
func (s *LanguageStrategy) compileSpj(problemID string) (*struct{ fileId string }, error) {
	// 读取特判源码
	spjCode, err := os.ReadFile(filepath.Join("data", "problems", problemID, "spj.cpp"))
	if err != nil {
		return nil, fmt.Errorf("failed to read SPJ code: %v", err)
	}

	// 构造编译请求 - 硬编码 SPJ 编译配置
	req := types.SandboxRequest{
		Cmd: []types.SandboxCmd{
			{
				Args: []string{
					"/usr/bin/g++",
					"spj.cpp",
					"-o", "spj",
					"-O2",
					"-std=c++17",
					"-I/usr/local/include",
				},
				Env: []string{"PATH=/usr/bin:/bin"},
				Files: []interface{}{
					map[string]string{"content": ""},
					map[string]interface{}{
						"name": "stdout",
						"max":  10240,
					},
					map[string]interface{}{
						"name": "stderr",
						"max":  10240,
					},
				},
				CpuLimit:    30000000000,  // 30s
				MemoryLimit: 512 << 20,    // 512MB
				ProcLimit:   50,
				CopyIn: map[string]interface{}{
					"spj.cpp": map[string]string{
						"content": string(spjCode),
					},
				},
				CopyOut:       []string{"stdout", "stderr"},
				CopyOutCached: []string{"spj"},
			},
		},
	}

	resp, err := sendRequest(s.judgeAddr, req)
	if err != nil {
		return nil, err
	}

	if resp[0].Status != "Accepted" {
		return nil, fmt.Errorf("compile error: %s", resp[0].Files["stderr"])
	}

	return &struct{ fileId string }{
		fileId: resp[0].FileIds["spj"],
	}, nil
}
