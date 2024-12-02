<template>
  <!-- 修改加载状态的显示逻辑 -->
  <div class="submission">
    <h1>评测记录</h1>

    <!-- 筛选器部分保持不变 -->
    <div class="filters">
      <input
        type="text"
        v-model="problemId"
        placeholder="题目ID..."
        class="search-input"
      />

      <input
        type="text"
        v-model="contestId"
        placeholder="比赛ID..."
        class="search-input"
      />

      <input
        type="text"
        v-model="username"
        placeholder="用户名..."
        class="search-input"
      />

      <select v-model="status">
        <option value="">全部状态</option>
        <option value="Pending">等待评测</option>
        <option value="Judging">评测中</option>
        <option value="Accepted">通过</option>
        <option value="Wrong Answer">答案错误</option>
        <option value="Time Limit Exceeded">超时</option>
        <option value="Memory Limit Exceeded">内存超限</option>
        <option value="Runtime Error">运行时错误</option>
        <option value="Compile Error">编译错误</option>
        <option value="System Error">系统错误</option>
      </select>

      <button class="search-btn" @click="handleSearch">
        <i class="fas fa-search"></i>
        搜索
      </button>
    </div>

    <!-- 表格部分 -->
    <table class="submission-table">
      <thead>
        <tr>
          <th>状态</th>
          <th>提交编号</th>
          <th>题目</th>
          <th>语言</th>
          <th>耗时</th>
          <th>内存</th>
          <th>代码长度</th>
          <th>提交者</th>
        </tr>
      </thead>
      <tbody>
        <!-- 加载中状态 -->
        <tr v-if="loading">
          <td colspan="8" class="loading">
            <div class="loading-spinner"></div>
            <div class="loading-text">加载中...</div>
          </td>
        </tr>

        <!-- 无数据状态 -->
        <tr v-else-if="submissions.length === 0">
          <td colspan="8" class="no-data">
            <div class="no-data-content">
              <i class="fas fa-inbox"></i>
              <p>暂无提交记录</p>
            </div>
          </td>
        </tr>

        <!-- 数据列表 -->
        <template v-else>
          <tr v-for="submission in submissions" :key="submission.ID">
            <td>
              <span :class="['status-badge', getStatusClass(submission.Status)]">
                {{ submission.Status }}
              </span>
            </td>
            <td>
              <router-link
                :to="`/submission/${submission.ID}`"
                target="_blank"
                rel="noopener noreferrer"
              >
                <span class="id-badge"> #{{ submission.ID }} </span>
              </router-link>
            </td>
            <td>
              <router-link
                :to="`/problem/${submission.ProblemID}`"
                target="_blank"
                rel="noopener noreferrer"
              >
                <span class="problem-id-badge"> #{{ submission.ProblemID }} </span>
              </router-link>
            </td>
            <td>
              <span :class="['language-badge', submission.Language.toLowerCase()]">
                {{ getLanguageDisplay(submission.Language) }}
              </span>
            </td>
            <td>
              <span class="time-badge"> {{ submission.TimeUsed }}ms </span>
            </td>
            <td>
              <span class="memory-badge"> {{ submission.MemoryUsed }}KB </span>
            </td>
            <td>
              <span class="code-length-badge">
                {{ formatCodeLength(submission.Code?.length) }}
              </span>
            </td>
            <td>
              <div class="user-info">
                <img
                  :src="submission.userAvatar || '/images/avatars/default-avatar.png'"
                  alt="avatar"
                />
                <router-link
                  :to="`/profile/${submission.username}`"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {{ submission.username }}
                </router-link>
              </div>
            </td>
          </tr>
        </template>
      </tbody>
    </table>

    <!-- 分页器只在有数据时显示 -->
    <div class="pagination" v-if="submissions.length > 0">
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
          @click="goToPage(Number(totalPages))"
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
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/modules/app'
import type { Submission } from '@/types/submission'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()

// 状态和搜索
const status = ref('')
const submissions = ref<Submission[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const loading = ref(false)

// 计算总页数
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

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

// 添加新的状态
const problemId = ref('')
const contestId = ref('')
const username = ref('')


// 添加一个新的 ref 来存储原始数据
const allSubmissions = ref<Submission[]>([])

// 修改 fetchSubmissions 函数
const fetchSubmissions = async () => {
  try {
    if (!userStore.isLoggedIn) {
      // console.log('等待用户登录...')
      return
    }

    loading.value = true
    // 始终使用基础URL
    const url = '/api/submissions'
    const params = new URLSearchParams()

    // 添加所有筛选参数
    if (problemId.value) {
      params.append('problemId', problemId.value)
    }
    if (username.value) {
      params.append('username', username.value)
    }
    if (contestId.value) {
      params.append('contestId', contestId.value)
    }
    if (status.value) {
      params.append('status', status.value)
    }

    const response = await fetch(`${url}${params.toString() ? '?' + params.toString() : ''}`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error('获取提交记录失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      // 保存原始数据
      allSubmissions.value = data.data.submissions.map((submission: Partial<Submission>) => ({
        ...submission,
        ID: submission.ID,
        ProblemID: submission.ProblemID,
        Language: submission.Language,
        Status: submission.Status,
        TimeUsed: submission.TimeUsed,
        MemoryUsed: submission.MemoryUsed,
        SubmitTime: submission.SubmitTime,
        problemTitle: submission.problemTitle,
        username: submission.username,
        userAvatar: submission.userAvatar,
      }))

      // 应用状态筛选
      let filteredData = [...allSubmissions.value]
      if (status.value) {
        filteredData = filteredData.filter((submission) => submission.Status === status.value)
      }

      // 更新显示的数据和总数
      submissions.value = filteredData
      total.value = filteredData.length
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    console.error('获取提交记录失败:', error)
    appStore.showNotification('error', '获取提交记录失败')
  } finally {
    loading.value = false
  }
}

// 监听状态变化时重新筛选数据
watch(status, () => {
  let filteredData = [...allSubmissions.value]
  if (status.value) {
    filteredData = filteredData.filter((submission) => submission.Status === status.value)
  }
  submissions.value = filteredData
  total.value = filteredData.length
})

// 修改搜索处理函数
const handleSearch = () => {
  // 构建查询参数
  const query: Record<string, string> = {}

  // 添加所有筛选参数
  if (status.value) {
    query.status = status.value
  }
  if (problemId.value) {
    query.problemId = problemId.value
  }
  if (contestId.value) {
    query.contestId = contestId.value
  }
  if (username.value) {
    query.username = username.value
  }

  // 更新路由
  router.replace({ query })

  // 重置页码并获取数据
  currentPage.value = 1
  fetchSubmissions()
}

// 修改页码变化处理函数
const goToPage = (page: number) => {
  currentPage.value = page
  fetchSubmissions()
}

// 修改每页显示数量
const handlePageSizeChange = () => {
  currentPage.value = 1
  fetchSubmissions()
}

// 更新状态样式函数
const getStatusClass = (status: string) => {
  return status.toLowerCase().replace(/\s+/g, '-')
}

// 格式化代码长度
const formatCodeLength = (length: number) => {
  if (!length) return '0B'
  if (length < 1024) return `${length}B`
  return `${(length / 1024).toFixed(1)}KB`
}

// 添加语言映射函数
const getLanguageDisplay = (lang: string) => {
  const langMap: { [key: string]: string } = {
    cpp: 'C++',
    java: 'Java',
    python: 'Python',
    go: 'Go',
    javascript: 'JavaScript',
    typescript: 'TypeScript',
  }
  return langMap[lang] || lang
}



// 修改 onMounted
onMounted(() => {
  if (!userStore.isLoggedIn) {
    router.push({
      path: '/sign-in',
      query: { redirect: route.fullPath },
    })
    return
  }

  // 设置默认分页值
  currentPage.value = 1
  pageSize.value = 20

  fetchSubmissions()
})

// 修改路由监听
watch(
  () => route.query,
  (newQuery: Record<string, string>) => {
    // 只在搜索参数变化时重新获取数据
    if (newQuery.problemId !== route.query.problemId ||
        newQuery.contestId !== route.query.contestId ||
        newQuery.username !== route.query.username) {
      currentPage.value = 1
      fetchSubmissions()
    }
  }
)

// 监听用户信息变化
watch(
  () => userStore.userInfo,
  (newUser: { username: string } | null) => {
    if (newUser && !route.query.problemId && !route.query.contestId && !route.query.username) {
      username.value = newUser.username
      fetchSubmissions()
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.submission {
  padding: 80px 2rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.filters {
  margin: 2rem 0;
  display: flex;
  gap: 1rem;
  border-radius: 4px;
}

.search-container {
  display: flex;
  gap: 0.5rem;
  flex: 2;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
}

.search-container input {
  flex: 1;
  padding: 0.5rem 1rem;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 0.95rem;
  width: 100%;
}

.search-container input::placeholder {
  color: var(--text-secondary);
  opacity: 0.7;
}

.filters select,
.filters input {
  padding: 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(0, 0, 0, 0.3);
  color: var(--text-primary);
}

.filters input {
  flex: 1;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
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
  gap: 0.5rem;
}

.search-btn:hover {
  background: var(--primary-color-dark);
}

.submission-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0 8px;
  background: transparent;
  margin: 1rem 0;
}

.submission-table th {
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

.submission-table th:first-child {
  border-top-left-radius: 8px;
  border-bottom-left-radius: 8px;
}

.submission-table th:last-child {
  border-top-right-radius: 8px;
  border-bottom-right-radius: 8px;
}

.submission-table tbody tr {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transition: all 0.3s ease;
}

.submission-table td {
  padding: 1rem 1.2rem;
  border: none;
  background: transparent;
}

.submission-table td:first-child {
  border-top-left-radius: 8px;
  border-bottom-left-radius: 8px;
}

.submission-table td:last-child {
  border-top-right-radius: 8px;
  border-bottom-right-radius: 8px;
}

.submission-table tbody tr:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
}

/* 通用徽章样式 */
.status-badge,
.language-badge,
.time-badge,
.memory-badge {
  padding: 0.4rem 0.8rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 500;
  display: inline-block;
  text-align: center;
  color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
}

/* 状态徽章样式 */
.status-badge.accepted {
  background: linear-gradient(135deg, #00b09b, #96c93d);
  box-shadow: 0 2px 8px rgba(0, 176, 155, 0.3);
}

.status-badge.wrong-answer {
  background: linear-gradient(135deg, #ff416c, #ff4b2b);
  box-shadow: 0 2px 8px rgba(255, 65, 108, 0.3);
}

.status-badge.runtime-error {
  background: linear-gradient(135deg, #f7b733, #fc4a1a);
  box-shadow: 0 2px 8px rgba(247, 183, 51, 0.3);
}

.status-badge.time-limit-exceeded {
  background: linear-gradient(135deg, #834d9b, #d04ed6);
  box-shadow: 0 2px 8px rgba(131, 77, 155, 0.3);
}

.status-badge.memory-limit-exceeded {
  background: linear-gradient(135deg, #4568dc, #b06ab3);
  box-shadow: 0 2px 8px rgba(69, 104, 220, 0.3);
}

.status-badge.compile-error {
  background: linear-gradient(135deg, #373b44, #4286f4);
  box-shadow: 0 2px 8px rgba(55, 59, 68, 0.3);
}

.status-badge.pending,
.status-badge.judging {
  background: linear-gradient(135deg, #2c3e50, #3498db);
  box-shadow: 0 2px 8px rgba(44, 62, 80, 0.3);
  animation: pulse 2s infinite;
}

/* 语言徽章样式 */
.language-badge.cpp {
  background: linear-gradient(135deg, #0072c6, #00a4db);
  box-shadow: 0 2px 8px rgba(0, 114, 198, 0.3);
}

.language-badge.c {
  background: linear-gradient(135deg, #283593, #5c6bc0);
  box-shadow: 0 2px 8px rgba(40, 53, 147, 0.3);
}

.language-badge.java {
  background: linear-gradient(135deg, #f44336, #ff7043);
  box-shadow: 0 2px 8px rgba(244, 67, 54, 0.3);
}

.language-badge.python {
  background: linear-gradient(135deg, #ffd54f, #ffb300);
  box-shadow: 0 2px 8px rgba(255, 213, 79, 0.3);
}

.language-badge.go {
  background: linear-gradient(135deg, #00bcd4, #26c6da);
  box-shadow: 0 2px 8px rgba(0, 188, 212, 0.3);
}

/* 耗时徽章样式 */
.time-badge {
  background: linear-gradient(135deg, #4caf50, #81c784);
  box-shadow: 0 2px 8px rgba(76, 175, 80, 0.3);
}

/* 内存徽章样式 */
.memory-badge {
  background: linear-gradient(135deg, #9c27b0, #ba68c8);
  box-shadow: 0 2px 8px rgba(156, 39, 176, 0.3);
}

/* 悬停效果 */
.status-badge:hover,
.language-badge:hover,
.time-badge:hover,
.memory-badge:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

/* 脉冲动画 */
@keyframes pulse {
  0% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.05);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

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

.user-info {
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

.user-info img {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--border-color);
  transition: transform 0.3s ease;
}

.user-info:hover img {
  transform: scale(1.1);
}

.submission-table a {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 500;
  transition: all 0.3s ease;
}

.submission-table a:hover {
  color: var(--primary-color-dark);
  text-decoration: none;
}

tr:hover {
  background: var(--bg-hover);
}

a {
  color: var(--primary-color);
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
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

.search-type {
  min-width: 100px;
}

.search-id {
  min-width: 150px;
  padding: 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-secondary);
  color: var(--text-primary);
}

/* 加载动画样式 */
.loading {
  padding: 2rem !important;
  text-align: center;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 8px;
}

.loading-spinner {
  margin: 0 auto;
  width: 30px;
  height: 30px;
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s linear infinite;
}

.loading-text {
  margin-top: 0.5rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* 提交编号和代码长度徽章基础样式 */
.id-badge,
.code-length-badge {
  padding: 0.4rem 0.8rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 500;
  display: inline-block;
  text-align: center;
  transition: all 0.3s ease;
  color: var(--text-primary);
}

/* 提交编号徽章样式 */
.id-badge {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 代码长度徽章样式 */
.code-length-badge {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.03));
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

/* 悬停效果 */
.id-badge:hover,
.code-length-badge:hover {
  transform: translateY(-2px);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.15), rgba(255, 255, 255, 0.08));
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* 移除链接的默认样式 */
.submission-table a {
  text-decoration: none;
}

/* 移除链接悬停时的下划线 */
.submission-table a:hover {
  text-decoration: none;
}

/* 确保徽章在链接内部时保持样式 */
.submission-table a .id-badge {
  color: var(--text-primary);
}

/* 题目链接容器 */
.submission-table a {
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

/* 题号徽章样式 */
.problem-id-badge {
  padding: 0.4rem 0.8rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 500;
  display: inline-block;
  text-align: center;
  color: white; /* 改为白色 */
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.2),
    /* 稍微提高不透明度 */ rgba(255, 255, 255, 0.1)
  );
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

/* 题目标题样式 */
.problem-title {
  color: var(--text-primary);
  transition: all 0.3s ease;
  position: relative;
}

/* 悬停效果 */
.submission-table a:hover .problem-id-badge {
  transform: translateY(-2px);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.25), rgba(255, 255, 255, 0.15));
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.submission-table a:hover .problem-title {
  color: var(--primary-color);
}

/* 添加题目标题的下划线动画 */
.problem-title::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 100%;
  height: 1px;
  background: var(--primary-color);
  transform: scaleX(0);
  transform-origin: right;
  transition: transform 0.3s ease;
}

.submission-table a:hover .problem-title::after {
  transform: scaleX(1);
  transform-origin: left;
}

/* System Error - 系统错误 */
.status-badge.system-error {
  background: linear-gradient(135deg, #cb356b, #bd3f32);
  box-shadow: 0 2px 8px rgba(203, 53, 107, 0.3);
}

/* 添加无数据状态的样式 */
.no-data {
  padding: 3rem !important;
  text-align: center;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 8px;
}

.no-data-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  color: var(--text-secondary);
}

.no-data-content i {
  font-size: 2rem;
  opacity: 0.5;
}

.no-data-content p {
  font-size: 1rem;
  margin: 0;
}

/* 修改输入框样式 */
.search-input {
  padding: 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
  min-width: 120px;
}

.search-input::placeholder {
  color: var(--text-secondary);
  opacity: 0.7;
}
</style>
