import { defineStore } from 'pinia'
import { problemsApi } from '@/api/problems'

// 修改 Problem 接口定义，确保与后端返回的数据结构一致
export interface Problem {
  id: number
  title: string
  difficulty: 1 | 2 | 3 | 4 | 5
  acceptedCount: number
  submissionCount: number
  source: string
  role: 'user' | 'admin'
  status: string // 添加 status 字段，表示当前用户对该题目的解题状态
}

// 其他接口定义保持不变
export interface ProblemListParams {
  page: number
  pageSize: number
  difficulty?: number
  search?: string
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

// 添加后端返回的问题数据类型
interface RawProblem extends Omit<Problem, 'status'> {
  status?: string
}

interface ProblemsState {
  problems: Problem[]
  loading: boolean
  total: number
  sortBy: string
  sortOrder: 'asc' | 'desc'
}

export const useProblemsStore = defineStore('problems', {
  state: (): ProblemsState => ({
    problems: [],
    loading: false,
    total: 0,
    sortBy: 'id',
    sortOrder: 'asc',
  }),

  actions: {
    async fetchProblems(params: ProblemListParams) {
      this.loading = true
      try {
        const response = await problemsApi.getProblems(params)
        // 确保后端返回的数据包含 status 字段
        this.problems = response.data.problems.map((problem: RawProblem) => ({
          ...problem,
          status: problem.status || 'attempted', // 提供默认值
        }))
        this.total = response.data.total
      } catch (error) {
        console.error('获取题目列表失败:', error)
      } finally {
        this.loading = false
      }
    },

    updateSort(field: string) {
      if (this.sortBy === field) {
        this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc'
      } else {
        this.sortBy = field
        this.sortOrder = 'desc'
      }
    },
  },
})
