<template>
  <div class="problem-edit">
    <div class="edit-container">
      <div class="edit-header">
        <h1>编辑题目</h1>
        <div class="header-actions">
          <el-button type="primary" @click="save" :loading="isSaving">
            {{ isSaving ? '保存中...' : '保存修改' }}
          </el-button>
        </div>
      </div>

      <div class="edit-form" v-loading="loading">
        <div class="form-row">
          <div class="form-group title-input">
            <el-input v-model="formData.title" placeholder="输入题目标题..." />
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
            <el-select v-model="formData.difficulty" placeholder="选择难度">
              <el-option
                v-for="option in difficultyOptions"
                :key="option.value"
                :value="option.value"
                :label="option.label"
              />
            </el-select>
          </div>

          <div class="form-group">
            <el-input v-model="formData.source" placeholder="题目来源" />
          </div>

          <div class="form-group">
            <el-select v-model="formData.role" placeholder="选择权限">
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
              <el-radio-button value="编辑">编辑</el-radio-button>
              <el-radio-button value="预览">预览</el-radio-button>
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
            />
            <div v-else class="markdown-preview markdown-body" v-html="renderedContent"></div>
          </div>
        </div>

        <div class="spj-section">
          <div class="spj-header">
            <el-checkbox v-model="formData.useSPJ">使用特殊判题(SPJ)</el-checkbox>
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

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
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
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)
const isSaving = ref(false)
const currentTab = ref('编辑')
const newTag = ref('')

// 语言选项
const languageOptions = [
  { value: 'c', label: 'C' },
  { value: 'cpp', label: 'C++' },
  { value: 'java', label: 'Java' },
  { value: 'python', label: 'Python' },
  { value: 'go', label: 'Go' },
]

// 难度选项
const difficultyOptions = [
  { value: 1, label: '★' },
  { value: 2, label: '★★' },
  { value: 3, label: '★★★' },
  { value: 4, label: '★★★★' },
  { value: 5, label: '★★★★★' },
]

// 表单数据
const formData = ref({
  title: '',
  content: '',
  difficulty: 1,
  source: '',
  tags: [] as string[],
  role: 'admin' as 'user' | 'admin',
  languages: [] as string[],
  timeLimit: 1000,
  memoryLimit: 128,
  useSPJ: false,
  spjCode: '',
})

// Markdown预览
const renderedContent = computed(() => {
  try {
    return marked(formData.value.content || '')
  } catch (error) {
    console.error('Markdown渲染错误:', error)
    return '渲染错误'
  }
})

// 获取题目数据
const fetchProblem = async () => {
  const problemId = route.params.id
  loading.value = true
  try {
    // 确保 userStore.token 存在
    if (!userStore.token) {
      ElMessage.error('未登录或 token 已过期')
      router.push('/sign-in')
      return
    }

    // 获取基本题目信息
    const response = await fetch(`/api/admin/problems/${problemId}`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
        'Content-Type': 'application/json',
      },
    })

    if (response.status === 403) {
      ElMessage.error('没有访问权限')
      router.push('/admin/problem/manage')
      return
    }

    if (!response.ok) {
      throw new Error('获取题目数据失败')
    }

    const result = await response.json()
    if (result.code === 200) {
      const problem = result.data

      formData.value = {
        title: problem.title || '',
        content: problem.content || '',
        difficulty: problem.difficulty || 1,
        source: problem.source || '',
        tags: Array.isArray(problem.tags)
          ? problem.tags
          : (problem.tags || '').split(',').filter(Boolean),
        role: problem.role || 'admin',
        languages: Array.isArray(problem.languages)
          ? problem.languages
          : (problem.languages || '').split(',').filter(Boolean),
        timeLimit: Number(problem.timeLimit) || 1000,
        memoryLimit: Number(problem.memoryLimit) || 128,
        useSPJ: Boolean(problem.useSPJ),
        spjCode: '',
      }

      // 如果是 SPJ 题目，获取 SPJ 代码
      if (formData.value.useSPJ) {
        try {
          const spjResponse = await fetch(`/api/admin/problems/${problemId}/spj`, {
            headers: {
              Authorization: `Bearer ${userStore.token}`,
            },
          })

          if (spjResponse.ok) {
            const spjResult = await spjResponse.json()
            if (spjResult.code === 200) {
              formData.value.spjCode = spjResult.data
            }
          } else {
            const errorText = await spjResponse.text()
            console.warn('获取SPJ代码失败:', errorText)
          }
        } catch (error) {
          console.error('获取SPJ代码时发生错误:', error)
        }
      } else {
        console.log('不是SPJ题目，跳过获取SPJ代码')
      }
    } else {
      throw new Error(result.message)
    }
  } catch (error) {
    console.error('获取题目数据失败:', error)
    ElMessage.error(error instanceof Error ? error.message : '获取题目数据失败')
  } finally {
    loading.value = false
  }
}

// 保存修改
const save = async () => {
  if (isSaving.value) return
  if (!formData.value.title || !formData.value.content) {
    ElMessage.error('标题和内容不能为空')
    return
  }

  const problemId = route.params.id
  isSaving.value = true
  try {
    const response = await fetch(`/api/admin/problems/${problemId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify(formData.value),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.message || '保存失败')
    }

    const result = await response.json()
    if (result.code === 200) {
      ElMessage.success('保存成功')
      router.push('/admin/problem/manage')
    } else {
      throw new Error(result.message)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    isSaving.value = false
  }
}

// 标签操作
const handleAddTag = () => {
  const tag = newTag.value.trim()
  if (tag && !formData.value.tags.includes(tag)) {
    formData.value.tags.push(tag)
  }
  newTag.value = ''
}

const handleRemoveTag = (tag: string) => {
  formData.value.tags = formData.value.tags.filter((t) => t !== tag)
}

// 在 script setup 部分添加图片压缩工具函数
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

onMounted(() => {
  fetchProblem()
})
</script>

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
