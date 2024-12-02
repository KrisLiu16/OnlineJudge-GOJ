<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { marked } from 'marked'
import katex from 'katex'
import 'katex/dist/katex.min.css'
import markedKatex from 'marked-katex-extension'

// 配置 marked
marked.use(
  markedKatex({
    throwOnError: false,
    output: 'html',
    displayMode: false,
    katex,
  }),
)

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isEdit = computed(() => route.path.includes('/edit/'))
const discussionId = computed(() => route.params.id as string)

const title = ref('')
const content = ref('')
const category = ref('discussion')
const currentTab = ref('编辑')
const isSaving = ref(false)

const renderedContent = computed(() => {
  try {
    return marked(content.value || '')
  } catch (error) {
    console.error('Markdown渲染错误:', error)
    return '渲染错误'
  }
})

// 自动保存草稿
const saveDraft = () => {
  localStorage.setItem(
    'discussDraft',
    JSON.stringify({
      title: title.value,
      content: content.value,
      category: category.value,
      lastSaved: new Date().toISOString(),
    }),
  )
}

// 加载草稿
const loadDraft = () => {
  const draft = localStorage.getItem('discussDraft')
  if (draft) {
    const { title: draftTitle, content: draftContent, category: draftCategory } = JSON.parse(draft)
    title.value = draftTitle
    content.value = draftContent
    category.value = draftCategory
  }
}

// 清除草稿
const clearDraft = () => {
  localStorage.removeItem('discussDraft')
}

// 添加登录检查
onMounted(async () => {
  if (!userStore.isAuthenticated) {
    router.push({
      name: 'sign-in',
      query: { redirect: route.fullPath },
    })
    return
  }

  if (isEdit.value) {
    try {
      const response = await fetch(`/api/discussions/${discussionId.value}`)
      if (!response.ok) throw new Error('获取讨论详情失败')
      const discussion = await response.json()

      title.value = discussion.title
      content.value = discussion.content
      category.value = discussion.category
    } catch (error) {
      console.error('获取讨论详情失败:', error)
      router.push('/discuss')
    }
  } else {
    loadDraft()
  }
})

const save = async () => {
  if (isSaving.value) return
  if (!title.value || !content.value) {
    alert('标题和内容不能为空')
    return
  }

  isSaving.value = true
  try {
    const url = isEdit.value ? `/api/discussions/${discussionId.value}` : '/api/discussions'

    const response = await fetch(url, {
      method: isEdit.value ? 'PUT' : 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify({
        title: title.value,
        content: content.value,
        category: category.value,
      }),
    })

    if (!response.ok) throw new Error('保存失败')

    clearDraft()
    router.push('/discuss')
  } catch (error) {
    console.error('保存失败:', error)
    alert('保存失败，请重试')
  } finally {
    isSaving.value = false
  }
}

const handleDelete = async () => {
  if (!confirm('确定要删除这篇讨论吗？')) return

  try {
    const response = await fetch(`/api/discussions/${discussionId.value}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) throw new Error('删除失败')

    router.push('/discuss')
  } catch (error) {
    console.error('删除失败:', error)
    alert('删除失败，请重试')
  }
}

const cancel = () => {
  if (content.value && !confirm('确定要放弃编辑吗？')) return
  router.back()
}

// 自动保存
let autoSaveTimer: number
onMounted(() => {
  autoSaveTimer = window.setInterval(saveDraft, 30000) // 每30秒自动保存
})

onUnmounted(() => {
  clearInterval(autoSaveTimer)
})
</script>

<template>
  <!-- 添加未登录提示（可选） -->
  <div v-if="!userStore.isAuthenticated" class="discuss-edit">
    <div class="edit-container">
      <h2>请先登录后再发布讨论</h2>
      <button
        @click="router.push({ name: 'sign-in', query: { redirect: route.fullPath } })"
        class="btn btn-primary"
      >
        去登录
      </button>
    </div>
  </div>

  <!-- 原有模板 -->
  <div v-else class="discuss-edit">
    <div class="edit-container">
      <div class="edit-header">
        <h1>{{ isEdit ? '编辑讨论' : '发布讨论' }}</h1>
        <div class="header-actions">
          <button @click="cancel" class="btn btn-secondary">取消</button>
          <button v-if="isEdit" @click="handleDelete" class="btn btn-danger">删除</button>
          <button @click="save" class="btn btn-primary" :disabled="isSaving">
            {{ isSaving ? '保存中...' : isEdit ? '保存修改' : '发布' }}
          </button>
        </div>
      </div>

      <div class="edit-form">
        <div class="form-group">
          <input
            v-model="title"
            type="text"
            placeholder="输入标题..."
            class="title-input"
            @input="saveDraft"
          />
        </div>

        <div class="form-group">
          <select v-model="category" class="category-select" @change="saveDraft">
            <option value="discussion">讨论</option>
            <option value="solution">题解</option>
            <option value="notice">公告</option>
          </select>
        </div>

        <div class="editor-container">
          <div class="editor-toolbar">
            <button
              v-for="tab in ['编辑', '预览']"
              :key="tab"
              @click="currentTab = tab"
              :class="{ active: currentTab === tab }"
              class="tab-btn"
            >
              {{ tab }}
            </button>
          </div>

          <div class="editor-content">
            <textarea
              v-if="currentTab === '编辑'"
              v-model="content"
              placeholder="使用 Markdown 编写内容..."
              class="markdown-editor"
              @input="saveDraft"
            ></textarea>
            <div v-else class="markdown-preview markdown-body" v-html="renderedContent"></div>
          </div>
        </div>

        <div class="markdown-tips">
          <h4>Markdown 和数学公式提示：</h4>
          <ul>
            <li>支持标准 Markdown 语法</li>
            <li>行内公式：$E = mc^2$</li>
            <li>块级公式：$$\sum_{i=1}^n i = \frac{n(n+1)}{2}$$</li>
            <li>代码块：```language</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.discuss-edit {
  min-height: calc(100vh - 64px);
  padding: 80px 2rem 2rem;
  background: var(--nav-bg-dark);
  backdrop-filter: blur(10px);
}

.edit-container {
  max-width: 1000px;
  margin: 0 auto;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
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

.header-actions {
  display: flex;
  gap: 1rem;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--primary-color);
  color: white;
}

.btn-secondary {
  background: var(--bg-darker);
  color: var(--text-light);
}

.btn-danger {
  background: var(--error-color);
  color: white;
}

.form-group {
  margin-bottom: 1.5rem;
}

.title-input {
  width: 100%;
  padding: 1rem;
  font-size: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.2);
  color: var(--text-light);
}

.category-select {
  padding: 0.5rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.2);
  color: var(--text-light);
}

.editor-container {
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.2);
}

.editor-toolbar {
  display: flex;
  gap: 1rem;
  padding: 0.5rem;
  background: rgba(0, 0, 0, 0.2);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.tab-btn {
  padding: 0.5rem 1rem;
  border: none;
  background: none;
  color: var(--text-gray);
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.tab-btn.active {
  background: var(--primary-color);
  color: white;
}

.editor-content {
  min-height: 400px;
}

.markdown-editor {
  width: 100%;
  height: 400px;
  padding: 1rem;
  border: none;
  resize: vertical;
  background: transparent;
  color: var(--text-light);
  font-family: monospace;
  line-height: 1.6;
}

.markdown-preview {
  padding: 1rem;
  min-height: 400px;
  background: transparent;
  overflow-y: auto;
}

.markdown-tips {
  margin-top: 1.5rem;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.markdown-tips h4 {
  margin-bottom: 0.5rem;
  color: var(--text-light);
}

.markdown-tips ul {
  margin: 0;
  padding-left: 1.5rem;
  color: var(--text-gray);
}

.markdown-tips li {
  margin: 0.25rem 0;
}

/* 适配浅色主题 */
:global(.light-theme) .markdown-body {
  color: var(--text-dark);
}

:global(.light-theme) .markdown-body blockquote {
  background: rgba(0, 0, 0, 0.03);
}

:global(.light-theme) .markdown-body pre,
:global(.light-theme) .markdown-body code {
  background: rgba(0, 0, 0, 0.05);
}

:global(.light-theme) .markdown-body table th {
  background: rgba(0, 0, 0, 0.03);
}

:global(.light-theme) .markdown-body table tr:nth-child(2n) {
  background: rgba(0, 0, 0, 0.01);
}
</style>
