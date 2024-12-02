<template>
  <div class="participants">
    <h1>参与者列表</h1>

    <div class="participants-list">
      <table v-if="!loading">
        <thead>
          <tr>
            <th>用户</th>
            <th>状态</th>
            <th>注册时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="participant in participants" :key="participant.id">
            <td>
              <div class="user-info">
                <img
                  :src="participant.avatar || '/images/avatars/default-avatar.png'"
                  alt="avatar"
                  class="avatar"
                />
                <router-link :to="`/profile/${participant.username}`" class="username">
                  {{ participant.username }}
                </router-link>
              </div>
            </td>
            <td>
              <span :class="['status-badge', getStatusClass(participant.status)]">
                {{ getStatusText(participant.status) }}
              </span>
            </td>
            <td>{{ formatTime(participant.createdAt) }}</td>
          </tr>
        </tbody>
      </table>

      <div v-else class="loading">
        <div class="loading-spinner"></div>
        <div class="loading-text">加载中...</div>
      </div>
    </div>

    <div class="pagination">
      <div class="pagination-info">
        共 {{ total }} 名参与者
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
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const contestId = route.params.id as string

const participants = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const loading = ref(true)

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const displayedPages = computed(() => {
  const delta = 2
  const range = []
  const rangeWithDots = []
  let l

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

const fetchParticipants = async () => {
  try {
    loading.value = true
    const response = await fetch(
      `/api/contests/${contestId}/participants?page=${currentPage.value}&pageSize=${pageSize.value}`,
      {
        headers: {
          Authorization: `Bearer ${userStore.token}`,
        },
      },
    )

    if (!response.ok) {
      throw new Error('获取参与者列表失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      participants.value = data.data.participants
      total.value = data.data.total
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取参与者列表失败')
  } finally {
    loading.value = false
  }
}

const formatTime = (time: string) => {
  const date = new Date(time)
  return date.toLocaleString()
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'registered':
      return 'status-registered'
    case 'participating':
      return 'status-participating'
    default:
      return 'status-unknown'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'registered':
      return '已报名'
    case 'participating':
      return '参赛中'
    default:
      return '未知'
  }
}

const handlePageSizeChange = () => {
  currentPage.value = 1
  fetchParticipants()
}

const goToPage = (page: number) => {
  currentPage.value = page
  fetchParticipants()
}

onMounted(() => {
  if (!userStore.token) {
    router.push('/sign-in')
    return
  }
  fetchParticipants()
})
</script>

<style scoped>
.participants {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  margin-top: 64px;
}

.participants-list {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 1rem;
  margin-top: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

th {
  font-weight: 500;
  color: var(--primary-color);
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
}

.username {
  color: var(--text-primary);
  text-decoration: none;
  font-weight: 500;
}

.username:hover {
  color: var(--primary-color);
}

.status-badge {
  padding: 0.3rem 0.8rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-registered {
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  color: white;
}

.status-participating {
  background: linear-gradient(135deg, #43e97b, #38f9d7);
  color: white;
}

.status-unknown {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-light);
}

/* 分页样式保持不变 */
.pagination {
  margin-top: 2rem;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

/* ... 其他分页样式 ... */

/* 加载动画样式 */
.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
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
</style>
