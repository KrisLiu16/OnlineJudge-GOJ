<template>
  <div class="discuss">
    <div class="discuss-header">
      <h1 class="glowing-text">讨论区</h1>
      <router-link to="/discuss/new" class="new-post-btn">
        <i class="fas fa-plus"></i> 发布讨论
      </router-link>
    </div>

    <div v-if="isLoading" class="loading">
      <div class="loading-spinner"></div>
      <div class="loading-text">加载中...</div>
    </div>

    <div v-else class="post-list">
      <div class="post-grid">
        <router-link
          v-for="discussion in discussions"
          :key="discussion.id"
          :to="`/discuss/${discussion.id}`"
          class="post-card"
        >
          <div class="post-header">
            <div class="user-info">
              <img :src="discussion.author.avatar" alt="avatar" class="avatar" />
              <div class="user-meta">
                <span class="username" :class="{ admin: discussion.author.role === 'admin' }">
                  {{ discussion.author.username }}
                </span>
                <span class="time">{{ formatTime(discussion.createdAt) }}</span>
              </div>
            </div>
            <span :class="['category-badge', getCategoryClass(discussion.category)]">
              {{ getCategoryText(discussion.category) }}
            </span>
          </div>

          <div class="post-body">
            <h3 class="post-title">{{ discussion.title }}</h3>
            <p class="post-content markdown-body" v-html="discussion.renderedContent"></p>
          </div>

          <div class="post-footer">
            <div class="stats">
              <span class="stat-item">
                <i class="far fa-comment"></i>
                <span class="stat-value">{{ discussion.stats.comments }}</span>
              </span>
              <span class="stat-item">
                <i class="far fa-thumbs-up"></i>
                <span class="stat-value">{{ discussion.stats.likes }}</span>
              </span>
              <span class="stat-item">
                <i class="far fa-star"></i>
                <span class="stat-value">{{ discussion.stats.stars }}</span>
              </span>
            </div>
          </div>
        </router-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.discuss {
  min-height: calc(100vh - 64px - 150px);
  padding: 80px 2rem 2rem;
  max-width: 1000px;
  margin: 0 auto;
}

.discuss-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.glowing-text {
  color: var(--text-primary);
  text-shadow: 0 0 10px var(--primary-color);
  font-size: 2rem;
  margin: 0;
}

.new-post-btn {
  background: var(--primary-color);
  color: white;
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  text-decoration: none;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
}

.new-post-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(var(--primary-color-rgb), 0.2);
}

.post-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.post-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 1.5rem;
  padding: 1rem;
}

.post-card {
  display: flex;
  flex-direction: column;
  padding: 1.5rem;
  height: 100%;
  min-height: 300px;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.3s ease;
}

.post-card:hover {
  transform: translateY(-4px);
  background: rgba(255, 255, 255, 0.08);
  border-color: var(--primary-color);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
}

.post-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-meta {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(var(--primary-color-rgb), 0.3);
}

.username {
  font-weight: 600;
  color: var(--primary-color);
}

.time {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.post-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.post-title {
  font-size: 1.3rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.4;
}

.post-content {
  color: var(--text-secondary);
  font-size: 0.95rem;
  line-height: 1.6;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  line-clamp: 3;
  -webkit-box-orient: vertical;
  margin: 0;
}

.post-footer {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.stats {
  display: flex;
  gap: 1.5rem;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.stat-value {
  font-weight: 500;
}

/* 加载动画样式 */
.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 12px;
  margin: 2rem 0;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s linear infinite;
}

.loading-text {
  margin-top: 1rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .discuss {
    padding: 60px 1rem 1rem;
  }

  .discuss-header {
    flex-direction: column;
    gap: 1rem;
    align-items: flex-start;
  }

  .post-grid {
    grid-template-columns: 1fr;
  }

  .post-card {
    min-height: auto;
  }

  .post-header {
    flex-direction: column;
    gap: 1rem;
  }

  .category-badge {
    align-self: flex-start;
  }
}

/* 分类徽章基础样式 */
.category-badge {
  padding: 0.4rem 1.2rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
  color: white;
  background: linear-gradient(135deg, var(--start-color), var(--end-color));
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  white-space: nowrap;
}

/* 讨论徽章 */
.category-discussion {
  --start-color: #4facfe;
  --end-color: #00f2fe;
}

/* 题解徽章 */
.category-solution {
  --start-color: #43e97b;
  --end-color: #38f9d7;
}

/* 公告徽章 */
.category-announcement {
  --start-color: #fa709a;
  --end-color: #fee140;
}

/* 提问徽章 */
.category-question {
  --start-color: #e0c3fc;
  --end-color: #8ec5fc;
}

.category-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

/* 动画效果 */
.post-grid {
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

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { renderMarkdown } from '@/utils/markdown'

interface Discussion {
  id: string
  title: string
  content: string
  renderedContent?: string
  category: string
  createdAt: string
  author: {
    id: number
    username: string
    avatar: string
    role?: string
  }
  stats: {
    likes: number
    comments: number
    stars: number
  }
}

interface RawDiscussion {
  id: string
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: string | null
  UserID: number
  Title: string
  Content: string
  Category: string
  Likes: number
  Comments: number
  Stars: number
  author: {
    id: number
    username: string
    avatar: string
    role?: string
  }
  stats: {
    likes: number
    comments: number
    stars: number
  }
}

const discussions = ref<Discussion[]>([])
const isLoading = ref(true)

const fetchDiscussions = async () => {
  try {
    const response = await fetch('/api/discussions')
    if (!response.ok) throw new Error('获取讨论列表失败')
    const data = await response.json()

    discussions.value = await Promise.all(
      data.discussions.map(async (d: RawDiscussion) => {
        const renderedContent = await renderMarkdown(d.Content?.slice(0, 200) + '...' || '')

        return {
          id: d.id,
          title: d.Title || '',
          content: d.Content || '',
          renderedContent,
          category: d.Category || '',
          createdAt: d.CreatedAt || '',
          author: {
            id: d.author?.id || 0,
            username: d.author?.username || '',
            avatar: d.author?.avatar || '/images/avatars/default-avatar.png',
            role: d.author?.role || 'user',
          },
          stats: {
            likes: d.Likes || 0,
            comments: d.Comments || 0,
            stars: d.Stars || 0,
          },
        }
      }),
    )
  } catch (error) {
    console.error('获取讨论列表失败:', error)
  } finally {
    isLoading.value = false
  }
}

const formatTime = (time: string) => {
  const date = new Date(time)
  return date.toLocaleString()
}

// 添加分类处理函数
const getCategoryText = (category: string) => {
  const categoryMap: Record<string, string> = {
    discussion: '讨论',
    solution: '题解',
    announcement: '公告',
    question: '提问',
  }
  return categoryMap[category] || category
}

const getCategoryClass = (category: string) => {
  return `category-${category.toLowerCase()}`
}

onMounted(() => {
  fetchDiscussions()
})
</script>
