<template>
  <div class="contest-edit">
    <div class="edit-container">
      <div class="edit-header">
        <h1>编辑比赛</h1>
        <div class="header-actions">
          <el-button type="primary" @click="save" :loading="isSaving">
            {{ isSaving ? '保存中...' : '保存修改' }}
          </el-button>
        </div>
      </div>

      <div class="edit-form" v-loading="loading">
        <div class="form-row">
          <div class="form-group title-input">
            <el-input v-model="formData.title" placeholder="输入比赛标题..." />
          </div>
          <div class="form-group">
            <el-select v-model="formData.role" placeholder="选择权限">
              <el-option value="public" label="公开" />
              <el-option value="private" label="私有" />
            </el-select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <el-date-picker
              v-model="startDate"
              type="date"
              placeholder="选择开始日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
            />
          </div>
          <div class="form-group">
            <el-time-picker
              v-model="startTime"
              placeholder="选择开始时间"
              format="HH:mm"
              value-format="HH:mm"
            />
          </div>
          <div class="form-group">
            <el-date-picker
              v-model="endDate"
              type="date"
              placeholder="选择结束日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
            />
          </div>
          <div class="form-group">
            <el-time-picker
              v-model="endTime"
              placeholder="选择结束时间"
              format="HH:mm"
              value-format="HH:mm"
            />
          </div>
        </div>

        <div class="form-row full-width">
          <div class="problems-input">
            <div class="input-row">
              <el-input
                v-model="newProblem"
                placeholder="输入题目ID (支持单个ID如10001或范围如10001-10005)"
                @keyup.enter="handleAddProblem"
              >
                <template #append>
                  <el-button @click="handleAddProblem">添加</el-button>
                </template>
              </el-input>
              <el-button type="danger" @click="clearProblems" :disabled="!formData.problems.length">
                清空全部
              </el-button>
            </div>
            <div class="problems-container">
              <el-tag
                v-for="problem in formData.problems"
                :key="problem"
                closable
                @close="handleRemoveProblem(problem)"
                class="problem-item"
              >
                {{ problem }}
              </el-tag>
            </div>
          </div>
        </div>

        <div class="editor-container">
          <div class="editor-toolbar">
            <el-radio-group v-model="currentTab" size="large">
              <el-radio-button label="编辑">编辑</el-radio-button>
              <el-radio-button label="预览">预览</el-radio-button>
            </el-radio-group>
          </div>

          <div class="editor-content">
            <el-input
              v-if="currentTab === '编辑'"
              v-model="formData.description"
              type="textarea"
              :rows="15"
              placeholder="使用 Markdown 编写比赛说明..."
            />
            <div v-else class="markdown-preview markdown-body" v-html="renderedContent"></div>
          </div>
        </div>
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

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const currentTab = ref('编辑')
const loading = ref(false)
const isSaving = ref(false)

// 表单数据
const formData = ref({
  title: '',
  description: '',
  startTime: '',
  endTime: '',
  role: 'public',
  problems: [] as string[],
})

// Markdown预览
const renderedContent = computed(() => {
  try {
    return marked(formData.value.description || '')
  } catch (error) {
    console.error('Markdown渲染错误:', error)
    return '渲染错误'
  }
})

// 添加时间相关的 ref
const startDate = ref('')
const startTime = ref('')
const endDate = ref('')
const endTime = ref('')

// 计算完整的开始和结束时间
const computeDateTime = () => {
  // 创建带时区的ISO字符串
  const startDateTime = new Date(`${startDate.value}T${startTime.value}:00`)
  const endDateTime = new Date(`${endDate.value}T${endTime.value}:00`)

  formData.value.startTime = startDateTime.toISOString()
  formData.value.endTime = endDateTime.toISOString()
}

// 解析时间到日期和时间字段
const parseDateTime = (dateTimeStr: string) => {
  const dateTime = new Date(dateTimeStr)
  return {
    date: dateTime.toISOString().split('T')[0],
    time: dateTime.toLocaleTimeString('en-US', {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
    }),
  }
}

// 获取比赛数据
const fetchContest = async () => {
  const contestId = route.params.id
  loading.value = true
  try {
    const response = await fetch(`/api/contests/${contestId}`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('获取比赛数据失败')
    }

    const data = await response.json()
    if (data.code === 200) {
      const contest = data.data
      formData.value = {
        ...contest,
        problems: contest.problems.split(','),
      }

      // 解析开始和结束时间
      const start = parseDateTime(contest.startTime)
      const end = parseDateTime(contest.endTime)
      startDate.value = start.date
      startTime.value = start.time
      endDate.value = end.date
      endTime.value = end.time
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取比赛数据失败')
    router.push('/admin/contest/manage')
  } finally {
    loading.value = false
  }
}

// 保存比赛
const save = async () => {
  if (isSaving.value) return
  if (!formData.value.title || !formData.value.description) {
    ElMessage.error('标题和说明不能为空')
    return
  }

  if (!startDate.value || !startTime.value || !endDate.value || !endTime.value) {
    ElMessage.error('请设置比赛时间')
    return
  }

  // 计算完整时间
  computeDateTime()

  if (formData.value.problems.length === 0) {
    ElMessage.error('请选择比赛题目')
    return
  }

  const contestId = route.params.id
  isSaving.value = true
  try {
    const response = await fetch(`/api/contests/${contestId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify(formData.value),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.message || '更新失败')
    }

    ElMessage.success('更新成功')
    router.push('/admin/contest/manage')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '更新失败')
  } finally {
    isSaving.value = false
  }
}

const newProblem = ref('')

// 处理添加题目
const handleAddProblem = () => {
  const input = newProblem.value.trim()
  if (!input) return

  // 处理范围输入 (如 "10001-10005")
  if (input.includes('-')) {
    const [start, end] = input.split('-').map((num) => parseInt(num))
    if (!isNaN(start) && !isNaN(end) && start <= end) {
      for (let i = start; i <= end; i++) {
        const problemId = i.toString().padStart(5, '0')
        if (!formData.value.problems.includes(problemId)) {
          formData.value.problems.push(problemId)
        }
      }
    } else {
      ElMessage.error('无效的题目范围')
    }
  } else {
    // 处理单个题目ID
    const problemId = input.padStart(5, '0')
    if (!formData.value.problems.includes(problemId)) {
      formData.value.problems.push(problemId)
    }
  }
  newProblem.value = '' // 清空输入
}

// 删除题目
const handleRemoveProblem = (problem: string) => {
  formData.value.problems = formData.value.problems.filter((p) => p !== problem)
}

// 清空所有题目
const clearProblems = () => {
  formData.value.problems = []
}

onMounted(() => {
  fetchContest()
})
</script>

<style scoped>
.contest-edit {
  min-height: calc(100vh - 64px);
  padding: 20px;
}

.edit-container {
  max-width: 1000px;
  margin: 0 auto;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.edit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
  align-items: center;
}

.form-row.full-width {
  grid-template-columns: 1fr;
}

.title-input {
  grid-column: 1;
}

.editor-container {
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.2);
}

.editor-toolbar {
  padding: 0.5rem;
  background: rgba(0, 0, 0, 0.2);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.editor-content {
  min-height: 400px;
}

.markdown-preview {
  padding: 1rem;
  min-height: 400px;
  background: transparent;
  overflow-y: auto;
}

.problems-select {
  width: 100%;
}

:deep(.el-select) {
  width: 100%;
}

:deep(.el-date-picker),
:deep(.el-time-picker) {
  width: 100%;
}

.problems-input {
  width: 100%;
}

.problems-container {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.problem-item {
  margin-right: 4px;
  margin-bottom: 4px;
}

:deep(.el-tag) {
  background: rgba(0, 105, 150, 0.1) !important;
  border: 1px solid rgba(0, 105, 150, 0.2) !important;
  color: #006996 !important;
  border-radius: 12px !important;
  padding: 2px 10px !important;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  transition: all 0.3s ease;
}

:deep(.el-tag:hover) {
  background: rgba(0, 105, 150, 0.2) !important;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 105, 150, 0.2);
}

.input-row {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.input-row .el-input {
  flex: 1;
}
</style>
