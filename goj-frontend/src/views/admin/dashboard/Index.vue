<template>
  <div class="dashboard">
    <div class="welcome-panel">
      <div class="welcome-content">
        <img src="/images/logo/GOJ_LOGO.svg" alt="logo" class="logo" />
        <h1 class="welcome-title">欢迎你，{{ username }}</h1>
        <div class="current-time">
          {{ currentTime }}
        </div>
      </div>
    </div>

    <div class="dashboard-stats">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card class="stat-card">
            <template #header>
              <div class="card-header">
                <el-icon><Document /></el-icon>
                <span>题目总数</span>
              </div>
            </template>
            <div class="card-value">{{ stats.basicStats?.problemCount || 0 }}</div>
          </el-card>
        </el-col>

        <el-col :span="8">
          <el-card class="stat-card">
            <template #header>
              <div class="card-header">
                <el-icon><User /></el-icon>
                <span>用户总数</span>
              </div>
            </template>
            <div class="card-value">{{ stats.basicStats?.userCount || 0 }}</div>
          </el-card>
        </el-col>

        <el-col :span="8">
          <el-card class="stat-card">
            <template #header>
              <div class="card-header">
                <el-icon><Timer /></el-icon>
                <span>今日提交</span>
              </div>
            </template>
            <div class="card-value">{{ stats.basicStats?.todaySubmissions || 0 }}</div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div class="system-stats">
      <h2 class="section-title">系统状态</h2>

      <el-row :gutter="20" class="stat-row">
        <el-col :span="12">
          <el-card class="stat-card">
            <template #header>
              <div class="card-header">
                <el-icon><Connection /></el-icon>
                <span>Go运行时</span>
              </div>
            </template>
            <div class="detail-list">
              <div class="detail-item">
                <span>协程数量:</span>
                <span>{{ stats.systemStats?.goRoutines || 0 }}</span>
              </div>
              <div class="detail-item">
                <span>堆对象数:</span>
                <span>{{ formatNumber(stats.systemStats?.heapObjects) }}</span>
              </div>
              <div class="detail-item">
                <span>堆内存:</span>
                <span>{{ formatBytes(stats.systemStats?.heapAlloc) }}</span>
              </div>
              <div class="detail-item">
                <span>栈内存:</span>
                <span>{{ formatBytes(stats.systemStats?.stackInUse) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :span="12">
          <el-card class="stat-card">
            <template #header>
              <div class="card-header">
                <el-icon><Box /></el-icon>
                <span>Redis状态</span>
                <el-button
                  type="primary"
                  size="small"
                  @click="clearCache"
                  :loading="clearingCache"
                  style="margin-left: auto"
                >
                  清除缓存
                </el-button>
              </div>
            </template>
            <div class="detail-list">
              <div class="detail-item">
                <span>键总数:</span>
                <span>{{ stats.systemStats?.redisKeyCount || 0 }}</span>
              </div>
              <div class="detail-item">
                <span>连接数:</span>
                <span>{{ stats.systemStats?.redisConnectedClients || 0 }}</span>
              </div>
              <div class="detail-item">
                <span>命中率:</span>
                <span>{{ (stats.systemStats?.redisHitRate || 0).toFixed(2) }}%</span>
              </div>
              <div class="detail-item">
                <span>过期键数:</span>
                <span>{{ stats.systemStats?.redisExpiredKeys || 0 }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="stat-row">
        <el-col :span="24">
          <el-card class="stat-card">
            <template #header>
              <div class="card-header">
                <el-icon><Monitor /></el-icon>
                <span>应用状态</span>
              </div>
            </template>
            <div class="detail-list horizontal">
              <div class="detail-item">
                <span>运行时间:</span>
                <span>{{ formatUptime(stats.systemStats?.uptimeSeconds) }}</span>
              </div>
              <div class="detail-item">
                <span>请求/分钟:</span>
                <span>{{ stats.systemStats?.requestsPerMin?.toFixed(2) || 0 }}</span>
              </div>
              <div class="detail-item">
                <span>错误率:</span>
                <span>{{ (stats.systemStats?.errorRate || 0).toFixed(2) }}%</span>
              </div>
              <div class="detail-item">
                <span>平均响应:</span>
                <span>{{ stats.systemStats?.averageResponse?.toFixed(2) || 0 }}ms</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import {
  Document,
  User,
  Timer,
  Monitor,
  Connection,
  List as DataBase,
  HomeFilled,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/modules/user'

const userStore = useUserStore()
const username = ref(userStore.userInfo?.username || 'Admin')
const currentTime = ref('')
let timer: number

// 格式化数字
const formatNumber = (num: number | undefined) => {
  if (!num) return '0'
  return num.toLocaleString()
}

// 格式化字节
const formatBytes = (bytes: number | undefined) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(2)} ${units[unitIndex]}`
}

// 格式化运行时间
const formatUptime = (seconds: number | undefined) => {
  if (!seconds) return '0秒'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${days}天${hours}小时${minutes}分钟`
}

// 统计数据
interface Stats {
  basicStats: {
    problemCount: number
    userCount: number
    todaySubmissions: number
  }
  systemStats: {
    goRoutines: number
    heapObjects: number
    heapAlloc: number
    stackInUse: number
    redisKeyCount: number
    redisConnectedClients: number
    uptimeSeconds: number
    requestsPerMin: number
    errorRate: number
    averageResponse: number
  }
}

const stats = ref<Stats>({
  basicStats: {
    problemCount: 0,
    userCount: 0,
    todaySubmissions: 0,
  },
  systemStats: {
    goRoutines: 0,
    heapObjects: 0,
    heapAlloc: 0,
    stackInUse: 0,
    redisKeyCount: 0,
    redisConnectedClients: 0,
    uptimeSeconds: 0,
    requestsPerMin: 0,
    errorRate: 0,
    averageResponse: 0,
  },
})

const updateTime = () => {
  const now = new Date()
  const options: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }
  currentTime.value = now.toLocaleString('zh-CN', options)
}

// 获取统计数据
const fetchStats = async () => {
  try {
    // 从 store 获取 token
    const token = userStore.token
    if (!token) {
      throw new Error('未登录或 token 已过期')
    }

    const response = await fetch('/api/admin/stats', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.message || '获取统计数据失败')
    }

    const result = await response.json()
    if (result.code === 200) {
      stats.value = result.data
    } else {
      throw new Error(result.message || '获取统计数据失败')
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
    ElMessage.error(error instanceof Error ? error.message : '获取统计数据失败')
  }
}

// 定时刷新统计数据
let statsTimer: number

// 添加清除缓存相关的状态
const clearingCache = ref(false)

// 添加清除缓存的方法
const clearCache = async () => {
  try {
    clearingCache.value = true
    const token = userStore.token
    if (!token) {
      throw new Error('未登录或 token 已过期')
    }

    const response = await fetch('/api/admin/cache/clear', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.message || '清除缓存失败')
    }

    const result = await response.json()
    if (result.code === 200) {
      ElMessage.success({
        message: '缓存已清除',
        duration: 2000,
      })
      // 短暂延迟后刷新统计数据
      setTimeout(async () => {
        await fetchStats()
      }, 500)
    } else {
      throw new Error(result.message || '清除缓存失败')
    }
  } catch (error) {
    console.error('清除缓存失败:', error)
    ElMessage.error({
      message: error instanceof Error ? error.message : '清除缓存失败',
      duration: 3000,
    })
  } finally {
    clearingCache.value = false
  }
}

onMounted(() => {
  updateTime()
  timer = window.setInterval(updateTime, 1000)

  // 首次获取统计数据
  fetchStats()
  // 每5分钟刷新一次统计数据
  statsTimer = window.setInterval(fetchStats, 5 * 60 * 1000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
  if (statsTimer) {
    clearInterval(statsTimer)
  }
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.welcome-panel {
  margin-bottom: 2rem;
  padding: 2rem;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.welcome-content {
  text-align: center;
}

.welcome-title {
  font-size: 2.5rem;
  font-weight: bold;
  margin-bottom: 1rem;
  background: linear-gradient(45deg, var(--primary-color), #006996);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

.current-time {
  font-size: 1.2rem;
  color: var(--text-gray);
}

.dashboard-stats {
  margin-top: 2rem;
}

.stat-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.1rem;
  color: var(--text-light);
  justify-content: space-between;
}

.card-value {
  font-size: 2rem;
  font-weight: bold;
  text-align: center;
  padding: 1rem 0;
  background: linear-gradient(45deg, var(--primary-color), #006996);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

:deep(.el-card) {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-light);
}

:deep(.el-card__header) {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.system-stats {
  margin-top: 2rem;
}

.section-title {
  font-size: 1.5rem;
  margin-bottom: 1rem;
  color: var(--text-light);
}

.subsection-title {
  font-size: 1.2rem;
  margin: 1.5rem 0 1rem;
  color: var(--text-light);
}

.stat-row {
  margin-bottom: 1.5rem;
}

.detail-list {
  padding: 0.5rem;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  color: var(--text-light);
}

:deep(.el-progress-dashboard__text) {
  color: var(--text-light) !important;
}

.detail-list.horizontal {
  display: flex;
  justify-content: space-around;
  flex-wrap: wrap;
}

.detail-list.horizontal .detail-item {
  flex: 1;
  min-width: 200px;
  text-align: center;
  margin: 10px;
}
</style>
