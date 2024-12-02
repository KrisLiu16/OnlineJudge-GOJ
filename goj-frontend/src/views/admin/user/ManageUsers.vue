<template>
  <div class="manage-users">
    <div class="users-header">
      <h1>用户管理</h1>
      <div class="filter-section">
        <el-button type="primary" @click="batchCreateDialogVisible = true">
          <el-icon><Plus /></el-icon>
          批量创建
        </el-button>
        <div class="search-box">
          <input v-model="searchQuery" type="text" placeholder="搜索用户..." />
        </div>
        <div>
          <button @click="debounceSearch" class="page-btn">搜索</button>
        </div>
      </div>
    </div>

    <div class="users-list">
      <el-table :data="users" style="width: 100%" v-loading="loading">
        <el-table-column prop="ID" label="ID" width="80" fixed="left" />

        <el-table-column prop="Username" label="用户名" min-width="120">
          <template #default="{ row }">
            <router-link :to="`/user/${row.ID}`" class="user-link">
              {{ row.Username }}
            </router-link>
          </template>
        </el-table-column>

        <el-table-column prop="Email" label="邮箱" min-width="200" />
        <el-table-column prop="Role" label="角色" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getRoleType(row.Role)">
              {{ getRoleText(row.Role) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="CreatedAt" label="注册时间" width="180">
          <template #default="{ row }">
            {{ new Date(row.CreatedAt).toLocaleString() }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button
                type="primary"
                size="small"
                @click="handleSetRole(row)"
                :disabled="row.Role === 'admin'"
              >
                设置角色
              </el-button>
              <el-button
                type="danger"
                size="small"
                @click="handleDelete(row)"
                :disabled="row.Role === 'admin'"
              >
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

    <el-dialog v-model="roleDialogVisible" title="设置用户角色" width="30%">
      <el-form :model="roleForm">
        <el-form-item label="角色">
          <el-select v-model="roleForm.role">
            <el-option label="普通用户" value="user" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="roleDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmSetRole">确定</el-button>
        </span>
      </template>
    </el-dialog>

    <el-dialog v-model="batchCreateDialogVisible" title="批量创建用户" width="30%">
      <el-form :model="batchCreateForm" label-width="100px">
        <el-form-item label="起始编号">
          <el-input-number
            v-model="batchCreateForm.startNumber"
            :min="1"
            :max="999999"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item label="创建数量">
          <el-input-number
            v-model="batchCreateForm.count"
            :min="1"
            :max="1000"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item label="用户名前缀">
          <el-input
            v-model="batchCreateForm.prefix"
            placeholder="可选，如: student_"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="batchCreateDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleBatchCreate" :loading="creating">
            创建
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/modules/user'
import { Plus } from '@element-plus/icons-vue'

const userStore = useUserStore()

// 添加类型定义
interface User {
  ID: number
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: null | string
  Username: string
  Email: string
  PasswordHash: string
  Avatar: string
  Bio: string
  Role: 'admin' | 'user'
  Submissions: number
  AcceptedProblems: number
  Rating: number
}

// 状态变量
const users = ref<User[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchQuery = ref('')
const roleDialogVisible = ref(false)
const roleForm = ref({
  userId: '',
  role: 'user',
})
const batchCreateDialogVisible = ref(false)
const creating = ref(false)
const batchCreateForm = ref({
  startNumber: 1,
  count: 50,
  prefix: '',
})
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
// 获取用户列表
const fetchUsers = async () => {
  loading.value = true
  try {
    const response = await fetch(
      `/api/admin/users?page=${currentPage.value}&pageSize=${pageSize.value}&search=${searchQuery.value}`,
      {
        headers: {
          Authorization: `Bearer ${userStore.token}`,
        },
      },
    )

    if (!response.ok) {
      throw new Error('获取用户列表失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      users.value = data.data.users
      total.value = data.data.total
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取用户表失败')
  } finally {
    loading.value = false
  }
}

// 搜索防抖
let searchTimeout: ReturnType<typeof setTimeout>
const debounceSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchUsers()
  }, 300)
}

// 获取角色类型
const getRoleType = (role: string) => {
  switch (role.toLowerCase()) {
    case 'admin':
      return 'danger'
    default:
      return 'info'
  }
}

// 获取角色文本
const getRoleText = (role: string) => {
  switch (role.toLowerCase()) {
    case 'admin':
      return '管理员'
    default:
      return '普通用户'
  }
}

// 设置角色
const handleSetRole = (row: User) => {
  roleForm.value.userId = row.ID.toString()
  roleForm.value.role = row.Role.toLowerCase()
  roleDialogVisible.value = true
}

// 确认设置角色
const confirmSetRole = async () => {
  try {
    const response = await fetch(`/api/admin/users/${roleForm.value.userId}/role`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify({ role: roleForm.value.role.toLowerCase() }),
    })

    if (!response.ok) {
      throw new Error('设置角色失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      ElMessage.success('设置成功')
      roleDialogVisible.value = false
      fetchUsers()
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '设置角色失败')
  }
}

// 删除用户
const handleDelete = async (row: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${row.Username}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    const response = await fetch(`/api/admin/users/${row.ID}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('删除用户失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      ElMessage.success('删除成功')
      fetchUsers()
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : '删除用户失败')
    }
  }
}

// 跳转到指定页
const goToPage = (page: number) => {
  currentPage.value = page
  fetchUsers()
}

// 修改每页显示数量
const handlePageSizeChange = () => {
  currentPage.value = 1
  fetchUsers()
}

// 批量创建用户
const handleBatchCreate = async () => {
  if (batchCreateForm.value.count <= 0 || batchCreateForm.value.startNumber <= 0) {
    ElMessage.error('请输入有效的数字')
    return
  }

  try {
    creating.value = true
    const response = await fetch('/api/admin/users/batch-create', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify(batchCreateForm.value),
    })

    // 这里我们期望直接获取文件流
    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.message || '创建失败')
    }

    // 获取文件名
    const contentDisposition = response.headers.get('Content-Disposition')
    const filename = contentDisposition
      ? contentDisposition.split('filename=')[1].replace(/"/g, '')
      : 'users.xlsx'

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

    ElMessage.success('批量创建成功，已下载用户信息表')
    batchCreateDialogVisible.value = false
    fetchUsers() // 刷新用户列表
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '批量创建失败')
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
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
.manage-users {
  padding: 20px;
}

.users-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.users-list {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 16px;
  margin-bottom: 20px;
  overflow-x: auto;
}

.user-link {
  color: var(--el-color-primary);
  text-decoration: none;
}

.user-link:hover {
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

/* 修改弹窗相关样式 */
:deep(.el-dialog) {
  background: #1e1e2d !important; /* 深色背景 */
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

:deep(.el-dialog__header) {
  padding: 20px 24px;
  margin-right: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

:deep(.el-dialog__title) {
  color: #e1e1e1 !important;
  font-size: 18px;
  font-weight: 600;
  line-height: 24px;
}

:deep(.el-dialog__body) {
  padding: 24px;
  color: #e1e1e1;
}

:deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

:deep(.el-form-item__label) {
  color: #e1e1e1 !important;
  font-weight: 500;
}

/* 输入框样式 */
:deep(.el-input__wrapper),
:deep(.el-input-number__wrapper) {
  background: rgba(255, 255, 255, 0.04) !important;
  box-shadow: none !important;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}

:deep(.el-input__wrapper:hover),
:deep(.el-input-number__wrapper:hover) {
  border-color: rgba(255, 255, 255, 0.2);
}

:deep(.el-input__wrapper.is-focus),
:deep(.el-input-number__wrapper.is-focus) {
  border-color: var(--el-color-primary) !important;
  box-shadow: 0 0 0 1px var(--el-color-primary) !important;
}

:deep(.el-input__inner) {
  color: #e1e1e1 !important;
  height: 36px;
  line-height: 36px;
}

/* 数字输入框的控制按钮 */
:deep(.el-input-number__decrease),
:deep(.el-input-number__increase) {
  background: transparent !important;
  border-color: rgba(255, 255, 255, 0.1) !important;
  color: #e1e1e1 !important;
}

:deep(.el-input-number__decrease:hover),
:deep(.el-input-number__increase:hover) {
  color: var(--el-color-primary) !important;
  background: rgba(255, 255, 255, 0.04) !important;
}

/* 按钮样式 */
:deep(.el-button) {
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.04);
  color: #e1e1e1;
  height: 36px;
  padding: 0 16px;
  font-weight: 500;
  transition: all 0.3s ease;
}

:deep(.el-button:hover) {
  border-color: rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.08);
}

:deep(.el-button--primary) {
  background: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: white;
}

:deep(.el-button--primary:hover) {
  background: var(--el-color-primary-light-3);
  border-color: var(--el-color-primary-light-3);
}

/* 表单项间距 */
:deep(.el-form-item) {
  margin-bottom: 24px;
}

/* 关闭按钮 */
:deep(.el-dialog__headerbtn .el-dialog__close) {
  color: #e1e1e1;
}

:deep(.el-dialog__headerbtn .el-dialog__close:hover) {
  color: var(--el-color-primary);
}
</style>
