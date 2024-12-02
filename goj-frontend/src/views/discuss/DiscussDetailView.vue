<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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

interface Comment {
  id: number
  content: string
  createdAt: string
  author: {
    id: number
    username: string
    avatar: string
  }
}

interface Discussion {
  id: string
  title: string
  content: string
  category: string
  createdAt: string
  author: {
    id: number
    username: string
    avatar: string
    role: string
  }
  stats: {
    likes: number
    comments: number
    stars: number
  }
  interactions?: {
    isLiked: boolean
    isStarred: boolean
  }
}

const discussion = ref<Discussion>({
  id: '',
  title: '',
  content: '',
  category: '',
  createdAt: '',
  author: {
    id: 0,
    username: '',
    avatar: '',
    role: '',
  },
  stats: {
    likes: 0,
    comments: 0,
    stars: 0,
  },
})

const comments = ref<Comment[]>([])
const newComment = ref('')
const isLiked = ref(false)
const isStarred = ref(false)
const isLoading = ref(true)

// 检查权限
const canEdit = computed(() => {
  return (
    userStore.isLoggedIn &&
    (userStore.userInfo?.role === 'admin' ||
      Number(userStore.userInfo?.id) === discussion.value.author.id)
  )
})

const fetchDiscussion = async () => {
  try {
    const response = await fetch(`/api/discussions/${route.params.id}`, {
      headers: {
        Authorization: userStore.token ? `Bearer ${userStore.token}` : '',
      },
    })
    if (!response.ok) throw new Error('获取讨论详情失败')
    const data = await response.json()
    discussion.value = data
    isLiked.value = data.interactions?.isLiked || false
    isStarred.value = data.interactions?.isStarred || false
  } catch (error) {
    console.error('获取讨论详情失败:', error)
    router.push('/discuss')
  } finally {
    isLoading.value = false
  }
}

const fetchComments = async () => {
  try {
    const response = await fetch(`/api/discussions/${route.params.id}/comments`)
    if (!response.ok) throw new Error('获取评论失败')
    const data = await response.json()
    comments.value = data.comments || []
  } catch (error) {
    console.error('获取评论失败:', error)
    comments.value = []
  }
}

const submitComment = async () => {
  if (!newComment.value.trim()) return
  if (!userStore.isLoggedIn) {
    router.push('/login')
    return
  }

  try {
    const response = await fetch(`/api/discussions/${route.params.id}/comments`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`,
      },
      body: JSON.stringify({ content: newComment.value }),
    })

    if (!response.ok) throw new Error('发表评论失败')

    newComment.value = ''
    await fetchComments()
  } catch (error) {
    console.error('发表评论失败:', error)
    alert('发表评论失败，请重试')
  }
}

const handleLike = async () => {
  if (!userStore.isLoggedIn) {
    router.push('/login')
    return
  }

  try {
    const response = await fetch(`/api/discussions/${route.params.id}/like`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) throw new Error('操作失败')

    const data = await response.json()
    isLiked.value = data.isLiked
    discussion.value.stats.likes += data.isLiked ? 1 : -1
  } catch (error) {
    console.error('操作失败:', error)
  }
}

const handleStar = async () => {
  if (!userStore.isLoggedIn) {
    router.push('/login')
    return
  }

  try {
    const response = await fetch(`/api/discussions/${route.params.id}/star`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) throw new Error('操作失败')

    const data = await response.json()
    isStarred.value = data.isStarred
    discussion.value.stats.stars += data.isStarred ? 1 : -1
  } catch (error) {
    console.error('操作失败:', error)
  }
}

const handleDelete = async () => {
  if (!confirm('确定要删除这篇讨论吗？')) return

  try {
    const response = await fetch(`/api/discussions/${discussion.value.id}`, {
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

const formatTime = (time: string) => {
  const date = new Date(time)
  return date.toLocaleString()
}

onMounted(() => {
  fetchDiscussion()
  fetchComments()
})
</script>

<template>
  <div class="discuss-detail" v-if="!isLoading">
    <div class="post-content">
      <div class="post-header">
        <h1>{{ discussion.title }}</h1>
        <div class="post-meta">
          <img :src="discussion.author.avatar" alt="avatar" class="avatar" />
          <router-link
            :to="`/profile/${discussion.author.username}`"
            class="username"
            :class="{ admin: discussion.author.role === 'admin' }"
          >
            {{ discussion.author.username }}
          </router-link>
          <span class="time">{{ formatTime(discussion.createdAt) }}</span>
          <span class="tag">{{ discussion.category }}</span>

          <div v-if="canEdit" class="actions">
            <router-link :to="`/discuss/edit/${discussion.id}`" class="edit-btn">
              编辑
            </router-link>
            <button @click="handleDelete" class="delete-btn">删除</button>
          </div>
        </div>
      </div>

      <div class="markdown-content markdown-body" v-html="marked(discussion.content)"></div>

      <div class="post-actions">
        <button @click="handleLike" :class="{ active: isLiked }" class="action-btn">
          <i class="far fa-thumbs-up"></i> {{ discussion.stats.likes }}
        </button>
        <button @click="handleStar" :class="{ active: isStarred }" class="action-btn">
          <i class="far fa-star"></i> {{ discussion.stats.stars }}
        </button>
      </div>
    </div>

    <div class="comments-section">
      <h2>评论 ({{ comments.length }})</h2>
      <div class="comment-editor" v-if="userStore.isLoggedIn">
        <textarea
          v-model="newComment"
          placeholder="写下你的评论..."
          class="comment-input"
        ></textarea>
        <button @click="submitComment" class="submit-btn">发表评论</button>
      </div>
      <div v-else class="login-tip"><router-link to="/login">登录</router-link> 后参与评论</div>

      <div class="comment-list" v-if="comments && comments.length">
        <div v-for="comment in comments" :key="comment.id" class="comment">
          <div class="comment-header">
            <img :src="comment.author.avatar" alt="avatar" class="avatar" />
            <router-link
              :to="`/profile/${comment.author.username}`"
              class="username"
              :class="{ admin: comment.author.role === 'admin' }"
            >
              {{ comment.author.username }}
            </router-link>
            <span class="time">{{ formatTime(comment.createdAt) }}</span>
          </div>
          <div class="comment-content markdown-body" v-html="marked(comment.content)"></div>
        </div>
      </div>
      <div v-else class="no-comments">暂无评论</div>
    </div>
  </div>
  <div v-else class="loading">加载中...</div>
</template>

<style scoped>
.discuss-detail {
  max-width: 800px;
  margin: 80px auto 2rem;
  padding: 0 2rem;
}

.post-header {
  margin-bottom: 2rem;
}

.post-meta {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-top: 1rem;
  color: var(--text-gray);
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
}

.tag {
  background: rgba(var(--primary-color-rgb), 0.1);
  color: var(--primary-color);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
}

.actions {
  margin-left: auto;
  display: flex;
  gap: 0.5rem;
}

.edit-btn,
.delete-btn {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  border: none;
  cursor: pointer;
  transition: all 0.3s;
  font-size: 0.9rem;
}

.edit-btn {
  background: var(--primary-color);
  color: white;
  text-decoration: none;
}

.delete-btn {
  background: var(--error-color);
  color: white;
}

.edit-btn:hover,
.delete-btn:hover {
  opacity: 0.8;
}

.post-actions {
  margin-top: 2rem;
  display: flex;
  gap: 1rem;
}

.action-btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-light);
  cursor: pointer;
  transition: all 0.3s;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.action-btn.active {
  background: var(--primary-color);
  color: white;
}

.comments-section {
  margin-top: 3rem;
  padding-top: 2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.comment-editor {
  margin: 1rem 0 2rem;
}

.comment-input {
  width: 100%;
  min-height: 100px;
  padding: 1rem;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-light);
  margin-bottom: 1rem;
  resize: vertical;
}

.submit-btn {
  background: var(--primary-color);
  color: white;
  padding: 0.5rem 1.5rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.submit-btn:hover {
  opacity: 0.8;
}

.login-tip {
  text-align: center;
  padding: 1rem;
  color: var(--text-gray);
}

.login-tip a {
  color: var(--primary-color);
  text-decoration: none;
}

.comment {
  background: rgba(255, 255, 255, 0.05);
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.comment-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.loading {
  text-align: center;
  padding: 2rem;
  color: var(--text-gray);
}

/* 用户名样式 */
.username {
  color: var(--primary-color);
  text-decoration: none;
  transition: color 0.2s;
}

.username:hover {
  text-decoration: underline;
  opacity: 0.8;
}

/* 适配浅色主题 */
:global(.light-theme) .username {
  color: var(--primary-color-light);
}
</style>
