<template>
  <div class="manage-problems">
    <div class="problems-header">
      <h1>题目管理</h1>
      <div class="filter-section">
        <div class="search-box">
          <input v-model="searchQuery" type="text" placeholder="搜索题目内容..." />
        </div>
        <div>
          <button @click="debounceSearch" class="page-btn">搜索</button>
        </div>
      </div>
    </div>

    <div class="problems-list">
      <el-table :data="problems" style="width: 100%" v-loading="loading">
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'success'" size="small">
              {{ row.role === 'admin' ? '管理' : '公开' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="id" label="题号" width="100" sortable>
          <template #default="{ row }">
            <span class="problem-id">{{ row.id }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="title" label="标题" min-width="200">
          <template #default="{ row }">
            <router-link :to="`/problem/${row.id}`" class="problem-link">
              {{ row.title.length > 48 ? row.title.substring(0, 48) + '...' : row.title }}
            </router-link>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="handleEdit(row)"> 编辑 </el-button>
              <el-button type="warning" size="small" @click="handleManageData(row)">
                管理数据
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

// 添加类型定义
interface Problem {
  id: number
  title: string
  role: 'admin' | 'public'
}

interface ApiResponse<T> {
  code: number
  message?: string
  data?: T
}

interface ProblemListResponse {
  problems: Problem[]
  total: number
}

// 状态变量
const problems = ref<Problem[]>([])
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

// 获取题目列表
const fetchProblems = async () => {
  loading.value = true
  try {
    const response = await fetch(
      `/api/admin/problems?page=${currentPage.value}&pageSize=${pageSize.value}&search=${searchQuery.value}`,
      {
        headers: {
          Authorization: `Bearer ${userStore.token}`,
        },
      },
    )

    if (!response.ok) {
      throw new Error('获取题目列表失败')
    }

    const data = await response.json() as ApiResponse<ProblemListResponse>
    if (data.code === 200 && data.data) {
      problems.value = data.data.problems
      total.value = data.data.total
    } else {
      throw new Error(data.message || '获取题目列表失败')
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取题目列表失败')
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
    fetchProblems()
  }, 300)
}

// 编辑题目
const handleEdit = (row: Problem) => {
  router.push(`/admin/problem/edit/${row.id}`)
}

// 删除题目
const handleDelete = async (row: Problem) => {
  try {
    await ElMessageBox.confirm(`确定要删除题目 "${row.title}" 吗？此操作不可恢复。`, '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })

    deleting.value = true
    const response = await fetch(`/api/admin/problems/${row.id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('删除题目失败')
    }

    const data = await response.json() as ApiResponse<void>
    if (data.code === 200) {
      ElMessage.success('删除成功')
      fetchProblems()
    } else {
      throw new Error(data.message || '删除题目失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : '删除题目失败')
    }
  } finally {
    deleting.value = false
  }
}

// 跳转到指定页
const goToPage = (page: number) => {
  currentPage.value = page
  fetchProblems()
}

// 修改每页显示数量
const handlePageSizeChange = () => {
  currentPage.value = 1
  fetchProblems()
}

// 管理数据
const handleManageData = (row: Problem) => {
  router.push(`/admin/problem/data/${row.id}`)
}

onMounted(() => {
  fetchProblems()
})
</script>

<style scoped>
.manage-problems {
  padding: 20px;
}

.problems-header {
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

.problems-list {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
}

.problem-link {
  color: var(--el-color-primary);
  text-decoration: none;
}

.problem-link:hover {
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

/* 添加题号样式 */
.problem-id {
  color: #ffffff;
  font-weight: bold;
  font-size: 1.1em;
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

/* 响应式调整 */
@media (max-width: 768px) {
  .pagination-buttons {
    flex-wrap: wrap;
    justify-content: center;
  }

  .page-btn {
    padding: 0.4rem 0.8rem;
    font-size: 0.9rem;
  }
}

/* 题目标签样式 */
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

/* 不同类型标签使用不同的颜色 */
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

:deep(.el-tag--danger) {
  background: rgba(245, 108, 108, 0.1) !important;
  border-color: rgba(245, 108, 108, 0.2) !important;
  color: #f56c6c !important;
}

:deep(.el-tag--info) {
  background: rgba(144, 147, 153, 0.1) !important;
  border-color: rgba(144, 147, 153, 0.2) !important;
  color: #909399 !important;
}
</style>
