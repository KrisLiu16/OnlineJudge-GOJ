<template>
  <div class="contests">
    <div class="contests-header">
      <h1>比赛</h1>
      <div class="filter-section">
        <div class="search-box">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索比赛..."
            @keyup.enter="handleSearch"
          />
        </div>
        <div>
          <button @click="handleSearch" class="page-btn">搜索</button>
        </div>
      </div>
    </div>

    <div class="contest-list">
      <template v-if="!store.loading">
        <div
          class="contest-card"
          v-for="contest in store.contests"
          :key="contest.id"
          @click="openContest(contest)"
        >
          <div class="contest-status" :class="getStatusClass(contest)">
            {{ getStatusText(contest) }}
          </div>
          <h3 class="contest-title">{{ contest.title }}</h3>
          <div class="time-section">
            <div class="time-item">
              <span class="time-label">开始：</span>
              <span class="time-value">{{ new Date(contest.startTime).toLocaleString() }}</span>
            </div>
            <div class="time-item">
              <span class="time-label">结束：</span>
              <span class="time-value">{{ new Date(contest.endTime).toLocaleString() }}</span>
            </div>
          </div>
          <p class="description">
            {{
              contest.description.length > 20
                ? contest.description.substring(0, 18) + '...'
                : contest.description
            }}
          </p>
          <div class="button-group" @click.stop>
            <button
              class="join-btn"
              :class="{ disabled: isContestEnded(contest) }"
              @click.stop="joinContest(contest)"
            >
              {{ getButtonText(contest) }}
            </button>
            <router-link :to="`/contest/${contest.id}/rank`" class="rank-btn" @click.stop>
              排名
            </router-link>
            <!-- <button
              class="participants-btn"
              @click.stop="viewParticipants(contest)"
            >
              参与者
            </button> -->
          </div>
        </div>
      </template>
      <div v-else class="loading-container">
        <div class="loading-spinner"></div>
        <div class="loading-text">加载中...</div>
      </div>
    </div>

    <div class="pagination">
      <div class="pagination-info">
        共 {{ store.total }} 条记录
        <select v-model="pageSize" @change="handlePageSizeChange" class="page-size-select">
          <option v-for="size in [20, 50, 100]" :key="size" :value="size">{{ size }} 条/页</option>
        </select>
      </div>
      <div class="pagination-buttons">
        <button @click="goToPage(1)" :disabled="currentPage === 1" class="page-btn">首页</button>
        <button @click="goToPage(currentPage - 1)" :disabled="currentPage === 1" class="page-btn">
          上一页
        </button>
        <div class="page-numbers">
          <button
            v-for="pageNum in displayedPages"
            :key="pageNum"
            @click="goToPage(pageNum)"
            :class="['page-btn', { active: currentPage === pageNum }]"
          >
            {{ pageNum }}
          </button>
        </div>
        <button
          @click="goToPage(currentPage + 1)"
          :disabled="currentPage === totalPages"
          class="page-btn"
        >
          下一页
        </button>
        <button
          @click="goToPage(totalPages)"
          :disabled="currentPage === totalPages"
          class="page-btn"
        >
          末页
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useContestsStore } from '@/stores/modules/contests'
import { useUserStore } from '@/stores/modules/user'
import { useRouter } from 'vue-router'
import type { Contest } from '@/api/contests'

const store = useContestsStore()
const userStore = useUserStore()
const router = useRouter()

// 分页相关
const currentPage = ref(1)
const pageSize = ref(20)
const totalPages = computed(() => Math.ceil(store.total / pageSize.value))

// 计算显示的页码范围
const displayedPages = computed(() => {
  const delta = 2 // 当前页前后显示的页数
  const range: number[] = []
  const rangeWithDots: (number | string)[] = []
  let l: number | undefined

  for (let i = 1; i <= totalPages.value; i++) {
    if (
      i === 1 ||
      i === totalPages.value ||
      (i >= currentPage.value - delta && i <= currentPage.value + delta)
    ) {
      range.push(i)
    }
  }

  range.forEach((i) => {
    if (l) {
      if (i - l === 2) {
        rangeWithDots.push(l + 1)
      } else if (i - l !== 1) {
        rangeWithDots.push('...')
      }
    }
    rangeWithDots.push(i)
    l = i
  })

  return rangeWithDots
})

// 页面跳转
const goToPage = (page: number) => {
  currentPage.value = page
  fetchContests()
}

// 每页条数变化
const handlePageSizeChange = () => {
  currentPage.value = 1
  fetchContests()
}

// 获取比赛列表
const fetchContests = async () => {
  const token = userStore.token
  if (!token) {
    router.push('/sign-in')
    return
  }

  await store.fetchContests({
    page: currentPage.value,
    pageSize: pageSize.value,
    search: searchQuery.value,
    token: token,
  })
}

// 初始化
onMounted(() => {
  if (!userStore.token) {
    router.push('/sign-in')
    return
  }
  fetchContests()
})

// 比赛状态相关函数
const getStatusClass = (contest: Contest) => {
  switch (contest.status) {
    case 'running':
      return 'status-ongoing'
    case 'not_started':
      return 'status-upcoming'
    default:
      return 'status-ended'
  }
}

const getStatusText = (contest: Contest) => {
  switch (contest.status) {
    case 'running':
      return '进行中'
    case 'not_started':
      return '即将开始'
    default:
      return '已结束'
  }
}

const isContestEnded = (contest: Contest) => contest.status === 'ended'

const getButtonText = (contest: Contest) => {
  switch (contest.status) {
    case 'running':
      return '进入比赛'
    case 'not_started':
      return '即将开始'
    default:
      return '查看结果'
  }
}

const searchQuery = ref('')

const handleSearch = () => {
  currentPage.value = 1
  fetchContests()
}

// const viewParticipants = (contest: Contest) => {
//   // TODO: 实现查看参与者列表功能
//   router.push(`/contest/${contest.id}/participants`)
// }

// 添加跳转函数
const joinContest = (contest: Contest) => {
  if (contest.status === 'ended') {
    router.push(`/contest/${contest.id}`)
    return
  }

  if (contest.status === 'running') {
    router.push(`/contest/${contest.id}`)
  }
}

// 添加点击整个卡片的处理
const openContest = (contest: Contest) => {
  router.push(`/contest/${contest.id}`)
}
</script>

<style scoped>
.pagination {
  margin-top: 2rem;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.pagination-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  color: var(--text-light);
}

.page-size-select {
  padding: 0.3rem 0.5rem;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--text-light);
  cursor: pointer;
}

.pagination-buttons {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.page-numbers {
  display: flex;
  gap: 0.5rem;
}

.page-btn {
  padding: 0.5rem 1rem;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-light);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  background: var(--primary-color);
  color: white;
}

.page-btn.active {
  background: var(--primary-color);
  color: white;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.contests {
  min-height: calc((100vh - 64px - 150px) * 1.5);
  padding: 80px 2rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.contest-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin-top: 2rem;
}

.contest-card {
  background: rgba(255, 255, 255, 0.05);
  padding: 2rem;
  border-radius: 12px;
  position: relative;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow: hidden;
  cursor: pointer;
}

.contest-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at top right, rgba(255, 255, 255, 0.1), transparent);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.contest-card:hover {
  transform: translateY(-4px);
  border-color: var(--primary-color);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
}

.contest-card:hover::before {
  opacity: 1;
}

.contest-title {
  font-size: 1.5rem;
  margin-bottom: 1rem;
  color: var(--text-light);
  transition: color 0.3s ease;
}

.contest-card:hover .contest-title {
  color: #006996;
}

.time-section {
  margin: 1rem 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  background: rgba(0, 105, 150, 0.05);
  padding: 0.75rem;
  border-radius: 8px;
}

.time-item {
  display: flex;
  align-items: center;
}

.time-label {
  color: var(--text-light);
  font-weight: 500;
  width: 50px;
}

.time-value {
  color: var(--text-light);
}

.description {
  color: var(--text-light);
  opacity: 0.8;
  margin: 1rem 0;
  line-height: 1.5;
  padding: 0.5rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 4px;
}

.button-group {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.5rem;
  position: relative;
  z-index: 2;
}

.join-btn,
.rank-btn,
.participants-btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-weight: 500;
}

.join-btn {
  background: #006996;
  color: white;
  flex: 2;
  position: relative;
  overflow: hidden;
}

.join-btn::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 150%;
  height: 150%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.2), transparent);
  transform: translate(-50%, -50%) scale(0);
  opacity: 0;
  transition:
    transform 0.6s ease,
    opacity 0.4s ease;
}

.join-btn:hover::after {
  transform: translate(-50%, -50%) scale(1);
  opacity: 1;
}

.rank-btn,
.participants-btn {
  background: rgba(0, 105, 150, 0.1);
  color: var(--text-light);
  flex: 1;
}

.join-btn:hover:not(.disabled),
.rank-btn:hover,
.participants-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 105, 150, 0.2);
}

.join-btn.disabled {
  background: var(--text-gray);
  cursor: default;
}

.contest-status {
  position: absolute;
  top: 1rem;
  right: 1rem;
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
  letter-spacing: 0.5px;
  backdrop-filter: blur(4px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.status-ongoing {
  background: linear-gradient(135deg, #006996, #0088be);
  color: white;
}

.status-upcoming {
  background: linear-gradient(135deg, #4caf50, #45a049);
  color: white;
}

.status-ended {
  background: linear-gradient(135deg, #9e9e9e, #757575);
  color: white;
}

.contests-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.filter-section {
  display: flex;
  gap: 1rem;
}

.search-box {
  display: flex;
  gap: 0.5rem;
}

.search-box input {
  width: 300px;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-light);
  transition: all 0.3s ease;
}

.search-box input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(var(--primary-color-rgb), 0.1);
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 12px;
  margin: 2rem 0;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s linear infinite;
}

.loading-text {
  margin-top: 1rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.rank-btn {
  background: rgba(0, 105, 150, 0.1);
  color: var(--text-light);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  text-decoration: none;
  transition: all 0.3s ease;
  text-align: center;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.rank-btn:hover {
  transform: translateY(-2px);
  background: rgba(0, 105, 150, 0.2);
  box-shadow: 0 4px 12px rgba(0, 105, 150, 0.2);
}
</style>
