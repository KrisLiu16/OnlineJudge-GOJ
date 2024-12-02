import request from './config'
import type { ApiResponse } from '@/types/user'

export interface Contest {
  id: string
  title: string
  description: string
  startTime: string
  endTime: string
  role: string
  status: string
  participantCount: number
  problems: string
}

export interface ContestListParams {
  page?: number
  pageSize?: number
  search?: string
  status?: string
  token?: string
}

export interface ContestListResponse {
  contests: Contest[]
  total: number
}

export const contestsApi = {
  getContests: (params: ContestListParams): Promise<ApiResponse<ContestListResponse>> =>
    request.get('/contests', { params }),
}

export const getContestStatus = (startTime: string, endTime: string): string => {
  const now = new Date()
  const start = new Date(startTime)
  const end = new Date(endTime)

  if (now < start) {
    return 'not_started'
  } else if (now > end) {
    return 'ended'
  } else {
    return 'running'
  }
}
