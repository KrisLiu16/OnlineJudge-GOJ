<template>
  <div class="app-container" :class="{ 'light-theme': isLightTheme }">
    <div v-if="appStore.notification.show" class="notification" :class="appStore.notification.type">
      {{ appStore.notification.message }}
    </div>
    <nav class="main-nav">
      <div class="nav-left">
        <router-link to="/" class="logo">
          <img src="/images/logo/GOJ_LOGO.svg" alt="goj" class="goj-logo" />
        </router-link>
      </div>
      <div v-if="!isMobileMenuClosed" class="mobile-overlay" @click="toggleMobileMenu"></div>
      <div class="nav-center" :class="{ 'mobile-hidden': isMobileMenuClosed }">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          @click="handleNavClick"
        >
          {{ item.name }}
        </router-link>
      </div>
      <button
        class="mobile-menu-btn"
        @click="toggleMobileMenu"
        :class="{ active: !isMobileMenuClosed }"
      >
        <span class="menu-icon"></span>
      </button>
      <div class="nav-right">
        <div class="theme-selector">
          <button class="theme-toggle" @click="cycleTheme">
            <span v-if="currentTheme === 'dark'">🌞</span>
            <span v-else-if="currentTheme === 'light'">🌙</span>
            <span v-else>🎨</span>
          </button>
          <div v-if="showColorPicker" class="color-picker-dropdown">
            <input type="color" v-model="customColor" @input="applyCustomTheme" />
          </div>
        </div>

        <template v-if="!userStore.isAuthenticated">
          <router-link to="/sign-in" class="sign-in">登录</router-link>
          <router-link to="/sign-up" class="sign-up">注册</router-link>
        </template>
        <div v-else class="user-menu" @click="toggleUserMenu" ref="userMenuRef">
          <img
            :src="userStore.userAvatar || '/images/avatars/default-avatar.png'"
            :alt="userStore.userInfo?.username"
            class="user-avatar"
          />
          <div v-show="showUserMenu" class="user-dropdown">
            <router-link to="/profile" class="dropdown-item">个人主页</router-link>
            <router-link to="/settings" class="dropdown-item">设置</router-link>
            <router-link
              v-if="userStore.userInfo?.role === 'admin'"
              to="/admin"
              class="dropdown-item admin-item"
            >
              后台管理
            </router-link>
            <button @click="handleLogout" class="dropdown-item">退出登录</button>
          </div>
        </div>
      </div>
    </nav>
    <main class="main-content">
      <router-view />
    </main>
    <FooterComponent v-if="showFooter" />
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, watch, onMounted, onUnmounted, computed, provide } from 'vue'
import FooterComponent from './components/FooterComponent.vue'
import { useUserStore } from './stores/modules/user'
import { useAppStore } from './stores/modules/app'
import { useRouter, useRoute } from 'vue-router'
import { markdownStyles } from './utils/markdown'
import mitt from 'mitt'

export default defineComponent({
  components: {
    FooterComponent,
  },
  setup() {
    const appStore = useAppStore()
    const isLightTheme = ref(localStorage.getItem('theme') === 'light')
    const userStore = useUserStore()
    const router = useRouter()
    const showUserMenu = ref(false)
    const userMenuRef = ref<HTMLElement | null>(null)
    const isMobileMenuClosed = ref(true)
    const isMobile = ref(false)
    const route = useRoute()
    const emitter = mitt()
    const currentTheme = ref(localStorage.getItem('theme') || 'dark')
    const customColor = ref('#00ADD8')
    const showColorPicker = ref(false)

    const toggleTheme = () => {
      isLightTheme.value = !isLightTheme.value
      localStorage.setItem('theme', isLightTheme.value ? 'light' : 'dark')
      emitter.emit('theme-change', isLightTheme.value)
    }

    const toggleUserMenu = () => {
      showUserMenu.value = !showUserMenu.value
    }
    const handleLogout = () => {
      userStore.$reset()
      showUserMenu.value = false
      router.push('/')
    }

    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.value && !userMenuRef.value.contains(event.target as Node)) {
        showUserMenu.value = false
      }
    }

    const checkMobile = () => {
      isMobile.value = window.innerWidth <= 768
      if (!isMobile.value) {
        isMobileMenuClosed.value = true
      }
    }

    const toggleMobileMenu = () => {
      isMobileMenuClosed.value = !isMobileMenuClosed.value
    }

    watch(
      isLightTheme,
      (newValue: boolean) => {
        document.body.className = newValue ? 'light-theme' : ''
      },
      { immediate: true },
    )

    onMounted(async () => {
      document.addEventListener('click', handleClickOutside)
      if (userStore.isAuthenticated) {
        try {
          await userStore.fetchUserProfile()
        } catch (error) {
          console.error('Failed to fetch user profile:', error)
        }
      }
      checkMobile()
      window.addEventListener('resize', checkMobile)
      const styleEl = document.createElement('style')
      styleEl.innerHTML = markdownStyles
      document.head.appendChild(styleEl)
      applyTheme(currentTheme.value)
    })

    onUnmounted(() => {
      document.removeEventListener('click', handleClickOutside)
    })

    const navItems = [
      { path: '/', name: '首页' },
      { path: '/problems', name: '题库' },
      { path: '/contests', name: '比赛' },
      { path: '/submissions', name: '测评' },
      // { path: '/groups', name: '群组' },
      { path: '/discuss', name: '讨论' },
      { path: '/rank', name: '排名' },
    ]

    const handleNavClick = () => {
      if (isMobile.value) {
        isMobileMenuClosed.value = true
      }
    }

    const showFooter = computed(() => {
      return route.path === '/'
    })

    provide('isLightTheme', isLightTheme)

    const cycleTheme = () => {
      const themes = ['dark', 'light', 'custom']
      const currentIndex = themes.indexOf(currentTheme.value)
      const nextIndex = (currentIndex + 1) % themes.length
      currentTheme.value = themes[nextIndex]

      if (currentTheme.value === 'custom') {
        showColorPicker.value = true
      } else {
        showColorPicker.value = false
        localStorage.setItem('theme', currentTheme.value)
        applyTheme(currentTheme.value)
      }
    }

    const applyCustomTheme = () => {
      document.body.style.background = `linear-gradient(135deg, ${customColor.value} 0%, #2f495e 100%)`
      document.body.style.setProperty('--primary-color', customColor.value)
      localStorage.setItem('customColor', customColor.value)
    }

    const applyTheme = (theme: string) => {
      if (theme === 'custom') {
        const savedColor = localStorage.getItem('customColor')
        if (savedColor) {
          customColor.value = savedColor
          applyCustomTheme()
        }
      } else {
        document.body.style.background = ''
        document.body.style.setProperty('--primary-color', '#42b983')
        document.body.className = theme === 'light' ? 'light-theme' : ''
      }
    }

    return {
      appStore,
      isLightTheme,
      toggleTheme,
      userStore,
      showUserMenu,
      userMenuRef,
      toggleUserMenu,
      handleLogout,
      isMobileMenuClosed,
      toggleMobileMenu,
      navItems,
      handleNavClick,
      showFooter,
      cycleTheme,
      showColorPicker,
      customColor,
      applyCustomTheme,
      applyTheme,
      currentTheme,
    }
  },
})
</script>

<style>
:root {
  --primary-color: #42b983;
  --bg-dark: #1a1a1a;
  --bg-light: #42b983;
  --text-light: #ffffff;
  --text-dark: #2c3e50;
  --text-gray: #8a8a8a;
  --nav-bg-dark: rgba(0, 0, 0, 0.2);
  --nav-bg-light: rgba(255, 255, 255, 0.1);
}

body {
  margin: 0;
  background: linear-gradient(135deg, #1a1a1a 0%, #2f495e 100%);
  color: var(--text-light);
  transition: all 0.3s ease;
}

body.light-theme {
  background: linear-gradient(135deg, #42b983 0%, #2f495e 100%);
  color: var(--text-dark);
}

.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background-color: var(--nav-bg-dark);
  backdrop-filter: blur(10px);
  position: fixed;
  width: 100%;
  box-sizing: border-box;
  top: 0;
  z-index: 1000;
}

.light-theme .main-nav {
  background-color: var(--nav-bg-light);
}

.nav-left .logo {
  display: flex;
  align-items: center;
}

.goj-logo {
  height: 2rem;
  width: auto;
}

.nav-center {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 2rem;
}

.nav-center a {
  color: var(--text-light);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.3s;
}

.nav-center a:hover {
  color: var(--primary-color);
}

.nav-right {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.theme-toggle {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.2rem;
  padding: 0.5rem;
  color: var(--text-light);
  transition: transform 0.3s;
}

.theme-toggle:hover {
  transform: rotate(30deg);
}

.nav-right a {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  text-decoration: none;
  transition: all 0.3s;
}

.sign-in {
  color: var(--text-light);
}

.sign-up {
  background-color: var(--primary-color);
  color: var(--bg-dark);
}

.sign-up:hover {
  opacity: 0.9;
}

.light-theme .nav-center a,
.light-theme .sign-in,
.light-theme .theme-toggle {
  color: var(--text-light);
}

.main-content {
  flex: 1;
}

.user-menu {
  position: relative;
  cursor: pointer;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.user-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  background: var(--nav-bg-dark);
  backdrop-filter: blur(10px);
  border-radius: 4px;
  padding: 0.5rem 0;
  min-width: 150px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.dropdown-item {
  display: block;
  padding: 0.5rem 1rem;
  color: var(--text-light);
  text-decoration: none;
  transition: background-color 0.3s;
}

.dropdown-item:hover {
  background-color: var(--nav-bg-light);
}

button.dropdown-item {
  width: 100%;
  text-align: left;
  background: none;
  border: none;
  font: inherit;
  cursor: pointer;
}

.nav-left,
.nav-right {
  z-index: 1;
}

.notification {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  padding: 1rem 2rem;
  border-radius: 4px;
  z-index: 1001;
  animation: slideUp 0.3s ease-out;
  color: white;
  min-width: 200px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

@keyframes slideUp {
  from {
    transform: translate(-50%, 100%);
    opacity: 0;
  }
  to {
    transform: translate(-50%, 0);
    opacity: 1;
  }
}

.notification.success {
  background-color: #00ADD8;
}

.notification.error {
  background-color: #ff4d4f;
}

.notification.info {
  background-color: #1890ff;
}

.notification.warning {
  background-color: #faad14;
}

.mobile-menu-btn {
  display: none;
  background: none;
  border: none;
  color: var(--text-light);
  font-size: 1.5rem;
  padding: 0.5rem;
  cursor: pointer;
  z-index: 1000;
}

@media (max-width: 768px) {
  .nav-center {
    position: fixed;
    left: 50%;
    top: 64px;
    transform: translateX(-50%) translateY(-100%);
    width: 200px;
    background: #2c3e50;
    border-radius: 0 0 12px 12px;
    flex-direction: column;
    padding: 8px;
    gap: 4px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    z-index: 100;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    pointer-events: none;
  }

  .light-theme .nav-center {
    background: white;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }

  .nav-center:not(.mobile-hidden) {
    transform: translateX(-50%) translateY(0);
    pointer-events: auto;
  }

  .mobile-hidden {
    transform: translateX(-50%) translateY(-100%);
    opacity: 0;
    pointer-events: none;
  }

  .nav-center a {
    color: var(--text-light);
    text-decoration: none;
    padding: 10px 16px;
    text-align: center;
    border-radius: 6px;
    transition: all 0.2s;
    font-size: 0.95rem;
    width: 100%;
    box-sizing: border-box;
  }

  .light-theme .nav-center a {
    color: var(--text-dark);
  }

  .nav-center a:hover,
  .nav-center a.router-link-active {
    background: var(--primary-color);
    color: white;
  }

  .mobile-overlay {
    position: fixed;
    top: 64px;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.3);
    z-index: 99;
    opacity: 0;
    animation: fadeIn 0.2s forwards;
  }

  .mobile-menu-btn {
    display: block;
    position: absolute;
    right: 1rem;
    width: 40px;
    height: 40px;
    padding: 8px;
    z-index: 101;
    border-radius: 8px;
  }

  .mobile-menu-btn:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .menu-icon {
    position: relative;
    display: block;
    width: 20px;
    height: 2px;
    background: var(--text-light);
    margin: 10px auto;
    transition: all 0.3s;
  }

  .menu-icon::before,
  .menu-icon::after {
    content: '';
    position: absolute;
    width: 20px;
    height: 2px;
    background: var(--text-light);
    transition: all 0.3s;
  }

  .menu-icon::before {
    top: -6px;
  }

  .menu-icon::after {
    top: 6px;
  }

  .mobile-menu-btn.active .menu-icon {
    background: transparent;
  }

  .mobile-menu-btn.active .menu-icon::before {
    transform: rotate(45deg);
    top: 0;
  }

  .mobile-menu-btn.active .menu-icon::after {
    transform: rotate(-45deg);
    top: 0;
  }

  .nav-right {
    margin-right: 3.5rem;
  }
}

.admin-item {
  color: var(--primary-color) !important;
  font-weight: 500;
}

.light-theme .admin-item {
  color: var(--primary-color) !important;
}

.admin-item:hover {
  background: rgba(var(--primary-color-rgb), 0.1) !important;
}

.theme-selector {
  position: relative;
}

.color-picker-dropdown {
  position: absolute;
  top: 50%;
  right: 100%;
  padding: 10px;
  background: var(--nav-bg-dark);
  border-radius: 4px;
  margin-right: 5px;
  transform: translateY(-50%);
}

.color-picker-dropdown input[type='color'] {
  width: 40px;
  height: 40px;
  padding: 0;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
</style>
