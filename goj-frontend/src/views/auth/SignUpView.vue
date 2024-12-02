<template>
  <div class="sign-up-container">
    <div class="sign-up-card">
      <div class="card-header">
        <h2>创建账号</h2>
        <p class="subtitle">加入我们的编程社区</p>
      </div>

      <form class="sign-up-form" @submit.prevent="handleRegister">
        <div v-if="errorMessage" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ errorMessage }}
        </div>

        <div class="form-group">
          <label for="username">
            <i class="fas fa-user"></i>
            用户名
          </label>
          <input
            type="text"
            id="username"
            v-model="username"
            required
            placeholder="请输入用户名"
            :class="{ 'input-focus': usernameFocus }"
            @focus="usernameFocus = true"
            @blur="usernameFocus = false"
          />
        </div>

        <div class="form-group">
          <label for="email">
            <i class="fas fa-envelope"></i>
            邮箱
          </label>
          <input
            type="email"
            id="email"
            v-model="email"
            required
            placeholder="请输入邮箱"
            :class="{ 'input-focus': emailFocus }"
            @focus="emailFocus = true"
            @blur="emailFocus = false"
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
              required
              placeholder="请输入密码"
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

        <div class="form-group">
          <label for="confirmPassword">
            <i class="fas fa-lock"></i>
            确认密码
          </label>
          <div class="password-input">
            <input
              :type="showConfirmPassword ? 'text' : 'password'"
              id="confirmPassword"
              v-model="confirmPassword"
              required
              placeholder="请再次输入密码"
              :class="{ 'input-focus': confirmPasswordFocus }"
              @focus="confirmPasswordFocus = true"
              @blur="confirmPasswordFocus = false"
            />
            <i
              class="fas"
              :class="showConfirmPassword ? 'fa-eye-slash' : 'fa-eye'"
              @click="showConfirmPassword = !showConfirmPassword"
            ></i>
          </div>
        </div>

        <HumanVerify @verify-success="handleVerifySuccess" :threshold="90" />

        <button type="submit" class="submit-btn" :disabled="loading" :class="{ loading: loading }">
          <span class="btn-content">
            <i class="fas fa-user-plus" v-if="!loading"></i>
            <i class="fas fa-circle-notch fa-spin" v-else></i>
            {{ loading ? '注册中...' : '注册' }}
          </span>
        </button>

        <div class="sign-in-link">
          已有账号？
          <router-link to="/sign-in" class="highlight-link">立即登录</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import HumanVerify from '@/components/HumanVerify.vue'
import { ref } from 'vue'
import { useUserStore } from '@/stores/modules/user'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const router = useRouter()

const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const errorMessage = ref('')
const loading = ref(false)
const isVerified = ref(false)
const usernameFocus = ref(false)
const emailFocus = ref(false)
const passwordFocus = ref(false)
const confirmPasswordFocus = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)

interface ApiError {
  response?: {
    data?: {
      message?: string
    }
  }
}

const handleVerifySuccess = () => {
  isVerified.value = true
}

const handleRegister = async () => {
  // console.log('Store methods:', Object.keys(userStore))
  // console.log('=== SignUpView.handleRegister Start ===')
  console.log('Form data:', {
    username: username.value,
    email: email.value,
    password: '***',
    isVerified: isVerified.value,
  })

  if (!isVerified.value) {
    // console.log('Verification not completed')
    errorMessage.value = '请完成人机验证'
    return
  }

  if (!username.value || !email.value || !password.value || !confirmPassword.value) {
    // console.log('Missing required fields')
    errorMessage.value = '请填写所有必填项'
    return
  }

  if (password.value !== confirmPassword.value) {
    // console.log('Password mismatch')
    errorMessage.value = '两次输入的密码不一致'
    return
  }

  if (password.value.length < 6) {
    // console.log('Password too short')
    errorMessage.value = '密码长度至少为6位'
    return
  }

  try {
    loading.value = true
    // console.log('Calling userStore.register...')
    await userStore.register(username.value, email.value, password.value)
    // console.log('Registration successful')
    router.push('/')
  } catch (error) {
    // console.error('=== Register Handler Error ===')
    const apiError = error as ApiError
    // console.error('Error:', apiError)
    // console.error('Response:', apiError.response)
    errorMessage.value = apiError.response?.data?.message || '注册失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.sign-up-container {
  min-height: calc(100vh - 64px);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 80px 2rem 2rem;
  background: linear-gradient(135deg, var(--bg-dark) 0%, var(--nav-bg-dark) 100%);
}

.sign-up-card {
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

.sign-up-card:hover {
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
