<template>
  <div class="not-found">
    <div class="hero-section">
      <h1 class="title">
        <div class="logo-container">
          <span class="system-text">404</span>
        </div>
      </h1>
      <p class="subtitle">抱歉，您访问的页面不存在，或正在开发中</p>
      <div class="cta-buttons">
        <router-link to="/" class="cta-primary">返回首页</router-link>
        <router-link to="/problems" class="cta-secondary">去刷题</router-link>
      </div>
    </div>

    <div class="features-section">
      <div class="feature-card">
        <div class="feature-icon">🏠</div>
        <h3>返回首页</h3>
        <p>回到系统主页开始探索</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">📚</div>
        <h3>题目列表</h3>
        <p>浏览所有可用的编程题目</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">❓</div>
        <h3>需要帮助？</h3>
        <p>访问帮助中心获取支持</p>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'NotFoundView',
})
</script>

<style scoped>
.not-found {
  min-height: calc((100vh - 64px - 150px) * 1.5);
  padding-top: 64px;
}

.hero-section {
  height: 80vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  background: var(--nav-bg-dark);
  backdrop-filter: blur(10px);
  padding: 2rem;
}

.title {
  font-size: 6rem;
  margin-bottom: 1rem;
  color: var(--primary-color);
}

.subtitle {
  font-size: 1.5rem;
  color: var(--text-light);
  opacity: 0.8;
  margin-bottom: 2rem;
}

.cta-buttons {
  display: flex;
  gap: 1rem;
}

.cta-buttons a {
  padding: 1rem 2rem;
  border-radius: 8px;
  text-decoration: none;
  font-weight: bold;
  transition: all 0.3s ease;
}

.cta-primary {
  background-color: var(--primary-color);
  color: var(--bg-dark);
}

.cta-primary:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(var(--primary-color-rgb), 0.3);
}

.cta-secondary {
  border: 2px solid var(--text-light);
  color: var(--text-light);
}

.cta-secondary:hover {
  transform: translateY(-4px);
  background: rgba(255, 255, 255, 0.1);
  box-shadow: 0 4px 12px rgba(255, 255, 255, 0.1);
}

.features-section {
  padding: 4rem 2rem;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  background: var(--nav-bg-dark);
  backdrop-filter: blur(10px);
}

.feature-card {
  background: rgba(255, 255, 255, 0.05);
  padding: 2rem;
  border-radius: 12px;
  text-align: center;
  transition: transform 0.3s;
}

.feature-card:hover {
  transform: translateY(-5px);
}

.feature-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.feature-card h3 {
  color: var(--primary-color);
  margin-bottom: 1rem;
}

.feature-card p {
  color: var(--text-light);
  opacity: 0.8;
}
</style>
