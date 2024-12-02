<template>
  <div class="contest-rank">
    <div class="rank-header">
      <div class="title-section">
        <h1 class="glowing-text">比赛排名</h1>
        <span class="contest-badge">{{ contestTitle }}</span>
      </div>
      <div class="control-section">
        <div class="rank-type-switch">
          <el-radio-group v-model="rankType" @change="handleRankTypeChange">
            <el-radio-button label="acm">ACM模式</el-radio-button>
            <el-radio-button label="ioi">IOI模式</el-radio-button>
          </el-radio-group>
        </div>
        <el-button
          type="primary"
          @click="exportRank"
          :loading="exporting"
          class="export-btn"
        >
          <i class="fas fa-download"></i>
          导出排名
        </el-button>
        <div class="search-section">
          <div class="search-box">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索用户..."
              @keyup.enter="handleSearch"
            />
            <button @click="handleSearch" class="search-btn">
              <i class="fas fa-search"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="rank-table">
      <table>
        <thead>
          <tr>
            <th class="rank-col">排名</th>
            <th class="user-col">用户</th>
            <template v-if="rankType === 'acm'">
              <th class="score-col">解题数</th>
              <th class="penalty-col">罚时</th>
            </template>
            <template v-else>
              <th class="score-col">总分</th>
            </template>
            <th v-for="problemId in problems" :key="problemId" class="problem-col">
              <router-link
                :to="`/contest/${contestId}/problem/${getProblemLabel(problemId)}`"
                class="problem-link"
                target="_blank"
              >
                {{ getProblemLabel(problemId) }}
              </router-link>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(rank, index) in currentPageData" :key="rank.userId">
            <td class="rank-col">
              <span class="rank-badge" :class="getRankClass(index + 1)">
                {{ index + 1 }}
              </span>
            </td>
            <td class="user-col">
              <div class="user-info">
                <img :src="rank.avatar" alt="avatar" class="avatar" />
                <span class="username">{{ rank.username }}</span>
              </div>
            </td>
            <template v-if="rankType === 'acm'">
              <td class="score-col">
                <span class="score-badge">{{ rank.solved }}</span>
              </td>
              <td class="penalty-col">
                <span class="penalty-badge">
                  {{ formatPenalty(rank.penalty) }}
                </span>
              </td>
            </template>
            <template v-else>
              <td class="score-col">
                <span class="score-badge">{{ rank.totalScore }}</span>
              </td>
            </template>
            <td
              v-for="problemId in problems"
              :key="problemId"
              class="problem-col"
              :class="{
                accepted: rank.problems[problemId] === 'Accepted',
                failed: rank.problems[problemId] && rank.problems[problemId] !== 'Accepted',
              }"
            >
              <div class="problem-status">
                <template v-if="rankType === 'ioi'">
                  <span :class="['score-badge', getScoreClass(rank.scores?.[problemId])]" v-if="rank.scores?.[problemId] !== undefined">
                    {{ rank.scores[problemId] }}
                  </span>
                  <span class="status-badge status-pending" v-else>-</span>
                </template>
                <template v-else>
                  <span :class="['status-badge', getStatusClass(rank.problems[problemId], rank.attempts?.[problemId])]">
                    {{ getStatusText(rank.problems[problemId], rank.attempts?.[problemId]) }}
                    <span v-if="rank.attempts && rank.attempts[problemId] > 0" class="attempts-badge">
                      ({{ rank.attempts[problemId] }})
                    </span>
                  </span>
                </template>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagination">
      <div class="pagination-info">
        共 {{ total }} 条记录
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
            @click="goToPage(Number(pageNum))"
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
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { ElMessage } from 'element-plus'

interface RankResponse {
  ranks: RankData[]
  problems: string[]
}

interface RankData {
  userId: number
  username: string
  avatar: string
  problems: Record<string, string>
  attempts: Record<string, number>
  solved: number
  penalty: number
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const contestId = route.params.id as string

const searchQuery = ref('')
const rankings = ref<RankData[]>([])
const problems = ref<string[]>([])

// 添加比赛标题
const contestTitle = ref('')

// 添加排名方式状态
const rankType = ref('acm')

// 添加排名方式切换处理函数
const handleRankTypeChange = () => {
  fetchRankings()
}

// 添加计算属性过滤排名
const filteredRankings = computed(() => {
  if (!searchQuery.value) return rankings.value

  const query = searchQuery.value.toLowerCase()
  return rankings.value.filter((rank: RankData) => rank.username.toLowerCase().includes(query))
})

// 添加防抖搜索
let searchTimeout: ReturnType<typeof setTimeout>
const handleSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    fetchRankings()
  }, 300)
}

const fetchRankings = async () => {
  try {
    // 获取比赛信息
    const contestResponse = await fetch(`/api/contests/${contestId}`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (contestResponse.ok) {
      const contestData = await contestResponse.json()
      if (contestData.code === 200) {
        contestTitle.value = contestData.data.title
      }
    }

    const url = new URL(`/api/contests/${contestId}/rank`, window.location.origin)
    if (searchQuery.value) {
      url.searchParams.set('username', searchQuery.value)
    }
    // 添加排名方式参数
    url.searchParams.set('type', rankType.value)

    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('获取排名失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      const responseData = data.data as RankResponse
      rankings.value = responseData.ranks
      problems.value = responseData.problems
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取数据失败')
  }
}

const getStatusClass = (status: string | undefined, attempts: number = 0) => {
  if (attempts > 0 && !status) {
    return 'status-failed'  // 有尝试但未通过
  }
  switch (status) {
    case 'Accepted':
      return 'status-accepted'
    case 'Wrong Answer':
    case 'Time Limit Exceeded':
    case 'Memory Limit Exceeded':
    case 'Runtime Error':
    case 'Compile Error':
      return 'status-failed'
    default:
      return 'status-pending'
  }
}

const getStatusText = (status: string | undefined, attempts: number = 0) => {
  if (attempts > 0 && !status) {
    return '未通过'  // 有尝试但未通过
  }
  if (!status) return '未提交'
  switch (status) {
    case 'Accepted':
      return 'AC'
    case 'Wrong Answer':
      return 'WA'
    case 'Time Limit Exceeded':
      return 'TLE'
    case 'Memory Limit Exceeded':
      return 'MLE'
    case 'Runtime Error':
      return 'RE'
    case 'Compile Error':
      return 'CE'
    default:
      return '?'
  }
}

// 添加题目标签转换函数
const getProblemLabel = (problemId: string) => {
  const index = problems.value.indexOf(problemId)
  return String.fromCharCode(65 + index) // A, B, C, D...
}

// 添加排名样式函数
const getRankClass = (rank: number) => {
  if (rank === 1) return 'rank-first'
  if (rank === 2) return 'rank-second'
  if (rank === 3) return 'rank-third'
  return ''
}

// 添加分页相关的响��式变量
const currentPage = ref(1)
const pageSize = ref(20)
const total = computed(() => filteredRankings.value.length)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 计算当前页的数据
const currentPageData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredRankings.value.slice(start, end)
})

// 计算显示的页码范围
const displayedPages = computed(() => {
  const delta = 2
  const range = []
  const rangeWithDots: (number | string)[] = []
  let l: number | null = null

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

// 页码变化处理函数
const handlePageSizeChange = () => {
  currentPage.value = 1
}

const goToPage = (page: number) => {
  currentPage.value = page
}

// 添加罚时格式化函数
const formatPenalty = (penalty: number) => {
  const hours = Math.floor(penalty / 60)
  const minutes = penalty % 60
  return `${hours}:${minutes.toString().padStart(2, '0')}`
}

// 添加得分样式函数
const getScoreClass = (score: number | undefined) => {
  if (score === undefined) return 'status-pending'
  if (score === 100) return 'score-perfect'
  if (score >= 60) return 'score-pass'
  return 'score-fail'
}

// 添加导出相关的响应式变量和方法
const exporting = ref(false)

const exportRank = async () => {
  try {
    exporting.value = true
    const response = await fetch(
      `/api/contests/${contestId}/rank/export?type=${rankType.value}`,
      {
        headers: {
          Authorization: `Bearer ${userStore.token}`,
        },
      }
    )

    if (!response.ok) {
      throw new Error('导出失败')
    }

    // 获取文件名
    const contentDisposition = response.headers.get('content-disposition')
    const filename = contentDisposition
      ? contentDisposition.split('filename=')[1]
      : `contest_rank_${Date.now()}.xlsx`

    // 下载文件
    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '导出失败')
  } finally {
    exporting.value = false
  }
}

onMounted(() => {
  if (!userStore.token) {
    router.push('/sign-in')
    return
  }
  fetchRankings()
})
</script>

<style scoped>
.contest-rank {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem;
  margin-top: 64px;
}

.rank-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.title-section {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.search-section {
  min-width: 300px;
}

.glowing-text {
  color: var(--text-primary);
  text-shadow: 0 0 10px var(--primary-color);
  font-size: 2rem;
  margin: 0;
}

.search-box {
  display: flex;
  gap: 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  padding: 0.5rem;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.search-box input {
  width: 200px;
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

.search-btn {
  padding: 0.5rem 1rem;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.search-btn:hover {
  background: var(--primary-color-dark);
  transform: translateY(-1px);
}

.rank-table {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 1rem;
  margin-top: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow-x: auto;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  min-width: 800px;
}

table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  white-space: nowrap;
}

th,
td {
  padding: 1rem;
  text-align: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.rank-col {
  width: 80px;
}

.rank-badge {
  display: inline-block;
  width: 36px;
  height: 36px;
  line-height: 36px;
  border-radius: 50%;
  font-weight: bold;
  background: rgba(255, 255, 255, 0.1);
}

.rank-first {
  background: linear-gradient(135deg, #ffd700, #ffa500);
  color: white;
  box-shadow: 0 0 10px rgba(255, 215, 0, 0.5);
}

.rank-second {
  background: linear-gradient(135deg, #c0c0c0, #a9a9a9);
  color: white;
  box-shadow: 0 0 10px rgba(192, 192, 192, 0.5);
}

.rank-third {
  background: linear-gradient(135deg, #cd7f32, #8b4513);
  color: white;
  box-shadow: 0 0 10px rgba(205, 127, 50, 0.5);
}

.user-col {
  width: 200px;
  text-align: left;
}

.score-col {
  width: 100px;
}

.score-badge {
  display: inline-block;
  padding: 0.3rem 1rem;
  background: linear-gradient(135deg, #006996, #0088be);
  color: white;
  border-radius: 12px;
  font-weight: bold;
}

.problem-col {
  width: 100px;
  transition: all 0.3s ease;
}

.problem-col:hover {
  background: rgba(255, 255, 255, 0.05);
}

.problem-col.accepted,
.problem-col.failed {
  background: none;
  border: none;
}

th {
  font-weight: 500;
  color: var(--text-light);
  position: sticky;
  top: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 1;
  backdrop-filter: blur(8px);
}

tr:hover {
  background: rgba(255, 255, 255, 0.03);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(var(--primary-color-rgb), 0.3);
  transition: all 0.3s ease;
}

tr:hover .avatar {
  border-color: var(--primary-color);
  transform: scale(1.1);
}

.username {
  font-weight: 500;
  color: var(--primary-color);
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

.status-badge {
  padding: 0.3rem 0.8rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  min-width: 60px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.status-accepted {
  background: linear-gradient(135deg, #00b09b, #96c93d);
  color: white;
  box-shadow: 0 2px 8px rgba(0, 176, 155, 0.2);
}

.status-failed {
  background: linear-gradient(135deg, #ff416c, #ff4b2b);
  color: white;
  box-shadow: 0 2px 8px rgba(255, 65, 108, 0.2);
}

.status-pending {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-light);
}

.problem-link {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 500;
  transition: all 0.3s ease;
  padding: 0.3rem 0.8rem;
  border-radius: 4px;
  background: rgba(var(--primary-color-rgb), 0.1);
}

.problem-link:hover {
  background: rgba(var(--primary-color-rgb), 0.2);
  transform: translateY(-1px);
}

@media (max-width: 768px) {
  .contest-rank {
    padding: 1rem;
  }

  .rank-header {
    flex-direction: column;
    gap: 1rem;
  }

  .title-section {
    flex-direction: column;
    align-items: flex-start;
  }

  .search-section {
    width: 100%;
  }

  .search-box input {
    width: 100%;
  }
}

.contest-info {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
}

.contest-badge {
  padding: 0.5rem 1.5rem;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  color: var(--text-light);
  font-size: 1.2rem;
  font-weight: 500;
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.attempts-badge {
  font-size: 0.8rem;
  font-weight: normal;
  opacity: 0.9;
  margin-left: 0.1rem;
}

.problem-status {
  display: flex;
  justify-content: center;
  align-items: center;
}

.penalty-col {
  width: 100px;
}

.penalty-badge {
  display: inline-block;
  padding: 0.3rem 1rem;
  background: linear-gradient(135deg, #834d9b, #d04ed6);
  color: white;
  border-radius: 12px;
  font-weight: bold;
  font-family: 'Fira Code', monospace;
  box-shadow: 0 2px 8px rgba(131, 77, 155, 0.3);
  transition: all 0.3s ease;
}

.penalty-badge:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(131, 77, 155, 0.4);
}

th.penalty-col,
td.penalty-col {
  position: sticky;
  left: 380px;
  background: inherit;
  z-index: 1;
}

tr:hover .penalty-badge {
  background: linear-gradient(135deg, #9d5bb9, #e45ee9);
}

.control-section {
  display: flex;
  align-items: center;
  gap: 2rem;
}

.rank-type-switch {
  display: flex;
  align-items: center;
}

/* 自定义 radio button 样式 */
:deep(.el-radio-group) {
  background: rgba(255, 255, 255, 0.05);
  padding: 4px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

:deep(.el-radio-button__inner) {
  background: transparent;
  border: none;
  color: var(--text-light);
  transition: all 0.3s ease;
}

:deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: var(--primary-color);
  color: white;
  box-shadow: none;
}

:deep(.el-radio-button:first-child .el-radio-button__inner) {
  border-radius: 4px;
}

:deep(.el-radio-button:last-child .el-radio-button__inner) {
  border-radius: 4px;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .control-section {
    flex-direction: column;
    gap: 1rem;
  }

  .rank-type-switch {
    width: 100%;
  }

  :deep(.el-radio-group) {
    width: 100%;
    display: flex;
  }

  :deep(.el-radio-button) {
    flex: 1;
  }
}

/* 添加IOI模式的分数样式 */
.score-perfect {
  background: linear-gradient(135deg, #00b09b, #96c93d);
  color: white;
  box-shadow: 0 2px 8px rgba(0, 176, 155, 0.2);
}

.score-pass {
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  color: white;
  box-shadow: 0 2px 8px rgba(79, 172, 254, 0.2);
}

.score-fail {
  background: linear-gradient(135deg, #ff758c, #ff7eb3);
  color: white;
  box-shadow: 0 2px 8px rgba(255, 117, 140, 0.2);
}

/* 调整分数显示的样式 */
.score-badge {
  padding: 0.3rem 0.8rem;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: 500;
  min-width: 60px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

/* 添加导出按钮样式 */
.export-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--primary-color);
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.export-btn:hover {
  background: var(--primary-color-dark);
  transform: translateY(-1px);
}

.export-btn i {
  font-size: 0.9rem;
}

/* 确保按钮在移动端也能正确显示 */
@media (max-width: 768px) {
  .control-section {
    flex-direction: column;
    gap: 1rem;
  }

  .export-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
