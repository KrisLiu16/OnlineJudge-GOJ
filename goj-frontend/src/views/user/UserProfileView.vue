<template>
  <div class="user-profile">
    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner"></div>
      <div class="loading-text">加载中...</div>
    </div>
    <div v-else class="profile-header">
      <div class="avatar-section">
        <img
          :src="profileData.avatar || '/images/avatars/default-avatar.png'"
          :alt="profileData.username"
          class="avatar"
        />
        <div v-if="isCurrentUser" class="upload-overlay">
          <button @click="handleAvatarUploadNotSupported" class="upload-btn">更换头像</button>
        </div>
      </div>

      <div class="user-info">
        <div class="username-container">
          <h2 :class="{ 'username admin': profileData.role === 'admin' }">
            {{ profileData.username }}
          </h2>
        </div>
        <p class="bio">{{ profileData.bio || '这个用户很懒，还没有写简介' }}</p>
        <div class="stats">
          <div class="stat-item">
            <span class="number">{{ profileData.submissions || 0 }}</span>
            <span class="label">提交</span>
          </div>
          <div class="stat-item">
            <span class="number">{{ profileData.acceptedProblems || 0 }}</span>
            <span class="label">AC题数</span>
          </div>
          <div class="stat-item">
            <span class="number">{{ profileData.rating || 1500 }}</span>
            <span class="label">积分</span>
          </div>
        </div>
      </div>
    </div>

    <div class="profile-content">
      <div class="tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          :class="['tab-btn', { active: currentTab === tab.key }]"
          @click="currentTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>

      <div class="tab-content">
        <template v-if="currentTab === 'submissions'">
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
              </tr>
            </thead>
            <tbody>
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
                    {{ submission.problemTitle }}
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
              </tr>
            </tbody>
          </table>
          <div class="pagination">
            <div class="pagination-info">
              共 {{ submissionsTotal }} 条记录
              <select
                v-model="submissionsPageSize"
                @change="handleSubmissionsPageSizeChange"
                class="page-size-select"
              >
                <option v-for="size in [20, 50, 100]" :key="size" :value="size">
                  {{ size }} 条/页
                </option>
              </select>
            </div>
            <div class="pagination-buttons">
              <button
                @click="handleSubmissionsPageChange(1)"
                :disabled="submissionsPage === 1"
                class="page-btn"
              >
                首页
              </button>
              <button
                @click="handleSubmissionsPageChange(submissionsPage - 1)"
                :disabled="submissionsPage === 1"
                class="page-btn"
              >
                上一页
              </button>
              <div class="page-numbers">
                <button
                  v-for="pageNum in submissionsDisplayedPages"
                  :key="pageNum"
                  @click="handleSubmissionsPageChange(Number(pageNum))"
                  :class="['page-btn', { active: submissionsPage === pageNum }]"
                >
                  {{ pageNum }}
                </button>
              </div>
              <button
                @click="handleSubmissionsPageChange(submissionsPage + 1)"
                :disabled="submissionsPage === submissionsTotalPages"
                class="page-btn"
              >
                下一页
              </button>
              <button
                @click="handleSubmissionsPageChange(submissionsTotalPages)"
                :disabled="submissionsPage === submissionsTotalPages"
                class="page-btn"
              >
                末页
              </button>
            </div>
          </div>
        </template>

        <template v-if="currentTab === 'problems'">
          <div class="locked-content">
            <el-icon class="lock-icon"><Lock /></el-icon>
            <p class="locked-text">不支持查看...</p>
            <p class="locked-description">!</p>
          </div>
        </template>

        <template v-if="currentTab === 'contests'">
          <div class="locked-content">
            <el-icon class="lock-icon"><Lock /></el-icon>
            <p class="locked-text">不支持查看...</p>
            <p class="locked-description">!</p>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useUserStore } from '@/stores/modules/user'
import { useAppStore } from '@/stores/modules/app'
import { useRoute } from 'vue-router'
import type { User } from '@/types/user'
import type { Submission } from '@/types/submission'
import { Lock } from '@element-plus/icons-vue'

const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()
const loading = ref(true)
const currentTab = ref('submissions')

// 用户数据
const profileData = ref<User>({
  id: 0,
  username: '',
  email: '',
  avatar: '',
  bio: '',
  role: 'user',
  submissions: 0,
  acceptedProblems: 0,
  rating: 1500,
  createdAt: '',
  updatedAt: '',
})

// 判断是否为当前登录用户的主页
const isCurrentUser = computed(() => {
  const username = route.params.username
  return !username || username === userStore.userInfo?.username
})

// 添加提交记录相关的状态
const submissions = ref<Submission[]>([])
const submissionsTotal = ref(0)
const submissionsPage = ref(1)
const submissionsPageSize = ref(20)
const submissionsLoading = ref(false)

// 计算总页数
const submissionsTotalPages = computed(() =>
  Math.ceil(submissionsTotal.value / submissionsPageSize.value),
)

// 计算显示的页码范围
const submissionsDisplayedPages = computed(() => {
  const delta = 2 // 当前页前后显示的页数
  const range: number[] = []
  const rangeWithDots: (number | string)[] = []
  let l: number | undefined

  for (let i = 1; i <= submissionsTotalPages.value; i++) {
    if (
      i === 1 ||
      i === submissionsTotalPages.value ||
      (i >= submissionsPage.value - delta && i <= submissionsPage.value + delta)
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

// 获取用户数据
const fetchUserData = async () => {
  try {
    loading.value = true
    // 根据路由参数决定获取哪个用户的资料
    const username = route.params.username
    const url = username
      ? `/api/users/profile/${username}` // 获取指定用户的资料
      : '/api/user/profile' // 获取当前登录用户的资料

    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error('Failed to fetch user profile')
    }

    const data = await response.json()
    if (data.code === 200) {
      profileData.value = data.data
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    console.error('Failed to fetch user profile:', error)
    appStore.showNotification('error', '获取用户资料失败')
  } finally {
    loading.value = false
  }
}

// 添加新的处理函数
const handleAvatarUploadNotSupported = () => {
  appStore.showNotification('info', '暂不支持修改头像功能')
}

// 获取提交记录
const fetchSubmissions = async () => {
  try {
    submissionsLoading.value = true
    const queryParams = new URLSearchParams({
      page: submissionsPage.value.toString(),
      pageSize: submissionsPageSize.value.toString(),
      username: route.params.username?.toString() || userStore.userInfo?.username || '',
    })

    const response = await fetch(`/api/submissions?${queryParams.toString()}`, {
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
      submissions.value = data.data.submissions.map((submission: Partial<Submission>) => ({
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
      submissionsTotal.value = data.data.total
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    console.error('获取提交记录失败:', error)
    appStore.showNotification('error', '获取提交记录失败')
  } finally {
    submissionsLoading.value = false
  }
}

// 处理页码变化
const handleSubmissionsPageChange = (page: number) => {
  submissionsPage.value = page
  fetchSubmissions()
}

// 处理每页条数变化
const handleSubmissionsPageSizeChange = () => {
  submissionsPage.value = 1
  fetchSubmissions()
}

// 获取状态样式
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

// 监听路由参数变化
watch(
  () => route.params.username,
  () => {
    fetchUserData()
    if (currentTab.value === 'submissions') {
      fetchSubmissions()
    }
  },
)

onMounted(() => {
  fetchUserData()
  if (currentTab.value === 'submissions') {
    fetchSubmissions()
  }
})

// 监听标签页变化
watch(currentTab, (newTab: string) => {
  if (newTab === 'submissions') {
    fetchSubmissions()
  }
})

// 定义可用的标签页
const tabs = [
  { key: 'submissions', label: '提交记录' },
  { key: 'problems', label: '已解决题目' },
  { key: 'contests', label: '比赛记录' },
]
</script>

<style scoped>
.user-profile {
  max-width: 1200px;
  margin: 2rem auto;
  padding: 0 1rem;
  margin-top: 5rem; /* 为顶部导航栏留出空间 */
}

.profile-header {
  display: flex;
  gap: 2rem;
  padding: 2rem;
  background: var(--nav-bg-dark);
  border-radius: 8px;
  backdrop-filter: blur(10px);
}

.avatar-section {
  position: relative;
  width: 150px;
  height: 150px;
}

.avatar {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.upload-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.7);
  padding: 0.5rem;
  opacity: 0;
  transition: opacity 0.3s;
  border-bottom-left-radius: 50%;
  border-bottom-right-radius: 50%;
}

.avatar-section:hover .upload-overlay {
  opacity: 1;
}

.upload-btn {
  width: 100%;
  padding: 0.5rem;
  background: var(--primary-color);
  border: none;
  border-radius: 4px;
  color: white;
  cursor: pointer;
}

.hidden {
  display: none;
}

.user-info {
  flex: 1;
}

.user-info h2 {
  margin: 0;
  color: var(--text-light);
}

.bio {
  color: var(--text-gray);
  margin: 1rem 0;
}

.stats {
  display: flex;
  gap: 2rem;
  margin-top: 1rem;
}

.stat-item {
  text-align: center;
}

.number {
  display: block;
  font-size: 1.5rem;
  font-weight: bold;
  color: var(--primary-color);
}

.label {
  color: var(--text-gray);
  font-size: 0.9rem;
}

.tabs {
  margin: 2rem 0;
  display: flex;
  gap: 1rem;
  border-bottom: 1px solid var(--nav-bg-light);
}

.tab-btn {
  padding: 0.5rem 1rem;
  background: none;
  border: none;
  color: var(--text-light);
  cursor: pointer;
  position: relative;
  transition: color 0.3s;
}

.tab-btn:hover {
  color: var(--primary-color);
}

.tab-btn.active {
  color: var(--primary-color);
}

.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: -1px;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--primary-color);
}

.tab-content {
  min-height: 300px;
  padding: 1rem;
  background: var(--nav-bg-dark);
  border-radius: 8px;
  backdrop-filter: blur(10px);
}

@media (max-width: 768px) {
  .profile-header {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .stats {
    justify-content: center;
  }

  .tabs {
    flex-wrap: wrap;
    justify-content: center;
  }
}

/* 添加 loading 样式 */
.loading {
  text-align: center;
  padding: 2rem;
  color: var(--text-gray);
}

/* 添加新的样式 */
.locked-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  color: var(--text-gray);
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  backdrop-filter: blur(4px);
  border: 1px dashed rgba(255, 255, 255, 0.1);
}

.lock-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
  animation: float 3s ease-in-out infinite;
}

.locked-text {
  font-size: 1.2rem;
  font-weight: 500;
  margin: 0.5rem 0;
}

.locked-description {
  font-size: 0.9rem;
  opacity: 0.7;
}

@keyframes float {
  0% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
  100% {
    transform: translateY(0px);
  }
}

/* 添加提交记录表格样式 */
.submission-table {
  width: 100%;
  border-collapse: collapse;
  background: var(--bg-secondary);
  border-radius: 8px;
  overflow: hidden;
}

.submission-table th,
.submission-table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
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

/* System Error - 系统错误 */
.status-badge.system-error {
  background: linear-gradient(135deg, #cb356b, #bd3f32);
  box-shadow: 0 2px 8px rgba(203, 53, 107, 0.3);
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

/* 提交编号和代码长度徽章样式 */
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
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 悬停效果 */
.status-badge:hover,
.language-badge:hover,
.time-badge:hover,
.memory-badge:hover,
.id-badge:hover,
.code-length-badge:hover {
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
</style>
