<script setup lang="ts">
import HumanVerify from '@/components/HumanVerify.vue'
import { ref } from 'vue'
import { useUserStore } from '@/stores/modules/user'
import { useRouter } from 'vue-router'
import { AxiosError } from 'axios'

// 修改响应数据的接口定义
interface ApiResponse {
  message: string
}

const userStore = useUserStore()
const router = useRouter()

const account = ref('')
const password = ref('')
const rememberMe = ref(false)
const errorMessage = ref('')
const loading = ref(false)
const isVerified = ref(false)
const accountFocus = ref(false)
const passwordFocus = ref(false)
const showPassword = ref(false)

const handleVerifySuccess = () => {
  isVerified.value = true
}

const handleLogin = async () => {
  console.log('=== SignInView.handleLogin Start ===')
  console.log('Form data:', {
    account: account.value,
    password: '***',
    isVerified: isVerified.value,
  })

  if (!isVerified.value) {
    errorMessage.value = '请完成人机验证'
    return
  }

  if (!account.value || !password.value) {
    errorMessage.value = '请填写所有必填项'
    return
  }

  // 添加基本的账户格式验证
  if (account.value.length < 2) {
    errorMessage.value = '账户长度至少为2个字符'
    return
  }

  try {
    loading.value = true
    await userStore.login(account.value, password.value)
    router.push('/')
  } catch (error) {
    const apiError = error as AxiosError<ApiResponse>
    errorMessage.value = apiError.response?.data?.message || '登录失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="sign-in-container">
    <div class="sign-in-card">
      <div class="card-header">
        <h2>欢迎回来</h2>
        <p class="subtitle">登录以继续您的编程之旅</p>
      </div>

      <form class="sign-in-form" @submit.prevent="handleLogin">
        <div v-if="errorMessage" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ errorMessage }}
        </div>

        <div class="form-group">
          <label for="account">
            <i class="fas fa-user"></i>
            账户
          </label>
          <input
            type="text"
            id="account"
            v-model="account"
            placeholder="请输入用户名或邮箱"
            required
            :class="{ 'input-focus': accountFocus }"
            @focus="accountFocus = true"
            @blur="accountFocus = false"
          />
        </div>

        <div class="form-group">
          <label for="password">
            <i class="fas fa-lock"></i>
            密码
          </label>
          <div class="password-input">
            <input
              :type="showPassword ? 'text' : 'password'"
              id="password"
              v-model="password"
              placeholder="请输入密码"
              required
              :class="{ 'input-focus': passwordFocus }"
              @focus="passwordFocus = true"
              @blur="passwordFocus = false"
            />
            <i
              class="fas"
              :class="showPassword ? 'fa-eye-slash' : 'fa-eye'"
              @click="showPassword = !showPassword"
            ></i>
          </div>
        </div>

        <HumanVerify @verify-success="handleVerifySuccess" :threshold="90" />

        <div class="form-options">
          <label class="remember-me">
            <input type="checkbox" v-model="rememberMe" />
            <span class="checkmark"></span>
            记住我
          </label>
          <router-link to="/forgot-password" class="forgot-password"> 忘记密码？ </router-link>
        </div>

        <button type="submit" class="submit-btn" :disabled="loading" :class="{ loading: loading }">
          <span class="btn-content">
            <i class="fas fa-sign-in-alt" v-if="!loading"></i>
            <i class="fas fa-circle-notch fa-spin" v-else></i>
            {{ loading ? '登录中...' : '登录' }}
          </span>
        </button>
      </form>

      <div class="sign-up-link">
        还没有账号？
        <router-link to="/sign-up" class="highlight-link">立即注册</router-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sign-in-container {
  min-height: calc(100vh - 64px);
  padding: 80px 2rem 2rem;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, var(--bg-dark) 0%, var(--nav-bg-dark) 100%);
}

.sign-in-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(20px);
  padding: 3rem;
  border-radius: 20px;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: transform 0.3s ease;
}

.sign-in-card:hover {
  transform: translateY(-5px);
}

.card-header {
  text-align: center;
  margin-bottom: 2rem;
}

.card-header h2 {
  font-size: 2rem;
  color: var(--text-light);
  margin-bottom: 0.5rem;
}

.subtitle {
  color: var(--text-gray);
  font-size: 0.9rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  color: var(--text-light);
  font-size: 0.9rem;
}

.form-group input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-light);
  transition: all 0.3s ease;
}

.input-focus {
  border-color: var(--primary-color) !important;
  box-shadow: 0 0 0 2px rgba(var(--primary-rgb), 0.2);
}

.password-input {
  position: relative;
}

.password-input i {
  position: absolute;
  right: 1rem;
  top: 50%;
  transform: translateY(-50%);
  cursor: pointer;
  color: var(--text-gray);
  transition: color 0.3s ease;
}

.password-input i:hover {
  color: var(--primary-color);
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 1.5rem 0;
}
.remember-me {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: var(--text-gray);
}

.submit-btn {
  width: 100%;
  padding: 0.75rem;
  border: none;
  border-radius: 8px;
  background: var(--primary-color);
  color: var(--bg-dark);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.submit-btn:hover {
  background: var(--primary-hover);
  transform: translateY(-2px);
}

.btn-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  background: rgba(255, 77, 79, 0.1);
  border-radius: 8px;
  color: #ff4d4f;
  margin-bottom: 1.5rem;
}

.highlight-link {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.3s ease;
}

.highlight-link:hover {
  color: var(--primary-hover);
  text-decoration: underline;
}
</style>
