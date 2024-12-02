<template>
  <div class="rank">
    <div class="rank-header">
      <h1 class="glowing-text">排行榜</h1>
      <div class="rank-info">
        <span>最后更新时间: {{ formatTime(rankStore.state.lastUpdateTime) }}</span>
      </div>
    </div>

    <div class="rank-list">
      <table class="rank-table">
        <thead>
          <tr>
            <th>排名</th>
            <th>用户</th>
            <th @click="rankStore.updateSort('solvedCount')" class="sortable">
              解题数
              <span class="sort-icon" v-if="rankStore.state.sortBy === 'solvedCount'">
                {{ rankStore.state.sortOrder === 'desc' ? '↓' : '↑' }}
              </span>
            </th>
            <th @click="rankStore.updateSort('submissions')" class="sortable">
              提交数
              <span class="sort-icon" v-if="rankStore.state.sortBy === 'submissions'">
                {{ rankStore.state.sortOrder === 'desc' ? '↓' : '↑' }}
              </span>
            </th>
            <th @click="rankStore.updateSort('score')" class="sortable">
              积分
              <span class="sort-icon" v-if="rankStore.state.sortBy === 'score'">
                {{ rankStore.state.sortOrder === 'desc' ? '↓' : '↑' }}
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <template v-if="!rankStore.state.loading">
            <tr
              v-for="(user, index) in rankStore.state.users"
              :key="user.username"
              :class="{ 'top-three': index < 3 }"
            >
              <td>{{ user.rank }}</td>
              <td>
                <div class="user-info">
                  <img
                    :src="user.avatar || '/images/avatars/default-avatar.png'"
                    :alt="user.username"
                    class="user-avatar"
                  />
                  <router-link :to="`/profile/${user.username}`" class="username-link">
                    {{ user.username }}
                  </router-link>
                </div>
              </td>
              <td>{{ user.solvedCount }}</td>
              <td>{{ user.submissions }}</td>
              <td class="score">{{ user.score }}</td>
            </tr>
          </template>
          <tr v-else>
            <td colspan="5" class="loading">
              <div class="loading-spinner"></div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagination">
      <div class="pagination-info">
        共 {{ rankStore.state.total }} 条记录
        <select v-model="pageSize" @change="handleSizeChange" class="page-size-select">
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
import { ref, computed, onMounted, watch } from 'vue'
import { useRankStore } from '@/stores/modules/rank'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const rankStore = useRankStore()
const pageSize = ref(20)
const searchQuery = ref('')

// 从路由参数获取当前页码
const currentPage = computed(() => {
  const page = Number(route.params.page) || 1
  return page > 0 ? page : 1
})

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(rankStore.state.total / pageSize.value)
})

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

const formatTime = (time: string) => {
  return new Date(time).toLocaleString()
}

const goToPage = (page: number) => {
  if (page === currentPage.value) return
  router.push(`/rank/${page}`)
}

const handleSizeChange = () => {
  goToPage(1) // 改变每页条数时回到第一页
}

const fetchData = () => {
  rankStore.fetchRankList({
    page: currentPage.value,
    pageSize: pageSize.value,
    sortBy: rankStore.state.sortBy,
    sortOrder: rankStore.state.sortOrder,
    search: searchQuery.value,
  })
}

// const debounceSearch = debounce(() => {
//   goToPage(1)
// }, 300)

// 监听路由参数变化
watch(
  () => [currentPage.value, pageSize.value],
  () => {
    fetchData()
  },
)

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.rank {
  padding: 80px 2rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.rank-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.rank-info {
  display: flex;
  gap: 2rem;
  align-items: center;
}

.search-box input {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-light);
}

.rank-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0 8px;
  background: transparent;
  margin: 1rem 0;
}

.rank-table th {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  padding: 1rem 1.2rem;
  font-weight: 600;
  text-align: left;
  color: var(--text-primary);
  border: none;
  white-space: nowrap;
  border-bottom: none;
}

.rank-table th:first-child {
  border-top-left-radius: 8px;
  border-bottom-left-radius: 8px;
}

.rank-table th:last-child {
  border-top-right-radius: 8px;
  border-bottom-right-radius: 8px;
}

.rank-table tbody tr {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transition: all 0.3s ease;
}

.rank-table td {
  padding: 1rem 1.2rem;
  border: none;
  background: transparent;
}

.rank-table td:first-child {
  border-top-left-radius: 8px;
  border-bottom-left-radius: 8px;
}

.rank-table td:last-child {
  border-top-right-radius: 8px;
  border-bottom-right-radius: 8px;
}

.rank-table tbody tr:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
}

/* 更新用户信息样式 */
.user-info {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0.3rem;
  border-radius: 20px;
  transition: all 0.3s ease;
}

.user-info:hover {
  background: rgba(255, 255, 255, 0.1);
}

.user-info img {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(255, 255, 255, 0.2);
  transition: transform 0.3s ease;
}

/* 排名前三的特殊样式 */
.top-three {
  background: rgba(255, 255, 255, 0.08) !important;
  backdrop-filter: blur(12px) !important;
  -webkit-backdrop-filter: blur(12px) !important;
}

.top-three:hover {
  background: rgba(255, 255, 255, 0.12) !important;
}

/* 排序图标样式 */
.sortable {
  cursor: pointer;
  transition: all 0.3s ease;
}

.sortable:hover {
  background: rgba(255, 255, 255, 0.15);
  color: var(--text-primary);
}

.sort-icon {
  display: inline-block;
  margin-left: 0.5rem;
  transition: transform 0.3s ease;
  color: var(--text-primary);
}

.sortable:hover .sort-icon {
  transform: scale(1.2);
}

/* 加载动画样式 */
.loading {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 8px;
}

.loading-spinner {
  margin: 0 auto;
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s linear infinite;
}

/* 移除不需要的样式 */
.glass-effect {
  background: transparent;
  backdrop-filter: none;
  border: none;
  border-radius: 0;
}

/* 更新链接样式 */
.username-link {
  color: var(--text-primary);
  text-decoration: none;
  transition: all 0.3s ease;
  font-weight: 500;
}

.username-link:hover {
  color: var(--primary-color);
  text-shadow: 0 0 10px rgba(var(--primary-color-rgb), 0.3);
}

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

/* 适配浅色主题 */
:global(.light-theme) .page-btn {
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-dark);
}

:global(.light-theme) .page-size-select {
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-dark);
}

@media (max-width: 768px) {
  .rank-header {
    flex-direction: column;
    gap: 1rem;
  }

  .rank-info {
    flex-direction: column;
    gap: 1rem;
  }

  .rank-table {
    font-size: 0.9rem;
  }

  .pagination-buttons {
    flex-wrap: wrap;
    justify-content: center;
  }

  .page-btn {
    padding: 0.4rem 0.8rem;
    font-size: 0.9rem;
  }
}

.username-link {
  color: var(--text-light);
  text-decoration: none;
  transition: color 0.2s;
}

.username-link:hover {
  color: var(--primary-color);
}

/* 适配浅色主题 */
:global(.light-theme) .username-link {
  color: var(--text-dark);
}

.glowing-text {
  color: var(--text-light);
  text-shadow: 0 0 10px var(--primary-color);
}

.glass-effect {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
}

.rank-table th {
  color: var(--primary-color);
  border-bottom: 2px solid rgba(var(--primary-color-rgb), 0.3);
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid rgba(var(--primary-color-rgb), 0.3);
}

.top-three {
  background: rgba(var(--primary-color-rgb), 0.1);
}

.score {
  color: var(--primary-color);
  font-weight: bold;
}

.sortable:hover {
  color: var(--primary-color);
  cursor: pointer;
}

.sort-icon {
  margin-left: 0.5rem;
  transition: transform 0.3s ease;
}

.sortable:hover .sort-icon {
  transform: scale(1.2);
}

.loading-spinner {
  border-top-color: var(--primary-color);
}

.page-btn {
  transition: all 0.3s ease;
}

.page-btn:hover:not(:disabled) {
  background: var(--primary-color);
  color: var(--bg-dark);
}

.username-link:hover {
  color: var(--primary-color);
  text-shadow: 0 0 5px rgba(var(--primary-color-rgb), 0.5);
}
</style>
