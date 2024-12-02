<template>
  <div class="manage-data">
    <div class="header-panel">
      <h1>测试数据管理 - 题目 {{ problemId }}</h1>
      <div class="tip-panel">
        <el-alert title="数据格式说明" type="info" :closable="false" show-icon>
          <p>输入文件后缀为 .in，输出文件后缀为 .out</p>
          <p>前缀名相同的输入输出文件会被视为一组测试数据</p>
          <p>例如：test1.in 和 test1.out 为一组测试数据</p>
        </el-alert>
      </div>
    </div>

    <div class="action-panel">
      <el-upload
        class="upload-area"
        action="/api/admin/problem/data/upload"
        :headers="uploadHeaders"
        :data="{ problemId }"
        multiple
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
        :before-upload="beforeUpload"
        :show-file-list="false"
        name="file"
      >
        <el-button type="primary">
          <el-icon><Upload /></el-icon>
          上传文件
        </el-button>
      </el-upload>

      <el-button type="success" @click="downloadAll">
        <el-icon><Download /></el-icon>
        下载全部
      </el-button>

      <el-button type="danger" @click="deleteSelected" :disabled="!selectedFiles.length">
        <el-icon><Delete /></el-icon>
        删除选中
      </el-button>
    </div>

    <div class="file-list-panel">
      <el-table
        v-loading="loading"
        :data="files"
        @selection-change="handleSelectionChange"
        class="file-table"
      >
        <el-table-column type="selection" width="55" />

        <el-table-column label="文件名" prop="name" sortable>
          <template #default="{ row }">
            <div class="file-name" @click="previewFile(row)">
              <el-icon><Document /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="大小" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.size) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="downloadFile(row)"> 下载 </el-button>
              <el-button type="danger" size="small" @click="deleteFile(row)"> 删除 </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <div class="pagination-info">
          共 {{ total }} 条记录
          <select v-model="pageSize" @change="handlePageSizeChange" class="page-size-select">
            <option v-for="size in [20, 50, 100]" :key="size" :value="size">
              {{ size }} 条/页
            </option>
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

    <!-- 文件预览对话框 -->
    <el-dialog
      v-model="previewDialog.visible"
      :title="previewDialog.fileName"
      width="60%"
      class="preview-dialog"
    >
      <pre class="file-content">{{ previewDialog.content }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Upload, Download, Delete } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/modules/user'

const route = useRoute()
const userStore = useUserStore()
const problemId = route.params.id as string

// 状态变量
const loading = ref(false)
const files = ref<TestFile[]>([])
const selectedFiles = ref<TestFile[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 文件预览对话框
const previewDialog = ref({
  visible: false,
  fileName: '',
  content: '',
})

// 上传相关
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${userStore.token}`,
}))

// 添加类型定义
interface TestFile {
  name: string
  size: number
}

interface ApiResponse<T> {
  code: number
  message?: string
  data?: T
}

interface FileListResponse {
  files: TestFile[]
  total: number
}

interface FileContentResponse {
  content: string
}

interface UploadResponse {
  code: number
  message?: string
  data?: {
    files?: TestFile[]
    total?: number
  }
}

// 格式化文件大小
const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

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

// 获取文件列表
const fetchFiles = async () => {
  loading.value = true
  try {
    const response = await fetch(
      `/api/admin/problem/data/${problemId}?page=${currentPage.value}&pageSize=${pageSize.value}`,
      {
        headers: {
          Authorization: `Bearer ${userStore.token}`,
        },
      },
    )

    if (!response.ok) {
      throw new Error('获取文件列表失败')
    }

    const data = await response.json() as ApiResponse<FileListResponse>
    if (data.code === 200 && data.data) {
      files.value = data.data.files
      total.value = data.data.total
      selectedFiles.value = []
    } else {
      throw new Error(data.message || '获取文件列表失败')
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取文件列表失败')
  } finally {
    loading.value = false
  }
}

// 上传前检查
const beforeUpload = (file: File): boolean => {
  const isValidType = file.name.endsWith('.in') || file.name.endsWith('.out')
  if (!isValidType) {
    ElMessage.error('只能上传 .in 或 .out 后缀的文件')
    return false
  }
  return true
}

// 上传成功
const handleUploadSuccess = (response: UploadResponse): void => {
  if (response.code === 200) {
    ElMessage.success('上传成功')
    fetchFiles()
  } else {
    ElMessage.error(response.message || '上传失败')
  }
}

// 上传失败
const handleUploadError = (error: Error): void => {
  console.error('Upload error:', error)
  ElMessage.error(error.message || '上传失败')
}

// 选择文件变化
const handleSelectionChange = (selection: TestFile[]) => {
  selectedFiles.value = selection
}

// 预览文件
const previewFile = async (file: TestFile) => {
  try {
    const response = await fetch(`/api/admin/problem/data/${problemId}/${file.name}`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('获取文件内容失败')
    }

    const data = await response.json() as ApiResponse<FileContentResponse>
    if (data.code === 200 && data.data) {
      previewDialog.value = {
        visible: true,
        fileName: file.name,
        content: data.data.content,
      }
    } else {
      throw new Error(data.message || '获取文件内容失败')
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取文件内容失败')
  }
}

// 下载文件
const downloadFile = async (file: TestFile) => {
  try {
    const response = await fetch(`/api/admin/problem/data/${problemId}/${file.name}/download`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('下载文件失败')
    }

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = file.name
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '下载文件失败')
  }
}

// 下载全部
const downloadAll = async () => {
  try {
    const response = await fetch(`/api/admin/problem/data/${problemId}/download-all`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('下载文件失败')
    }

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `problem_${problemId}_data.zip`
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '下载文件失败')
  }
}

// 删除文件
const deleteFile = async (file: TestFile) => {
  try {
    await ElMessageBox.confirm(`确定要删除文件 "${file.name}" 吗？`, '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })

    const response = await fetch(`/api/admin/problem/data/${problemId}/${file.name}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('删除文件失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      ElMessage.success('删除成功')
      fetchFiles()
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : '删除文件失败')
    }
  }
}

// 删除选中
const deleteSelected = async () => {
  if (selectedFiles.value.length === 0) return

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedFiles.value.length} 个文件吗？`,
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    const fileNames = selectedFiles.value.map((file: TestFile) => file.name)
    const response = await fetch(`/api/admin/problem/data/${problemId}/batch`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ files: fileNames }),
    })

    if (!response.ok) {
      throw new Error('删除文件失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      ElMessage.success('删除成功')
      fetchFiles()
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : '删除文件失败')
    }
  }
}

// 跳转到指定页
const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchFiles() // 重新获取数据
}

// 修改每页显示数量
const handlePageSizeChange = () => {
  currentPage.value = 1 // 切换每页数量时重置到第一页
  fetchFiles() // 重新获取数据
}

// 确保在组件挂载时获取数据
onMounted(() => {
  fetchFiles()
})
</script>

<style scoped>
.manage-data {
  padding: 20px;
}

.header-panel {
  margin-bottom: 2rem;
}

.tip-panel {
  margin-top: 1rem;
  max-width: 800px;
}

.action-panel {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.file-list-panel {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 1rem;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: var(--el-color-primary);
}

.file-name:hover {
  text-decoration: underline;
}

.preview-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.file-content {
  padding: 1rem;
  margin: 0;
  background: #fff;
  border-radius: 4px;
  max-height: 60vh;
  overflow: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: monospace;
  color: #333;
  border: 1px solid #eee;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.1);
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

:deep(.el-dialog) {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

:deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(0, 0, 0, 0.1);
  color: #333;
}

:deep(.el-dialog__title) {
  color: #333;
}

:deep(.el-alert) {
  background: transparent !important;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* 添加分页样式 */
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

/* 添加滚动条样式 */
.file-content::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.file-content::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.file-content::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 4px;
}

.file-content::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>
