<template>
  <div class="problems">
    <div class="problems-header">
      <h1>题目列表</h1>
      <div class="filter-section">
        <div class="difficulty-filter">
          <select v-model="selectedDifficulty" @change="handleFilter">
            <option value="">全部难度</option>
            <option v-for="level in 5" :key="level" :value="level">
              {{ getDifficultyLabel(level) }}
            </option>
          </select>
        </div>
        <div class="search-box">
          <input v-model="searchQuery" type="text" placeholder="搜索题目内容..." />
        </div>
        <div>
          <button @click="debounceSearch" class="page-btn">搜索</button>
        </div>
      </div>
    </div>

    <div class="problems-list">
      <table class="problems-table">
        <thead>
          <tr>
            <th>
              <span class="status">状态</span>
            </th>
            <th>题号</th>
            <th>标题</th>
            <th>难度</th>
            <th @click="handleSort('acceptedCount')" class="sortable">
              通过量
              <span class="sort-icon" v-if="store.sortBy === 'acceptedCount'">
                {{ store.sortOrder === 'desc' ? '↓' : '↑' }}
              </span>
            </th>
            <th @click="handleSort('submissionCount')" class="sortable">
              提交量
              <span class="sort-icon" v-if="store.sortBy === 'submissionCount'">
                {{ store.sortOrder === 'desc' ? '↓' : '↑' }}
              </span>
            </th>
            <th>来源</th>
          </tr>
        </thead>
        <tbody>
          <template v-if="!store.loading">
            <template v-if="filteredProblems.length > 0">
              <tr v-for="problem in filteredProblems" :key="problem.id">
                <td>{{ getProblemStatus(problem.status) }}</td>
                <td>
                  <router-link
                    :to="`/problem/${problem.id}`"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="problem-link"
                  >
                    {{ problem.id }}
                  </router-link>
                </td>
                <td>
                  <router-link
                    :to="`/problem/${problem.id}`"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="problem-link"
                  >
                    {{
                      problem.title.length > 16
                        ? problem.title.substring(0, 16) + '...'
                        : problem.title
                    }}
                  </router-link>
                </td>
                <td>
                  <div class="difficulty-display">
                    <div class="difficulty-progress">
                      <div
                        class="difficulty-bar"
                        :style="{
                          width: `${problem.difficulty * 20}%`,
                          background: getDifficultyGradient(problem.difficulty)
                        }"
                      ></div>
                    </div>
                    <span class="difficulty-text">{{ getDifficultyLabel(problem.difficulty) }}</span>
                  </div>
                </td>
                <td>{{ problem.acceptedCount }}</td>
                <td>{{ problem.submissionCount }}</td>
                <td>
                  {{
                    problem.source.length > 15
                      ? problem.source.substring(0, 7) + '...' + problem.source.substring(problem.source.length - 10)
                      : problem.source
                  }}
                </td>
              </tr>
            </template>
            <tr v-else>
              <td colspan="7" class="no-data">
                <div class="no-data-content">
                  <i class="fas fa-inbox"></i>
                  <span>暂无数据</span>
                </div>
              </td>
            </tr>
          </template>
          <tr v-else>
            <td colspan="7" class="loading">
              <div class="loading-spinner"></div>
              <div class="loading-text">加载中...</div>
            </td>
          </tr>
        </tbody>
      </table>
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

<style scoped>
.problems {
  padding: 80px 2rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.problems-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.filter-section {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.difficulty-filter select {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(0, 0, 0, 0.3);
  color: var(--text-light);
  width: 100px;
}

.search-box input {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-light);
}

.problems-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0 8px;
  background: transparent;
  margin: 1rem 0;
}

.problems-table th {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  padding: 1rem 1.2rem;
  font-weight: 600;
  text-align: left;
  color: var(--text-primary);
  border: none;
  white-space: nowrap;
}

.problems-table th:first-child {
  border-top-left-radius: 8px;
  border-bottom-left-radius: 8px;
}

.problems-table th:last-child {
  border-top-right-radius: 8px;
  border-bottom-right-radius: 8px;
}

.problems-table tbody tr {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transition: all 0.3s ease;
}

.problems-table td {
  padding: 1rem 1.2rem;
  border: none;
  background: transparent;
}

.problems-table td:first-child {
  border-top-left-radius: 8px;
  border-bottom-left-radius: 8px;
}

.problems-table td:last-child {
  border-top-right-radius: 8px;
  border-bottom-right-radius: 8px;
}

.problems-table tbody tr:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
}

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

.problem-link {
  color: var(--text-primary);
  text-decoration: none;
  transition: all 0.3s ease;
  font-weight: 500;
}

.problem-link:hover {
  color: var(--primary-color);
  text-shadow: 0 0 10px rgba(var(--primary-color-rgb), 0.3);
}

.difficulty-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
}

.difficulty-progress {
  width: 60px;
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
  position: relative;
}

.difficulty-bar {
  height: 100%;
  border-radius: 3px;
  transition: all 0.3s ease;
}

.difficulty-text {
  font-size: 0.85rem;
  color: var(--text-light);
  white-space: nowrap;
  min-width: 32px;
}

.difficulty, .star {
  display: none;
}

.loading {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 8px;
  padding: 2rem !important;
  text-align: center;
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

.loading-text {
  margin-top: 1rem;
  color: var(--text-light);
  font-size: 0.9rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.status {
  padding: 0.5rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  width: 32px;
  height: 32px;
}

.light-theme .page-btn {
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-dark);
}

.light-theme .page-size-select {
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-dark);
}

.light-theme .problem-link {
  color: var(--text-dark);
}

@media (max-width: 768px) {
  .problems-header {
    flex-direction: column;
    gap: 1rem;
  }

  .filter-section {
    flex-direction: column;
    gap: 1rem;
  }

  .problems-table {
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

.no-data {
  padding: 3rem !important;
  text-align: center;
  background: rgba(255, 255, 255, 0.02) !important;
}

.no-data-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  color: var(--text-light);
  opacity: 0.6;
}

.no-data-content i {
  font-size: 2rem;
}

.no-data-content span {
  font-size: 1rem;
}
</style>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useProblemsStore } from '@/stores/modules/problems'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'

const store = useProblemsStore()
const router = useRouter()
const userStore = useUserStore()
const currentPage = ref(1)
const pageSize = ref(20)
const selectedDifficulty = ref('')
const searchQuery = ref('')

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

const fetchProblems = () => {
  const token = userStore.token
  if (!token) {
    router.push('/sign-in')
    return
  }

  store.fetchProblems({
    page: currentPage.value,
    pageSize: pageSize.value,
    difficulty: selectedDifficulty.value ? Number(selectedDifficulty.value) : undefined,
    search: searchQuery.value,
    sortBy: store.sortBy,
    sortOrder: store.sortOrder,
  })
}

const handleSort = (field: string) => {
  store.updateSort(field)
  fetchProblems()
}

const handleFilter = () => {
  currentPage.value = 1
  fetchProblems()
}

const handlePageSizeChange = () => {
  currentPage.value = 1
  fetchProblems()
}

const goToPage = (page: number) => {
  currentPage.value = page
  fetchProblems()
}

let searchTimeout: ReturnType<typeof setTimeout>
const debounceSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchProblems()
  }, 300)
}

onMounted(() => {
  if (!userStore.token) {
    router.push('/sign-in')
    return
  }
  fetchProblems()
})

const getProblemStatus = (status?: string) => {
  switch (status) {
    case 'accepted':
      return '🟢' // 已通过
    case 'attempted':
      return '🔵' // 已尝试但未通过
    default:
      return '⚪' // 未尝试
  }
}

const getDifficultyLabel = (level: number) => {
  switch (level) {
    case 1:
      return '入门'
    case 2:
      return '简单'
    case 3:
      return '中等'
    case 4:
      return '困难'
    case 5:
      return '专家'
    default:
      return '未知'
  }
}

const getDifficultyGradient = (level: number) => {
  switch (level) {
    case 1:
      return 'linear-gradient(90deg, #00b09b, #96c93d)'
    case 2:
      return 'linear-gradient(90deg, #96c93d, #4facfe)'
    case 3:
      return 'linear-gradient(90deg, #4facfe, #ffd700)'
    case 4:
      return 'linear-gradient(90deg, #ffd700, #ff5858)'
    case 5:
      return 'linear-gradient(90deg, #ff5858, #ff0000)'
    default:
      return 'linear-gradient(90deg, #ccc, #ccc)'
  }
}

const filteredProblems = computed(() => {
  let result = store.problems

  // 难度筛选
  if (selectedDifficulty.value) {
    result = result.filter(problem => problem.difficulty === Number(selectedDifficulty.value))
  }

  // 搜索筛选
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(problem =>
      problem.title.toLowerCase().includes(query) ||
      problem.id.toString().includes(query) ||
      problem.source.toLowerCase().includes(query)
    )
  }

  return result
})
</script>
