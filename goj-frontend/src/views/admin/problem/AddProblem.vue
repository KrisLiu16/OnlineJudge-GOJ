<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { ElMessage } from 'element-plus'
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
    katex: katex,
  }),
)

const router = useRouter()
const userStore = useUserStore()

// 语言选项
const languageOptions = [
  { value: 'c', label: 'C' },
  { value: 'cpp', label: 'C++' },
  { value: 'java', label: 'Java' },
  { value: 'python', label: 'Python' },
  { value: 'go', label: 'Go' },
]

// 修改表单初始值，languages 默认包含所有语言
const defaultLanguages = languageOptions.map((lang) => lang.value)

// 添加新的 ref 用于临时存储输入的 tag
const newTag = ref('')

const formData = ref({
  title: '',
  content: '',
  difficulty: 1,
  source: '',
  tags: [] as string[], // 修为字符串数组
  role: 'admin' as 'user' | 'admin',
  languages: [...defaultLanguages], // 默认全选所有语言
  timeLimit: 1000, // 默认1000ms
  memoryLimit: 128, // 默认128MB
  useSPJ: false,
  spjCode: '',
})

const currentTab = ref('编辑')
const isSaving = ref(false)

const renderedContent = computed(() => {
  try {
    return marked(formData.value.content || '')
  } catch (error) {
    console.error('Markdown渲染错误:', error)
    return '渲染错误'
  }
})

// 难度选项
const difficultyOptions = [
  { value: 1, label: '★' },
  { value: 2, label: '★★' },
  { value: 3, label: '★★★' },
  { value: 4, label: '★★★★' },
  { value: 5, label: '★★★★★' },
]

// 自动保存草稿
const saveDraft = () => {
  localStorage.setItem(
    'problemDraft',
    JSON.stringify({
      ...formData.value,
      lastSaved: new Date().toISOString(),
    }),
  )
}

// 加载草稿
const loadDraft = () => {
  const draft = localStorage.getItem('problemDraft')
  if (draft) {
    const draftData = JSON.parse(draft)
    formData.value = {
      ...draftData,
      tags: Array.isArray(draftData.tags) ? draftData.tags : [], // 确保 tags 是数组
    }
  }
}

// 清除草稿和重置表单
const resetForm = () => {
  formData.value = {
    title: '',
    content: '',
    difficulty: 1,
    source: '',
    tags: [], // 重置为空数组
    role: 'admin',
    languages: [...defaultLanguages], // 重置时也保持全选
    timeLimit: 1000, // 重置为默认值
    memoryLimit: 128, // 重置为默认值
    useSPJ: false,
    spjCode: '',
  }
  newTag.value = '' // 清空 tag 输入
  localStorage.removeItem('problemDraft')
}

const handleSuccess = () => {
  router.push('/admin/problem/manage')
}

const save = async () => {
  if (isSaving.value) return
  if (!formData.value.title || !formData.value.content) {
    ElMessage.error('标题和内容不能为空')
    return
  }

  isSaving.value = true
  try {
    const response = await fetch('/api/problems/add', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify(formData.value),
    })
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.message || '添加失败')
    }

    ElMessage.success('添加成功')
    resetForm()
    handleSuccess()
  } catch (error) {
    console.error('添加失败:', error)
    ElMessage.error(error instanceof Error ? error.message : '添加失败，请重试')
  } finally {
    isSaving.value = false
  }
}

// 自动保存定时器
let autoSaveTimer: number
onMounted(() => {
  loadDraft()
  autoSaveTimer = window.setInterval(saveDraft, 30000)
})

onUnmounted(() => {
  clearInterval(autoSaveTimer)
})

// 添加 tag 的方法
const handleAddTag = () => {
  const tag = newTag.value.trim()
  if (tag && !formData.value.tags.includes(tag)) {
    formData.value.tags.push(tag)
  }
  newTag.value = '' // 清空输入
}

// 删除 tag 的方法
const handleRemoveTag = (tag: string) => {
  formData.value.tags = formData.value.tags.filter((t: string) => t !== tag)
}

// 添加图片压缩工具函数
const compressImage = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = (e) => {
      const img = new Image()
      img.src = e.target?.result as string
      img.onload = () => {
        const canvas = document.createElement('canvas')
        let width = img.width
        let height = img.height

        // 如果图片大于 1000px，按比例缩小
        const maxSize = 1000
        if (width > maxSize || height > maxSize) {
          if (width > height) {
            height = Math.round((height * maxSize) / width)
            width = maxSize
          } else {
            width = Math.round((width * maxSize) / height)
            height = maxSize
          }
        }

        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')
        ctx?.drawImage(img, 0, 0, width, height)

        // 压缩图片质量
        const compressedBase64 = canvas.toDataURL('image/jpeg', 0.6)
        resolve(compressedBase64)
      }
      img.onerror = reject
    }
    reader.onerror = reject
  })
}

// 添加图片上传处理函数
const handleImageUpload = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return

  const file = input.files[0]
  if (!file.type.startsWith('image/')) {
    ElMessage.error('请上传图片文件')
    return
  }

  try {
    const base64 = await compressImage(file)
    // 在光标位置插入 Markdown 图片语法
    const textarea = document.querySelector('.editor-content textarea') as HTMLTextAreaElement
    const cursorPos = textarea.selectionStart
    const textBefore = formData.value.content.substring(0, cursorPos)
    const textAfter = formData.value.content.substring(cursorPos)
    formData.value.content = `${textBefore}\n![image](${base64})\n${textAfter}`

    ElMessage.success('图片上传成功')
  } catch (error) {
    console.error('图片处理失败:', error)
    ElMessage.error('图片处理失败')
  }

  // 清空 input 值，允许重复上传同一张图片
  input.value = ''
}
</script>

<template>
  <div class="problem-edit">
    <div class="edit-container">
      <div class="edit-header">
        <h1>添加题目</h1>
        <div class="header-actions">
          <el-button type="primary" @click="save" :loading="isSaving">
            {{ isSaving ? '添加中...' : '添加题目' }}
          </el-button>
        </div>
      </div>

      <div class="edit-form">
        <div class="form-row">
          <div class="form-group title-input">
            <el-input v-model="formData.title" placeholder="输入题目标题..." @input="saveDraft" />
          </div>
          <div class="form-group limit-input">
            <el-input-number
              v-model="formData.timeLimit"
              :min="0"
              :max="100000"
              placeholder="时间限制"
            >
              <template #append>MS</template>
            </el-input-number>
          </div>
          <div class="form-group limit-input">
            <el-input-number
              v-model="formData.memoryLimit"
              :min="0"
              :max="2048"
              :step="128"
              placeholder="内存限制"
            >
              <template #append>MB</template>
            </el-input-number>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <el-select v-model="formData.difficulty" placeholder="选择难度" @change="saveDraft">
              <el-option
                v-for="option in difficultyOptions"
                :key="option.value"
                :value="option.value"
                :label="option.label"
              />
            </el-select>
          </div>

          <div class="form-group">
            <el-input v-model="formData.source" placeholder="题目来源" @input="saveDraft" />
          </div>

          <div class="form-group">
            <el-select v-model="formData.role" placeholder="选择权限" @change="saveDraft">
              <el-option value="admin" label="管理员" />
              <el-option value="user" label="用户" />
            </el-select>
          </div>
        </div>

        <div class="form-row full-width">
          <div class="form-group languages-select">
            <el-select
              v-model="formData.languages"
              multiple
              placeholder="选择可用语言"
              @change="saveDraft"
              :default-first-option="true"
            >
              <el-option
                v-for="lang in languageOptions"
                :key="lang.value"
                :value="lang.value"
                :label="lang.label"
              />
            </el-select>
          </div>
        </div>

        <div class="form-row full-width">
          <div class="tags-input">
            <el-input
              v-model="newTag"
              placeholder="输入标签后按回车添加"
              @keyup.enter="handleAddTag"
            >
              <template #append>
                <el-button @click="handleAddTag">添加</el-button>
              </template>
            </el-input>
            <div class="tags-container">
              <el-tag
                v-for="tag in formData.tags"
                :key="tag"
                closable
                @close="handleRemoveTag(tag)"
                class="tag-item"
              >
                {{ tag }}
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

            <div class="upload-btn" v-if="currentTab === '编辑'">
              <input
                type="file"
                accept="image/*"
                @change="handleImageUpload"
                style="display: none"
                ref="imageInput"
              >
              <el-button
                type="primary"
                size="small"
                @click="$refs.imageInput.click()"
              >
                上传图片
              </el-button>
            </div>
          </div>

          <div class="editor-content">
            <el-input
              v-if="currentTab === '编辑'"
              v-model="formData.content"
              type="textarea"
              :rows="15"
              placeholder="使用 Markdown 编写题目内容..."
              @input="saveDraft"
            />
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

        <div class="spj-section">
          <div class="spj-header">
            <el-checkbox v-model="formData.useSPJ" @change="saveDraft"
              >使用特殊判题(SPJ)</el-checkbox
            >
            <el-tooltip content="特殊判题用于处理多解或需要特殊比较的题目" placement="top">
              <el-icon><QuestionFilled /></el-icon>
            </el-tooltip>
          </div>

          <div v-if="formData.useSPJ" class="spj-editor">
            <div class="editor-header">
              <span>SPJ代码 (C++)</span>
              <el-tooltip content="支持testlib.h库,可用于实现复杂的判题逻辑" placement="top">
                <el-icon><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-input
              v-model="formData.spjCode"
              type="textarea"
              :rows="10"
              @input="saveDraft"
              :placeholder="`#include <testlib.h>
// SPJ代码示例
int main(int argc, char* argv[]) {
    registerTestlibCmd(argc, argv);
    // 读取选手输出
    double userAns = ouf.readDouble();
    // 读取标准输出
    double judgeAns = ans.readDouble();
    // 检查答案
    if (abs(userAns - judgeAns) < 1e-6) {
        quitf(_ok, &quot;答案正确&quot;);
    } else {
        quitf(_wa, &quot;答案错误&quot;);
    }
}`"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.problem-edit {
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
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
  align-items: center;
}
.form-row.full-width {
  grid-template-columns: 1fr;
}

.form-row .tags-input {
  grid-column: 1 / -1;
}

.form-row .languages-select {
  grid-column: 1 / -1;
}

.title-input {
  grid-column: 1;
}

.limit-input {
  width: 100%;
}

.limit-input :deep(.el-input-number) {
  width: 100%;
}

.editor-container {
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.2);
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.languages-select {
  flex-grow: 1;
}

.languages-select :deep(.el-select) {
  width: 100%;
}

.form-row .el-switch {
  margin-top: 8px;
}

.tags-input {
  width: 100%;
}

.tags-container {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-item {
  margin-right: 4px;
  margin-bottom: 4px;
}

.spj-section {
  margin-top: 2rem;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.spj-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 1rem;
}

.spj-editor {
  margin-top: 1rem;
}

.editor-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 0.5rem;
  color: var(--text-light);
}

.editor-header .el-icon {
  font-size: 16px;
  color: var(--el-color-info);
  cursor: help;
}

.upload-btn {
  margin-left: 1rem;
}
</style>

