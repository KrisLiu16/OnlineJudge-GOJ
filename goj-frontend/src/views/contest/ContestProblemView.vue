<template>
  <!-- 添加 loading 遮罩 -->
  <div v-if="!problem.id" class="loading-overlay">
    <div class="loading-spinner"></div>
    <div class="loading-text">加载中...</div>
  </div>

  <div class="problem-detail-container" v-show="problem.id" :class="{ 'fade-in': problem.id }">
    <!-- 左侧题目描述区域 -->
    <div class="problem-description" :style="{ width: leftWidth + 'px' }">
      <div class="problem-header">
        <div class="title-row">
          <div class="title-left">
            <span class="problem-status">{{ getProblemStatus(problem.status) }}</span>
            <div class="problem-title">
              <span class="problem-id">{{ problemIndex }}.</span>
              <h1>{{ problem.title }}</h1>
            </div>
          </div>
          <div class="title-actions">
            <router-link :to="`/submissions?problemId=${problem.id}`" class="action-btn">
              提交记录
            </router-link>
            <router-link :to="`/contest/${contestId}`" class="action-btn"> 返回比赛 </router-link>
          </div>
        </div>

        <div class="problem-tags" v-if="problem.tags?.length">
          <span v-for="tag in problem.tags" :key="tag" class="tag">
            {{ tag }}
          </span>
        </div>

        <div class="problem-meta">
          <div class="meta-left">
            <div class="difficulty-display">
                    <div class="difficulty-progress">
                      <div
                        class="difficulty-bar"
                        :style="{
                          width: `${problem.difficulty * 20}%`,
                          background: getDifficultyGradient(problem.difficulty)
                        }"
                      ></div>
                    </div>
                    <span class="difficulty-text">{{ getDifficultyLabel(problem.difficulty) }}</span>
                  </div>
          </div>
          <div class="meta-right">
            <span class="stats"> 提交：{{ problem.submissionCount }} </span>
            <span class="stats"> 通过：{{ problem.acceptedCount }} </span>
            <span class="stats">
              通过率：{{
                ((problem.acceptedCount / problem.submissionCount) * 100 || 0).toFixed(1)
              }}%
            </span>
          </div>
        </div>
      </div>

      <div class="problem-info">
        <div class="info-item">
          <span class="label">时间限制:</span>
          <span class="value">{{ problem.timeLimit }}ms</span>
        </div>
        <div class="info-item">
          <span class="label">内存限制:</span>
          <span class="value">{{ problem.memoryLimit }}MB</span>
        </div>
      </div>
      <div class="problem-content markdown-body" v-html="renderedContent"></div>
    </div>

    <!-- 可拖动分隔条 -->
    <div class="resizer" @mousedown="startResize"></div>

    <!-- 右侧代码编辑区域 -->
    <div class="code-editor" :style="{ width: `calc(100% - ${leftWidth}px)` }">
      <div class="editor-header">
        <select v-model="selectedLanguage" class="language-select">
          <option v-for="lang in availableLanguages" :key="lang.value" :value="lang.value">
            {{ lang.label }}
          </option>
        </select>
        <button @click="handleSubmit" class="submit-btn" :disabled="isSubmitting">
          {{ isSubmitting ? '提交中...' : '提交' }}
        </button>
      </div>

      <div class="editor-body">
        <div ref="editorContainer" class="monaco-editor"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted, watch, inject, type Ref } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import { ElMessage, ElMessageBox } from 'element-plus'
import katex from 'katex'
import 'katex/dist/katex.min.css'
import markedKatex from 'marked-katex-extension'
import { useUserStore } from '@/stores/modules/user'
import { useRouter } from 'vue-router'
import * as monaco from 'monaco-editor'

// 配置 marked 支持 KaTeX
marked.use(
  markedKatex({
    throwOnError: false,
    output: 'html',
    displayMode: false,
    strict: false,
    trust: true,
    macros: {
      '\\bm': '\\boldsymbol',
    },
    delimiters: [
      { left: '$$', right: '$$', display: true },
      { left: '$', right: '$', display: false },
      { left: '\\(', right: '\\)', display: false },
      { left: '\\[', right: '\\]', display: true },
    ],
    katex: {
      ...katex,
      strict: false,
    },
  }),
)

const route = useRoute()
const contestId = route.params.contestId as string
const problemIndex = route.params.problemIndex as string // 直接使用路由参数中的字母索引(A,B,C...)

const leftWidth = ref(window.innerWidth * 0.65) // 设置初始宽度为窗口宽度的65%
const minWidth = 400 // 最小宽度
const maxWidth = window.innerWidth * 0.8 // 最大宽度设为窗口宽度的80%

// 拖动相关状态
const isResizing = ref(false)
const startX = ref(0)
const startWidth = ref(0)

const userStore = useUserStore()
const router = useRouter()

const problem = ref({
  id: '',
  title: '',
  content: '',
  difficulty: 1,
  timeLimit: 1000,
  memoryLimit: 128,
  acceptedCount: 0,
  submissionCount: 0,
  languages: [] as string[],
  tags: [] as string[],
  Role: 'user',
  status: 'unattempted',
})

const code = ref('')
const selectedLanguage = ref('cpp')
const isSubmitting = ref(false)

const availableLanguages = computed(() => {
  const languageMap: Record<string, { value: string; label: string }> = {
    cpp: { value: 'cpp', label: 'C++' },
    c: { value: 'c', label: 'C' },
    java: { value: 'java', label: 'Java' },
    python: { value: 'python', label: 'Python' },
    go: { value: 'go', label: 'Go' },
  }

  return problem.value.languages
    .map((lang: string) => languageMap[lang as keyof typeof languageMap])
    .filter(Boolean)
})

const renderedContent = computed(() => {
  try {
    return marked(problem.value.content || '')
  } catch (error) {
    console.error('Markdown渲染错误:', error)
    return '渲染错误'
  }
})

// 拖动处理函数
const startResize = (e: MouseEvent) => {
  isResizing.value = true
  startX.value = e.pageX
  startWidth.value = leftWidth.value
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
}

const handleResize = (e: MouseEvent) => {
  if (!isResizing.value) return

  const diff = e.pageX - startX.value
  const newWidth = Math.min(Math.max(startWidth.value + diff, minWidth), maxWidth)
  leftWidth.value = newWidth
}

const stopResize = () => {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
}

// 获取题目列表并找到对应的题目ID
const fetchProblemId = async () => {
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
      const problems = data.data.problems
      const index = problemIndex.charCodeAt(0) - 65 // 将 A,B,C 转换为 0,1,2...
      if (index >= 0 && index < problems.length) {
        const problemId = problems[index].id
        // 使用实际的题目ID获取题目详情
        await fetchProblemDetail(problemId)
      } else {
        throw new Error('题目不存在')
      }
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '获取题目失败')
    router.push(`/contest/${contestId}`)
  }
}

// 修改获取题目详情的函数
const fetchProblemDetail = async (problemId: string) => {
  try {
    if (!userStore.token) {
      ElMessage.error('未登录或 token 已过期')
      router.push('/sign-in')
      return
    }

    const response = await fetch(`/api/contests/problem/${contestId}/${problemId}`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
        'Content-Type': 'application/json',
      },
    })

    if (response.status === 403) {
      ElMessage.error('没有访问权限')
      router.push('/problems')
      return
    }

    if (!response.ok) {
      throw new Error('获取题目失败')
    }

    const result = await response.json()
    if (result.code === 200) {
      problem.value = {
        ...result.data,
        status: result.data.status || 'unattempted',
      }

      // 设置默认语言
      if (problem.value.languages.length > 0) {
        selectedLanguage.value = 'cpp'
      }
    } else {
      throw new Error(result.message)
    }
  } catch (error) {
    console.error('获取题目失败:', error)
    ElMessage.error(error instanceof Error ? error.message : '获取题目失败')
    router.push(`/contest/${contestId}`)
  }
}

// 提交代码
const handleSubmit = async () => {
  if (!code.value.trim()) {
    ElMessage.warning('请输入代码')
    return
  }

  if (!userStore.token) {
    ElMessage.warning('请先登录')
    router.push('/sign-in')
    return
  }

  // 添加确认对话框
  const confirmed = await ElMessageBox.confirm('确认提交代码？', '提交确认', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'info',
  }).catch(() => false)

  if (!confirmed) return

  isSubmitting.value = true
  try {
    const response = await fetch('/api/submit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify({
        problemId: problem.value.id,
        contestId: contestId,
        language: selectedLanguage.value,
        code: code.value,
      }),
    })

    if (response.status === 403) {
      ElMessage.error('没有提交权限')
      return
    }

    if (!response.ok) {
      throw new Error('提交失败')
    }

    const result = await response.json()
    if (result.code === 200) {
      ElMessage.success('提交成功')
      // 在新标签页打开提交详情页面
      window.open(`/submission/${result.data.id}`, '_blank')
    } else {
      throw new Error(result.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '提交失败')
  } finally {
    isSubmitting.value = false
  }
}

const editorContainer = ref<HTMLElement>()
let editor: monaco.editor.IStandaloneCodeEditor

const isLightTheme = inject('isLightTheme') as Ref<boolean>

// 初始化编辑器
const initEditor = () => {
  if (!editorContainer.value) return


  editor = monaco.editor.create(editorContainer.value, {
    value: code.value,
    language: selectedLanguage.value,
    theme: isLightTheme.value ? 'vs' : 'vs-dark',
    automaticLayout: true,
    minimap: { enabled: false },
    fontSize: 14,
    tabSize: 4,
    scrollBeyondLastLine: false,
    suggestOnTriggerCharacters: true,
    formatOnType: true,
    formatOnPaste: true,
  })

  // 监听编辑器内容变化
  editor.onDidChangeModelContent(() => {
    code.value = editor.getValue()
  })

  // 监听语言变化
  watch(selectedLanguage, (newLang: string) => {
    monaco.editor.setModelLanguage(editor.getModel()!, newLang)
  })

  // 监听主题变化
  watch(isLightTheme, (isLight: boolean) => {
    monaco.editor.setTheme(isLight ? 'vs' : 'vs-dark')
  })
}

onMounted(() => {
  if (!userStore.token) {
    router.push('/sign-in')
    return
  }
  fetchProblemId()
  initEditor()
})

onUnmounted(() => {
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  if (editor) {
    editor.dispose()
  }
})
// 添加获取题目状态的函数
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
const getDifficultyLabel = (level: number) => {
  switch (level) {
    case 1:
      return '入门'
    case 2:
      return '简单'
    case 3:
      return '中等'
    case 4:
      return '困难'
    case 5:
      return '专家'
    default:
      return '未知'
  }
}

const getDifficultyGradient = (level: number) => {
  switch (level) {
    case 1:
      return 'linear-gradient(90deg, #00b09b, #96c93d)'
    case 2:
      return 'linear-gradient(90deg, #96c93d, #4facfe)'
    case 3:
      return 'linear-gradient(90deg, #4facfe, #ffd700)'
    case 4:
      return 'linear-gradient(90deg, #ffd700, #ff5858)'
    case 5:
      return 'linear-gradient(90deg, #ff5858, #ff0000)'
    default:
      return 'linear-gradient(90deg, #ccc, #ccc)'
  }
}

</script>

<style scoped>
.problem-detail-container {
  display: flex;
  position: fixed;
  top: 64px; /* 顶部导航栏的高 */
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden; /* 防止整个页面滚动 */
}

.problem-description {
  height: 100%;
  overflow-y: auto; /* 只允许左侧内容滚动 */
  padding: 2rem;
  background: var(--bg-color);
}

.title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.title-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.problem-status {
  font-size: 1.5rem;
}

.problem-title {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.problem-id {
  font-size: 1.5rem;
  color: var(--text-gray);
  font-weight: bold;
}

.title-left h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: bold;
}

.title-actions {
  display: flex;
  gap: 1rem;
}

.action-btn {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-light);
  text-decoration: none;
  transition: all 0.2s;
}

.action-btn:hover {
  background: var(--primary-color);
  color: white;
}

.problem-tags {
  display: flex;
  gap: 8px;
  margin: 1rem 0;
  flex-wrap: wrap;
}

.tag {
  padding: 4px 12px;
  border-radius: 6px;
  background: #006996;
  color: white;
  font-size: 0.9rem;
  transition: all 0.3s ease;
  font-weight: 500;
  letter-spacing: 0.3px;
  border: none;
  box-shadow: 0 2px 4px rgba(0, 105, 150, 0.2);
}

.tag:hover {
  background: #007bb1;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 105, 150, 0.3);
}

.resizer {
  width: 4px;
  height: 100%;
  background: rgba(255, 255, 255, 0.1);
  cursor: col-resize;
  transition: background 0.2s;
}

.resizer:hover {
  background: var(--primary-color);
}

.code-editor {
  height: 100%; /* 改为100%而不是100vh */
  display: flex;
  flex-direction: column;
  background: rgba(0, 0, 0, 0.2);
  overflow: hidden; /* 确保编辑器不会滚动 */
}

.problem-header {
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.problem-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 1rem;
  color: var(--text-light);
}

.meta-left,
.meta-right {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.meta-right .stats {
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.1);
  font-size: 0.9rem;
}

.difficulty-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
}

.difficulty-progress {
  width: 60px;
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
  position: relative;
}

.difficulty-bar {
  height: 100%;
  border-radius: 3px;
  transition: all 0.3s ease;
}

.difficulty-text {
  font-size: 0.85rem;
  color: var(--text-light);
  white-space: nowrap;
  min-width: 32px;
}

.difficulty, .star {
  display: none;
}

.problem-content {
  margin-bottom: 2rem;
}

.problem-info {
  display: flex;
  gap: 2rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
}

.info-item {
  display: flex;
  gap: 0.5rem;
}

.info-item .label {
  color: var(--text-light);
}

.editor-header {
  padding: 1rem;
  display: flex;
  gap: 1rem;
  background: rgba(0, 0, 0, 0.3);
}

.language-select {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.3); /* 修改背景为透明 */
  border: 1px solid rgba(255, 255, 255, 0.2); /* 边框颜色可以根据实际需求调整 */
  color: var(--text-light);
}

.submit-btn {
  padding: 0.5rem 1.5rem;
  border: none;
  border-radius: 4px;
  background: var(--primary-color);
  color: white;
  cursor: pointer;
  transition: all 0.2s;
}

.submit-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-body {
  flex: 1;
  padding: 1rem;
}

.monaco-editor {
  width: 100%;
  height: 100%;
}

@media (max-width: 768px) {
  .problem-detail-container {
    position: fixed;
    top: 64px;
    bottom: 0;
  }

  .problem-description,
  .code-editor {
    height: 50%;
  }
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
</style>
