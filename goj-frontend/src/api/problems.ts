import request from './config'
import type { ApiResponse } from '@/types/user'

export interface Problem {
  id: number
  title: string
  difficulty: 1 | 2 | 3 | 4 | 5
  acceptedCount: number
  submissionCount: number
  source: string
  role: 'user' | 'admin'
}

export interface ProblemListParams {
  page: number
  pageSize: number
  difficulty?: number
  search?: string
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

export interface ProblemListResponse {
  problems: Problem[]
  total: number
}

export const problemsApi = {
  getProblems: (params: ProblemListParams): Promise<ApiResponse<ProblemListResponse>> =>
    request.get('/problems', { params }),
}
