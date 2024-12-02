<template>
  <div class="manage-contests">
    <div class="contests-header">
      <h1>角色管理</h1>
      <div class="filter-section">
        <div class="search-box">
          <input v-model="searchQuery" type="text" placeholder="搜索角色..." />
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

        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="handleEdit(row)"> 编辑 </el-button>
              <el-button type="warning" size="small" @click="handleParticipants(row)">
                参与者
              </el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)"> 删除 </el-button>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'

const router = useRouter()
const userStore = useUserStore()

// 状态变量
const contests = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchQuery = ref('')
const deleting = ref(false)

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
const handleEdit = (row: { id: string }) => {
  router.push(`/admin/contest/edit/${row.id}`)
}
// 查看参与者
const handleParticipants = (row: { id: string }) => {
  router.push(`/admin/contest/participants/${row.id}`)
}

// 删除比赛
const handleDelete = async (row: { title: { id: string }; id: string }) => {
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
</style>
