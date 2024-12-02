<template>
  <footer class="footer">
    <div class="footer-content">
      <div class="footer-info">
        <div class="footer-links">
        <a href="/about">关于我们</a> |
        <a :href="settings.github" target="_blank">GitHub</a> |
        <a :href="'mailto:' + settings.email">联系我们</a>
      </div>
        <p>Copyright ©{{ new Date().getFullYear() }}, {{ settings.title }}</p>
        <p v-if="settings.icp">
          <a :href="settings.icpLink" target="_blank">{{ settings.icp }}</a>
        </p>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.footer {
  width: 100%;
  background: var(--nav-bg-dark);
  backdrop-filter: blur(10px);
  padding: 1.5rem;
  margin-top: auto;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}

.footer-left {
  text-align: center;
  margin-bottom: 0.5rem;
}

.logo-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.goj-logo {
  height: 2em;
  opacity: 0.9;
  transition: opacity 0.2s;
}

.goj-logo:hover {
  opacity: 1;
}

.logo-section p {
  margin: 0;
  font-size: 0.8rem;
  color: var(--text-light);
  opacity: 0.8;
  max-width: 600px;
  text-align: center;
}

.footer-links {
  text-align: center;
  color: var(--text-light);
  opacity: 0.5;
  font-size: 0.8rem;
}

.footer-links a {
  color: var(--text-light);
  text-decoration: none;
  padding: 0 0.5rem;
  transition: color 0.2s, opacity 0.2s;
  opacity: 0.8;
}

.footer-links a:hover {
  color: var(--primary-color);
  opacity: 1;
}

.footer-info {
  text-align: center;
  color: var(--text-light);
  opacity: 0.8;
  font-size: 0.75rem;
  line-height: 1.5;
}

.footer-info p {
  margin: 0;
}

.footer-info a {
  color: var(--text-light);
  text-decoration: none;
  transition: color 0.2s, opacity 0.2s;
  opacity: 0.7;
}

.footer-info a:hover {
  color: var(--primary-color);
  opacity: 1;
}

@media (max-width: 768px) {
  .footer {
    padding: 1rem;
  }

  .footer-links {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.5rem;
  }

  .logo-section p {
    padding: 0 1rem;
  }
}
</style>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const settings = ref({
  title: 'GO! Judge',
  subtitle: '快速、智能的在线评测系统',
  icp: '',
  icpLink: '',
  about: 'GOJ是一个高性能在线评测的平台，致力于提供快速、稳定的评测服务。',
  email: 'support@example.com',
  github: 'https://github.com/yourusername'
})

onMounted(async () => {
  try {
    const savedSettings = localStorage.getItem('websiteSettings')
    if (savedSettings) {
      settings.value = JSON.parse(savedSettings)
    }

    const response = await fetch('/api/website/settings')
    const data = await response.json()
    if (data.code === 200) {
      settings.value = data.data
    }
  } catch (error) {
    console.error('Failed to load website settings:', error)
  }
})
</script>
