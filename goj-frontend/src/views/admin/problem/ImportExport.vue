<template>
  <div class="import-export">
    <div class="header-panel">
      <h1>题目导入导出</h1>
      <div class="tip-panel">
        <el-alert title="导入导出说明" type="info" :closable="false" show-icon>
          <p>支持导入/导出题目的 ZIP 格式文件</p>
          <p>每个题目将被打包为单独的 ZIP 文件</p>
          <p>包含题目数据、测试用例和其他相关文件</p>
        </el-alert>
      </div>
    </div>

    <div class="action-panel">
      <!-- 导入部分 -->
      <div class="import-section">
        <h3>导入题目</h3>
        <el-upload
          class="upload-area"
          action="/api/admin/problems/import"
          :headers="uploadHeaders"
          :before-upload="beforeUpload"
          :on-success="handleUploadSuccess"
          :on-error="handleUploadError"
          :show-file-list="false"
          accept=".zip"
        >
          <el-button type="primary">
            <el-icon><Upload /></el-icon>
            选择文件
          </el-button>
          <template #tip>
            <div class="upload-tip">仅支持 .zip 格式</div>
          </template>
        </el-upload>
      </div>

      <!-- 导出部分 -->
      <div class="export-section">
        <h3>导出题目</h3>
        <div class="export-options">
          <el-checkbox v-model="exportOptions.includeTestcases">包含测试数据</el-checkbox>
          <el-checkbox v-model="exportOptions.includeSubmissions">包含提交记录</el-checkbox>
        </div>
        <div class="export-buttons">
          <el-button type="primary" @click="exportSelected" :disabled="!selectedProblems.length">
            <el-icon><Download /></el-icon>
            导出选中
          </el-button>
          <el-button type="success" @click="exportAll">
            <el-icon><Download /></el-icon>
            导出全部
          </el-button>
        </div>
      </div>
    </div>

    <div class="problem-list-panel">
      <el-table
        v-loading="loading"
        :data="problems"
        @selection-change="handleSelectionChange"
        class="problem-table"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column label="ID" prop="id" width="80" />
        <el-table-column label="标题" prop="title" />
        <el-table-column label="难度" width="100">
          <template #default="{ row }">
            <el-rate
              v-model="row.difficulty"
              :max="5"
              disabled
              text-color="#ff9900"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="exportProblem(row)">
              导出
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/modules/user'
import { Download, Upload } from '@element-plus/icons-vue'

const userStore = useUserStore()
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const problems = ref<Problem[]>([])
const selectedProblems = ref<Problem[]>([])

const exportOptions = ref({
  includeTestcases: true,
  includeSubmissions: false,
})

// 首先定义问题的接口
interface Problem {
  id: number
  title: string
  difficulty: number
}

// 上传相关配置
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${userStore.token}`,
}))

// 获取题目列表
const fetchProblems = async () => {
  loading.value = true
  try {
    const response = await fetch(
      `/api/admin/problems?page=${currentPage.value}&pageSize=${pageSize.value}`,
      {
        headers: {
          Authorization: `Bearer ${userStore.token}`,
        },
      },
    )
    if (!response.ok) throw new Error('获取题目列表失败')

    const data = await response.json()
    if (data.code === 200) {
      problems.value = data.data.problems
      total.value = data.data.total
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取题目列表失败')
  } finally {
    loading.value = false
  }
}

// 上传前检查
const beforeUpload = (file: File) => {
  const isZip = file.type === 'application/zip' || file.name.endsWith('.zip')
  if (!isZip) {
    ElMessage.error('只能上传 ZIP 格式的文件')
    return false
  }
  return true
}

// 上传成功
interface UploadResponse {
  code: number
  message?: string
  data?: {
    problems?: Problem[]
    total?: number
  }
}

const handleUploadSuccess = (response: UploadResponse) => {
  if (response.code === 200) {
    ElMessage.success('导入成功')
    fetchProblems()
  } else {
    ElMessage.error(response.message || '导入失败')
  }
}

// 上传失败
const handleUploadError = (error: Error) => {
  console.error('Upload error:', error)
  ElMessage.error('导入失败')
}

// 选择变化
const handleSelectionChange = (selection: Problem[]) => {
  selectedProblems.value = selection
}

// 导出单个题目
const exportProblem = async (problem: Problem) => {
  try {
    const response = await fetch(`/api/admin/problems/${problem.id}/export`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify(exportOptions.value),
    })

    if (!response.ok) throw new Error('导出失败')

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `problem_${problem.id}.zip`
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '导出失败')
  }
}

// 导出选中题目
const exportSelected = async () => {
  if (!selectedProblems.value.length) return

  try {
    const response = await fetch('/api/admin/problems/export-batch', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify({
        problemIds: selectedProblems.value.map((p: Problem) => p.id),
        ...exportOptions.value
      }),
    })

    if (!response.ok) throw new Error('导出失败')

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.download = `problems_${Date.now()}.zip`
    a.href = url
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '导出失败')
  }
}

// 导出全部题目
const exportAll = async () => {
  try {
    const response = await fetch('/api/admin/problems/export-all', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify(exportOptions.value),
    })

    if (!response.ok) throw new Error('导出失败')

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `all_problems_${Date.now()}.zip`
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '导出失败')
  }
}

// 分页处理
const handleSizeChange = (val: number) => {
  pageSize.value = val
  fetchProblems()
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
  fetchProblems()
}

// 初始化
fetchProblems()
</script>

<style scoped>
.import-export {
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
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.import-section,
.export-section {
  padding: 1rem;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.import-section h3,
.export-section h3 {
  margin-top: 0;
  margin-bottom: 1rem;
  color: var(--text-light);
}

.upload-tip {
  margin-top: 0.5rem;
  color: var(--text-gray);
  font-size: 0.9rem;
}

.export-options {
  margin-bottom: 1rem;
}

.export-buttons {
  display: flex;
  gap: 1rem;
}

.problem-list-panel {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 1rem;
}

:deep(.el-table) {
  background: transparent !important;
  color: white !important;
}

:deep(.el-table tr) {
  background: transparent !important;
}

:deep(.el-table th) {
  background: rgba(255, 255, 255, 0.05) !important;
  color: white !important;
}

:deep(.el-table td) {
  background: transparent !important;
  color: white !important;
}

.pagination {
  margin-top: 1rem;
  display: flex;
  justify-content: center;
}

:deep(.el-alert) {
  background: transparent !important;
  border: 1px solid rgba(255, 255, 255, 0.1);
}
</style>
