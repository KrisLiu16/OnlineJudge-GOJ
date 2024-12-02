import { defineStore } from 'pinia'
import { ref } from 'vue'

interface RankUser {
  username: string
  avatar?: string
  solvedCount: number
  acceptance: number
  score: number
  rank?: number
  submissions: number
}

interface RankState {
  users: RankUser[]
  total: number
  currentPage: number
  pageSize: number
  loading: boolean
  lastUpdateTime: string
  sortBy: 'score' | 'solvedCount' | 'submissions'
  sortOrder: 'asc' | 'desc'
}

export const useRankStore = defineStore('rank', () => {
  const state = ref<RankState>({
    users: [],
    total: 0,
    currentPage: 1,
    pageSize: 20,
    loading: false,
    lastUpdateTime: '',
    sortBy: 'score',
    sortOrder: 'desc',
  })

  const fetchRankList = async (params: {
    page: number
    pageSize: number
    sortBy: string
    sortOrder: string
    search?: string
  }) => {
    try {
      state.value.loading = true
      // TODO: 实现API调用
      const response = await fetch(`/api/rank?${new URLSearchParams(params)}`)
      const data = await response.json()

      state.value.users = data.users
      state.value.total = data.total
      state.value.lastUpdateTime = data.lastUpdateTime
    } catch (error) {
      console.error('Failed to fetch rank list:', error)
    } finally {
      state.value.loading = false
    }
  }

  const updateSort = (sortBy: 'score' | 'solvedCount' | 'submissions') => {
    if (state.value.sortBy === sortBy) {
      state.value.sortOrder = state.value.sortOrder === 'asc' ? 'desc' : 'asc'
    } else {
      state.value.sortBy = sortBy
      state.value.sortOrder = 'desc'
    }
    fetchRankList({
      page: state.value.currentPage,
      pageSize: state.value.pageSize,
      sortBy: state.value.sortBy,
      sortOrder: state.value.sortOrder,
    })
  }

  return {
    state,
    fetchRankList,
    updateSort,
  }
})
