// 评测结果类型
export interface JudgeResult {
    id: number
    userId: number
    problemId: string
    status: string
    timeUsed: number
    memoryUsed: number
    errorInfo: string
    testcasesStatus: string[]
    testCasesInfo: string[]
    testCaseResults: TestCaseResult[]
}

// 测试点结果类型
export interface TestCaseResult {
    status: string
    timeUsed: number
    memoryUsed: number
    errorInfo: string
}

