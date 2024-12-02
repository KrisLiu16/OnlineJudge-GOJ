<template>
  <div v-if="loading" class="loading-overlay">
    <div class="loading-spinner"></div>
    <div class="loading-text">加载中...</div>
  </div>

  <div class="contest-detail" v-show="!loading" :class="{ 'fade-in': !loading }">
    <div class="content-wrapper">
      <div class="left-section">
        <div class="contest-header">
          <div class="header-main">
            <h1>{{ contest.title }}</h1>
          </div>
          <div class="header-actions">
            <div class="contest-status" :class="getStatusClass(contest)">
              {{ getStatusText(contest) }}
            </div>
            <router-link
              :to="`/contest/${contestId}/rank`"
              target="_blank"
              class="rank-button"
            >
              <i class="fas fa-trophy"></i>
              排行榜
            </router-link>
          </div>
          <div class="contest-info">
            <div class="time-section">
              <div class="time-item">
                <span class="time-label">开始时间：</span>
                <span class="time-value">{{ new Date(contest.startTime).toLocaleString() }}</span>
              </div>
              <div class="time-item">
                <span class="time-label">结束时间：</span>
                <span class="time-value">{{ new Date(contest.endTime).toLocaleString() }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="problems-section">
          <h2>比赛题目</h2>
          <div class="problems-list">
            <table class="problems-table">
              <thead>
                <tr>
                  <th>状态</th>
                  <th>题号</th>
                  <th>标题</th>
                  <th>通过率</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(problem, index) in problems" :key="problem.id">
                  <td>{{ getProblemStatus(problem.status) }}</td>
                  <td>{{ String.fromCharCode(65 + index) }}</td>
                  <td>
                    <router-link
                      :to="`/contest/${contestId}/problem/${String.fromCharCode(65 + index)}`"
                      class="problem-link"
                      :class="{ disabled: !canAccessProblem }"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {{ problem.title }}
                    </router-link>
                  </td>
                  <td>
                    {{ problem.acceptedCount }}/{{ problem.submissionCount }} ({{
                      calculateAcceptanceRate(problem)
                    }}%)
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div class="right-section">
        <div class="contest-description markdown-body" v-html="renderedDescription"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { ElMessage } from 'element-plus'
import { marked } from 'marked'
import { getContestStatus } from '@/api/contests'

// 添加接口定义
interface Contest {
  id: string
  title: string
  description: string
  startTime: string
  endTime: string
  status: string
  participantCount: number
  problems: string
}

interface Problem {
  id: string
  title: string
  status: string
  acceptedCount: number
  submissionCount: number
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const contestId = route.params.id as string

const contest = ref<Contest>({
  id: '',
  title: '',
  description: '',
  startTime: '',
  endTime: '',
  status: '',
  participantCount: 0,
  problems: '',
})

const problems = ref<Problem[]>([])
const loading = ref(false)

const renderedDescription = computed(() => {
  try {
    return marked(contest.value.description || '')
  } catch (error) {
    console.error('Markdown渲染错误:', error)
    return '渲染错误'
  }
})

const canAccessProblem = computed(() => {
  const status = getContestStatus(contest.value.startTime, contest.value.endTime)
  return status === 'running' || status === 'ended'
})

// 修改函数签名，使用 Contest 接口
const getStatusText = (contest: Contest) => {
  const status = getContestStatus(contest.startTime, contest.endTime)
  switch (status) {
    case 'running':
      return '进行中'
    case 'not_started':
      return '即将开始'
    default:
      return '已结束'
  }
}

const getStatusClass = (contest: Contest) => {
  const status = getContestStatus(contest.startTime, contest.endTime)
  switch (status) {
    case 'running':
      return 'status-ongoing'
    case 'not_started':
      return 'status-upcoming'
    default:
      return 'status-ended'
  }
}

const calculateAcceptanceRate = (problem: Problem) => {
  if (!problem.submissionCount) return 0
  return ((problem.acceptedCount / problem.submissionCount) * 100).toFixed(1)
}

const fetchContestDetail = async () => {
  loading.value = true
  try {
    const response = await fetch(`/api/contests/${contestId}`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('获取比赛信息失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      contest.value = data.data
      // 获取题目详情
      await fetchProblems()
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取比赛信息失败')
    router.push('/contests')
  } finally {
    loading.value = false
  }
}

const fetchProblems = async () => {
  if (!contest.value.problems) return

  try {
    const response = await fetch(`/api/contests/problems/${contestId}`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('获取题目列表失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      problems.value = data.data.problems
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取题目列表失败')
  }
}

const getProblemStatus = (status: string) => {
  switch (status) {
    case 'accepted':
      return '🟢' // 已通过
    case 'attempted':
      return '🔵' // 已尝试但未通过
    default:
      return '⚪' // 未尝试
  }
}

onMounted(() => {
  if (!userStore.token) {
    router.push('/sign-in')
    return
  }
  fetchContestDetail()
})
</script>

<style scoped>
.contest-detail {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  margin-top: 64px;
}

.content-wrapper {
  display: flex;
  gap: 2rem;
  align-items: flex-start;
}

.left-section {
  flex: 1;
  min-width: 0; /* 防止flex子项溢出 */
}

.right-section {
  width: 300px;
  flex-shrink: 0;
}

.contest-header {
  background: rgba(255, 255, 255, 0.05);
  padding: 2rem;
  border-radius: 12px;
  margin-bottom: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.header-actions {
  display: flex;
  gap: 1rem;
  align-items: center;
  margin: 1rem 0;
}

.contest-status {
  position: static;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-weight: 500;
}

.rank-button {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: linear-gradient(135deg, #ffd700, #ffa500);
  color: #000;
  border-radius: 20px;
  text-decoration: none;
  font-weight: 500;
  transition: all 0.3s ease;
}

.contest-description {
  background: rgba(255, 255, 255, 0.05);
  padding: 2rem;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  position: sticky;
  top: 80px; /* 与顶部的距离 */
}

.problems-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 1rem;
}

.problems-table th,
.problems-table td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.problems-table th {
  font-weight: 500;
  color: #006996;
}

.problem-link {
  color: var(--primary-color);
  text-decoration: none;
  padding: 0.3rem 0.6rem;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.problem-link:hover:not(.disabled) {
  background: rgba(0, 105, 150, 0.1);
  transform: translateY(-1px);
}

.problem-link.disabled {
  color: #666;
  cursor: not-allowed;
  pointer-events: none;
  opacity: 0.7;
}

.participant-info {
  color: var(--text-light);
  font-size: 0.9rem;
}

.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--bg-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s linear infinite;
}

.loading-text {
  margin-top: 1rem;
  color: var(--text-light);
  font-size: 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.fade-in {
  animation: fadeIn 0.5s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.header-main {
  margin-bottom: 1rem;
}

.header-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1rem;
}

.rank-button {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: linear-gradient(135deg, #ffd700, #ffa500);
  color: #000;
  border-radius: 20px;
  text-decoration: none;
  font-weight: 500;
  transition: all 0.3s ease;
}

.rank-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 215, 0, 0.3);
}

.rank-button i {
  font-size: 1.1em;
}

/* 确保在小屏幕上正确换行 */
@media (max-width: 1200px) {
  .content-wrapper {
    flex-direction: column;
  }

  .right-section {
    width: 100%;
  }

  .contest-description {
    position: static;
  }
}
</style>
