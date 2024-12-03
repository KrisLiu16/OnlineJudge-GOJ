<template>
  <div class="home">
    <div class="hero-section">
      <h1 class="title">
        <div class="logo-container">
          <span class="system-text glowing-text">{{ settings.title }}</span>
        </div>
      </h1>
      <p class="subtitle">{{ settings.subtitle }}</p>
      <div class="cta-buttons">
        <router-link to="/problems" class="cta-primary neon-button">
          <span></span>
          <span></span>
          <span></span>
          <span></span>
          GO!
        </router-link>
      </div>
    </div>

    <div class="features-section">
      <div class="feature-card" v-html="settings.feature1"></div>
      <div class="feature-card" v-html="settings.feature2"></div>
      <div class="feature-card" v-html="settings.feature3"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { http } from '@/utils/http'
import type { WebsiteSettings } from '@/api/website'

const settings = ref<WebsiteSettings>({
  title: 'GO! Judge',
  subtitle: '快速、智能的在线评测系统',
  feature1: '<div class="feature-icon"><span class="icon-wrapper">📚</span></div><h3>丰富的题库</h3><p>包含各种难度的编程题目，从入门到进阶</p>',
  feature2: '<div class="feature-icon"><span class="icon-wrapper">🚀</span></div><h3>实时评测</h3><p>快速的代码执行和结果反馈</p>',
  feature3: '<div class="feature-icon"><span class="icon-wrapper">👥</span></div><h3>社区讨论</h3><p>与其他同学交流学习心得</p>',
  about: '',
  email: '',
  github: '',
  icp: '',
  icpLink: ''
})

onMounted(async () => {
  try {
    const response = await http.get('/website/settings')
    if (response.code === 200) {
      settings.value = response.data
      localStorage.setItem('websiteSettings', JSON.stringify(response.data))
    }
  } catch (error) {
    console.error('获取网站设置失败:', error)
    const savedSettings = localStorage.getItem('websiteSettings')
    if (savedSettings) {
      settings.value = JSON.parse(savedSettings)
    }
  }
})
</script>

<style scoped>
.home {
  margin: 70px auto auto;
  position: relative;
  overflow: hidden;
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
  font-size: 4rem;
  margin-bottom: 1rem;
  color: var(--text-light);
  text-shadow: 0 0 10px rgba(var(--primary-color-rgb), 0.5);
}

.subtitle {
  font-size: 1.5rem;
  color: var(--text-light);
  opacity: 0.8;
  margin-bottom: 2rem;
  text-shadow: 0 0 5px rgba(255, 255, 255, 0.3);
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
  background: transparent;
  color: var(--primary-color);
}

.cta-primary:hover::after {
  display: none;
}

.cta-secondary {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(5px);
  border: 1px solid rgba(255, 255, 255, 0.2);
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
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(var(--primary-color-rgb), 0.1);
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
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

.glowing-text {
  text-shadow:
    0 0 10px #00ADD8,
    0 0 20px #00ADD8,
    0 0 30px #00ADD8;
  animation: glow 2s ease-in-out infinite alternate;
}

@keyframes glow {
  from {
    text-shadow:
      0 0 10px #00ADD8,
      0 0 20px #00ADD8;
  }
  to {
    text-shadow:
      0 0 15px var(#0000),
      0 0 25px var(#0000),
      0 0 35px var(#0000);
  }
}

.feature-card {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
  border: 1px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-10px);
  border-color: var(--primary-color);
  box-shadow: 0 5px 15px rgba(var(--primary-color-rgb), 0.2);
}

.icon-wrapper {
  display: inline-block;
  padding: 1rem;
  background: rgba(var(--primary-color-rgb), 0.1);
  border-radius: 50%;
  transition: transform 0.3s ease;
}

.feature-card:hover .icon-wrapper {
  transform: scale(1.1) rotate(5deg);
}

.neon-button {
  position: relative;
  display: inline-block;
  padding: 1rem 2rem;
  color: #00ADD8 !important;
  background: transparent;
  text-transform: uppercase;
  letter-spacing: 2px;
  overflow: hidden;
  transition: 0.5s;
  border: 1px solid #00ADD8;
  box-shadow: none;
}

.neon-button:hover {
  background: #00ADD8;
  color: var(--bg-dark) !important;
  box-shadow:
    0 0 5px #00ADD8,
    0 0 25px #00ADD8,
    0 0 50px #00ADD8,
    0 0 100px #00ADD8;
}

.neon-button span {
  position: absolute;
  display: block;
}

.neon-button span:nth-child(1) {
  top: 0;
  left: -100%;
  width: 100%;
  height: 2px;
  background: linear-gradient(90deg, transparent, #00ADD8);
  animation: btn-anim1 1s linear infinite;
}

@keyframes btn-anim1 {
  0% {
    left: -100%;
  }
  50%,
  100% {
    left: 100%;
  }
}

.neon-button span:nth-child(2) {
  top: -100%;
  right: 0;
  width: 2px;
  height: 100%;
  background: linear-gradient(180deg, transparent, #00ADD8);
  animation: btn-anim2 1s linear infinite;
  animation-delay: 0.25s;
}

@keyframes btn-anim2 {
  0% {
    top: -100%;
  }
  50%,
  100% {
    top: 100%;
  }
}

.neon-button span:nth-child(3) {
  bottom: 0;
  right: -100%;
  width: 100%;
  height: 2px;
  background: linear-gradient(270deg, transparent, #00ADD8);
  animation: btn-anim3 1s linear infinite;
  animation-delay: 0.5s;
}

@keyframes btn-anim3 {
  0% {
    right: -100%;
  }
  50%,
  100% {
    right: 100%;
  }
}

.neon-button span:nth-child(4) {
  bottom: -100%;
  left: 0;
  width: 2px;
  height: 100%;
  background: linear-gradient(360deg, transparent, #00ADD8);
  animation: btn-anim4 1s linear infinite;
  animation-delay: 0.75s;
}

@keyframes btn-anim4 {
  0% {
    bottom: -100%;
  }
  50%,
  100% {
    bottom: 100%;
  }
}
</style>
