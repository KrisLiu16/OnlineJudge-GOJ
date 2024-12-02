import { defineStore } from 'pinia'
import { contestsApi, type Contest, type ContestListParams } from '@/api/contests'
import type { ApiResponse } from '@/types/user'

interface ContestsState {
  contests: Contest[]
  loading: boolean
  total: number
  sortBy: string
  sortOrder: 'asc' | 'desc'
}

export const useContestsStore = defineStore('contests', {
  state: (): ContestsState => ({
    contests: [],
    loading: false,
    total: 0,
    sortBy: 'startTime',
    sortOrder: 'desc',
  }),

  actions: {
    async fetchContests(params: ContestListParams) {
      this.loading = true
      try {
        const response = await contestsApi.getContests(params)
        if (response.code === 200) {
          this.contests = response.data.contests
          this.total = response.data.total
        }
      } catch (error) {
        console.error('获取比赛列表失败:', error)
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
