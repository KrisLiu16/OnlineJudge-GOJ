<template>
  <div class="manage-contests">
    <div class="contests-header">
      <h1>比赛管理</h1>
      <div class="filter-section">
        <div class="search-box">
          <input v-model="searchQuery" type="text" placeholder="搜索比赛..." />
        </div>
        <div>
          <button @click="debounceSearch" class="page-btn">搜索</button>
        </div>
      </div>
    </div>

    <div class="contests-list">
      <el-table :data="contests" style="width: 100%" v-loading="loading">
        <el-table-column label="状态" width="80" fixed="left">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="id" label="编号" width="100" sortable fixed="left">
          <template #default="{ row }">
            <span class="contest-id">{{ row.id }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="title" label="标题" min-width="200">
          <template #default="{ row }">
            <router-link :to="`/contest/${row.id}`" class="contest-link">
              {{ row.title }}
            </router-link>
          </template>
        </el-table-column>

        <el-table-column label="时间" width="300">
          <template #default="{ row }">
            <span class="time-text">
              {{ new Date(row.startTime).toLocaleString() }} -
              {{ new Date(row.endTime).toLocaleString() }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="人数" width="100" prop="participantCount" sortable>
          <template #default="{ row }">
            <span class="count-text">{{ row.participantCount }}</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="handleEdit(row)">
                编辑
              </el-button>
              <el-button
                type="warning"
                size="small"
                @click="handleOpenSubmissions(row)"
                :disabled="!isContestEnded(row)"
                :loading="openingSubmissions === row.id"
              >
                开放记录
              </el-button>
              <el-button
                type="success"
                size="small"
                @click="handleUpdateRating(row)"
                :disabled="!isContestEnded(row)"
                :loading="updatingRating === row.id"
              >
                更新Rating
              </el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">
                删除
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
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

    <!-- 添加对话框组件 -->
    <el-dialog
      v-model="ratingDialogVisible"
      title="更新Rating"
      width="30%"
      :close-on-click-modal="false"
    >
      <el-form>
        <el-form-item label="排名方式">
          <el-select v-model="selectedRankType" placeholder="请选择排名方式">
            <el-option label="ACM模式" value="acm" />
            <el-option label="IOI模式" value="ioi" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="ratingDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmUpdateRating" :loading="updatingRating === currentContestId">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'

// 定义比赛接口
interface Contest {
  id: number
  title: string
  startTime: string
  endTime: string
  participantCount: number
  status: 'running' | 'not_started' | 'ended'
  problems: string
  description: string
}

const router = useRouter()
const userStore = useUserStore()

// 使用类型定义
const contests = ref<Contest[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchQuery = ref('')
const deleting = ref(false)
const updatingRating = ref<number | null>(null)
const ratingDialogVisible = ref(false)
const selectedRankType = ref('acm')
const currentContestId = ref<number | null>(null)
const openingSubmissions = ref<number | null>(null)

// 计算总页数
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 计算要显示的页码
const displayedPages = computed(() => {
  const range = []
  const start = Math.max(1, currentPage.value - 2)
  const end = Math.min(totalPages.value, start + 4)

  for (let i = start; i <= end; i++) {
    range.push(i)
  }
  return range
})

// 获取比赛列表
const fetchContests = async () => {
  loading.value = true
  try {
    const response = await fetch(
      `/api/contests?page=${currentPage.value}&pageSize=${pageSize.value}&search=${searchQuery.value}`,
      {
        headers: {
          Authorization: `Bearer ${userStore.token}`,
        },
      },
    )

    if (!response.ok) {
      throw new Error('获取比赛列表失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      contests.value = data.data.contests
      total.value = data.data.total
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取比赛列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索防抖
let searchTimeout: number
const debounceSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = window.setTimeout(() => {
    currentPage.value = 1
    fetchContests()
  }, 300)
}

// 获取状态类型
const getStatusType = (status: string) => {
  switch (status) {
    case 'running':
      return 'success'
    case 'not_started':
      return 'warning'
    default:
      return 'info'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'running':
      return '进行中'
    case 'not_started':
      return '未开始'
    default:
      return '已结束'
  }
}

// 编辑比赛
const handleEdit = (row: Contest) => {
  router.push(`/admin/contest/edit/${row.id}`)
}

// 删除比赛
const handleDelete = async (row: Contest) => {
  try {
    await ElMessageBox.confirm(`确定要删除比赛 "${row.title}" 吗？此操作不可恢复。`, '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })

    deleting.value = true
    const response = await fetch(`/api/contests/${row.id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('删除比赛失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      ElMessage.success('删除成功')
      fetchContests()
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : '删除比赛失败')
    }
  } finally {
    deleting.value = false
  }
}

// 跳转到指定页
const goToPage = (page: number) => {
  currentPage.value = page
  fetchContests()
}

// 修改每页显示数量
const handlePageSizeChange = () => {
  currentPage.value = 1
  fetchContests()
}

// 检查比赛是否已结束
const isContestEnded = (contest: Contest) => {
  return new Date(contest.endTime) < new Date()
}

// 处理更新Rating
const handleUpdateRating = (row: Contest) => {
  currentContestId.value = row.id
  selectedRankType.value = 'acm' // 重置为默认值
  ratingDialogVisible.value = true
}

// 添加确认更新的函数
const confirmUpdateRating = async () => {
  if (!currentContestId.value) return

  try {
    updatingRating.value = currentContestId.value
    const response = await fetch(
      `/api/admin/contest/${currentContestId.value}/update-rating?type=${selectedRankType.value}`,
      {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${userStore.token}`,
          'Content-Type': 'application/json'
        }
      }
    )

    if (!response.ok) {
      throw new Error('更新Rating失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      ElMessage.success('Rating更新成功')
      ratingDialogVisible.value = false
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '更新Rating失败')
  } finally {
    updatingRating.value = null
  }
}

// 添加开放记录的处理函数
const handleOpenSubmissions = async (row: Contest) => {
  try {
    await ElMessageBox.confirm(
      `确定要开放比赛 "${row.title}" 的提交记录吗？开放后所有用户都可以查看。`,
      '确认开放',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    openingSubmissions.value = row.id
    const response = await fetch(`/api/admin/contests/${row.id}/open-submissions`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error('开放提交记录失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      ElMessage.success('提交记录已开放')
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : '开放提交记录失败')
    }
  } finally {
    openingSubmissions.value = null
  }
}

onMounted(() => {
  fetchContests()
})
</script>

<style scoped>
.manage-contests {
  padding: 20px;
}

.contests-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filter-section {
  display: flex;
  gap: 16px;
  align-items: center;
}

.search-box input {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-light);
}

.contests-list {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 16px;
  margin-bottom: 20px;
  overflow-x: auto;
}

.contest-link {
  color: var(--el-color-primary);
  text-decoration: none;
}

.contest-link:hover {
  text-decoration: underline;
}

/* 分页样式 */
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

/* Element Plus 表格样式覆盖 */
:deep(.el-table) {
  background: transparent !important;
}

:deep(.el-table tr) {
  background: transparent !important;
}

:deep(.el-table th) {
  background: rgba(255, 255, 255, 0.05) !important;
}

:deep(.el-table td) {
  background: transparent !important;
}

/* 比赛编号样式 */
.contest-id {
  color: #ffffff;
  font-weight: bold;
  font-size: 1.1em;
}

/* 标签样式 */
:deep(.el-tag) {
  background: rgba(var(--primary-color-rgb), 0.1) !important;
  border: 1px solid rgba(var(--primary-color-rgb), 0.2) !important;
  color: var(--primary-color) !important;
  border-radius: 12px !important;
  padding: 2px 10px !important;
  margin: 2px 4px !important;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  transition: all 0.3s ease;
}

:deep(.el-tag:hover) {
  background: rgba(var(--primary-color-rgb), 0.2) !important;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(var(--primary-color-rgb), 0.2);
}

/* 不同状态标签的颜色 */
:deep(.el-tag--success) {
  background: rgba(103, 194, 58, 0.1) !important;
  border-color: rgba(103, 194, 58, 0.2) !important;
  color: #67c23a !important;
}

:deep(.el-tag--warning) {
  background: rgba(230, 162, 60, 0.1) !important;
  border-color: rgba(230, 162, 60, 0.2) !important;
  color: #e6a23c !important;
}

:deep(.el-tag--info) {
  background: rgba(144, 147, 153, 0.1) !important;
  border-color: rgba(144, 147, 153, 0.2) !important;
  color: #909399 !important;
}

/* 调整表格内按钮组的样式 */
:deep(.el-button-group) {
  display: flex;
  gap: 4px;
}

:deep(.el-button-group .el-button) {
  margin-left: 0 !important;
}

/* 确保固定列的背景色正确 */
:deep(.el-table__fixed-right) {
  height: 100% !important;
  background: transparent !important;
}

:deep(.el-table__fixed-right-patch) {
  background: transparent !important;
}

:deep(.el-table__fixed-right .el-table__fixed-header-wrapper) {
  background: rgba(255, 255, 255, 0.05) !important;
}

:deep(.el-table__fixed-right .el-table__fixed-body-wrapper) {
  background: transparent !important;
}

/* 添加时间和人数文本样式 */
.time-text,
.count-text {
  color: var(--text-light);
}

/* 添加Rating按钮的特殊样式 */
:deep(.el-button--success) {
  background: rgba(103, 194, 58, 0.1) !important;
  border-color: rgba(103, 194, 58, 0.2) !important;
  color: #67c23a !important;
}

:deep(.el-button--success:hover) {
  background: rgba(103, 194, 58, 0.2) !important;
  color: #fff !important;
}

:deep(.el-button--success:disabled) {
  background: rgba(103, 194, 58, 0.05) !important;
  border-color: rgba(103, 194, 58, 0.1) !important;
  color: rgba(103, 194, 58, 0.5) !important;
  cursor: not-allowed;
}

/* 调整按钮组的间距 */
:deep(.el-button-group .el-button) {
  margin-right: 4px !important;
}

:deep(.el-button-group .el-button:last-child) {
  margin-right: 0 !important;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 添加开放记录按钮的样式 */
:deep(.el-button--warning) {
  background: rgba(230, 162, 60, 0.1) !important;
  border-color: rgba(230, 162, 60, 0.2) !important;
  color: #e6a23c !important;
}

:deep(.el-button--warning:hover) {
  background: rgba(230, 162, 60, 0.2) !important;
  color: #fff !important;
}

:deep(.el-button--warning:disabled) {
  background: rgba(230, 162, 60, 0.05) !important;
  border-color: rgba(230, 162, 60, 0.1) !important;
  color: rgba(230, 162, 60, 0.5) !important;
  cursor: not-allowed;
}
</style>
