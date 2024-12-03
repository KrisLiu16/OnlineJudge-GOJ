<template>
  <div class="about-container">
    <div class="about-header">
      <img src="/images/logo/GOJ_LOGO.svg" alt="GOJ Logo" class="about-logo" />
    </div>

    <div class="about-content">
      <div class="about-section">
        <h2>我的愿景</h2>
        <p>{{ settings.about }}</p>
      </div>

      <div class="about-section">
        <h2>核心特性</h2>
        <div class="features-grid">
          <div class="feature-card">
            <div class="feature-icon">🚀</div>
            <h3>高性能评测</h3>
            <p>采用Linux容器技术，提供快速、安全的代码执行环境</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">⚡</div>
            <h3>高并发后端</h3>
            <p>采用Golang，提供高并发、高性能的后端服务</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">🛡️</div>
            <h3>安全可靠</h3>
            <p>严格的安全机制，保障评测系统的稳定运行</p>
          </div>
        </div>
      </div>


      <div class="about-section">
        <h2>联系我</h2>
        <div class="contact-info">
          <a :href="'mailto:' + settings.email" class="contact-link">
            <span class="icon">📧</span>
            {{ settings.email }}
          </a>
          <a :href="settings.github" target="_blank" class="contact-link">
            <span class="icon">📦</span>
            GitHub
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getPublicWebsiteSettings } from '@/api/website'
import type { WebsiteSettings } from '@/api/website'

const settings = ref<WebsiteSettings>({
  title: 'GO! Judge',
  subtitle: '快速、智能的在线评测系统',
  about: 'GOJ是一个高性能在线评测的平台，致力于提供快速、稳定的评测服务。',
  email: 'support@example.com',
  github: 'https://github.com/yourusername',
  icp: '',
  icpLink: '',
  feature1: '',
  feature2: '',
  feature3: ''
})

onMounted(async () => {
  try {
    const { data } = await getPublicWebsiteSettings()
    if (data) {
      settings.value = data
      localStorage.setItem('websiteSettings', JSON.stringify(data))
    }
  } catch (error) {
    console.error('获取网站设置失败:', error)
  }
})
</script>

<style scoped>
.about-container {
  margin: 70px auto auto;
  max-width: 1200px;
  padding: 2rem 1.5rem;
}

.about-header {
  text-align: center;
  margin-bottom: 3rem;
}

.about-logo {
  height: 4rem;
  margin-bottom: 1rem;
}

.about-header h1 {
  font-size: 2rem;
  color: var(--text-light);
  margin: 0;
}

.about-content {
  display: flex;
  flex-direction: column;
  gap: 3rem;
}

.about-section {
  background: var(--nav-bg-dark);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  padding: 2rem;
}

.about-section h2 {
  color: var(--primary-color);
  font-size: 1.5rem;
  margin-bottom: 1.5rem;
  font-weight: 500;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
}

.feature-card {
  padding: 1.5rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  text-align: center;
}

.feature-icon {
  font-size: 2rem;
  margin-bottom: 1rem;
}

.feature-card h3 {
  color: var(--text-light);
  margin-bottom: 0.5rem;
  font-size: 1.1rem;
}

.feature-card p {
  color: var(--text-light);
  opacity: 0.8;
  font-size: 0.9rem;
  line-height: 1.5;
}

.tech-stack {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
  justify-content: center;
}

.tech-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.tech-logo {
  height: 2.5rem;
  opacity: 0.9;
}

.tech-item span {
  color: var(--text-light);
  font-size: 0.9rem;
}

.contact-info {
  display: flex;
  gap: 2rem;
  justify-content: center;
  flex-wrap: wrap;
}

.contact-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text-light);
  text-decoration: none;
  font-size: 0.9rem;
  opacity: 0.8;
  transition: opacity 0.2s, color 0.2s;
}

.contact-link:hover {
  color: var(--primary-color);
  opacity: 1;
}

.icon {
  font-size: 1.2rem;
}

@media (max-width: 768px) {
  .about-container {
    padding: 1rem;
  }

  .about-header {
    margin-bottom: 2rem;
  }

  .about-section {
    padding: 1.5rem;
  }

  .features-grid {
    grid-template-columns: 1fr;
  }

  .tech-stack {
    gap: 1.5rem;
  }

  .contact-info {
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }
}
</style>
