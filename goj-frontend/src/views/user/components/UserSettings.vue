<template>
  <div class="settings-container">
    <div class="settings-section">
      <h3>个人信息</h3>
      <div class="form-group">
        <label>邮箱</label>
        <input v-model="editForm.email" type="email" disabled class="input-disabled" />
      </div>
      <div class="form-group">
        <label>个人简介</label>
        <textarea v-model="editForm.bio" placeholder="写点什么介绍自己..."></textarea>
      </div>
      <div class="form-actions">
        <button @click="handleUpdateProfile" :disabled="updating" class="update-btn">
          {{ updating ? '更新中...' : '保存修改' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/stores/modules/user'
import { ElMessage } from 'element-plus'

interface ApiError {
  response?: {
    data?: {
      message?: string
    }
  }
}

const userStore = useUserStore()

const editForm = ref({
  email: userStore.userInfo?.email || '',
  bio: userStore.userInfo?.bio || '',
})

const updating = ref(false)

const handleUpdateProfile = async () => {
  try {
    updating.value = true
    await userStore.updateProfile({
      bio: editForm.value.bio,
    })
    ElMessage.success('个人信息更新成功')
  } catch (error: unknown) {
    const err = error as ApiError
    ElMessage.error(err.response?.data?.message || '更新失败')
  } finally {
    updating.value = false
  }
}
</script>

<style scoped>
.settings-container {
  max-width: 800px;
  margin: 5rem auto 0;
  padding: 0 1rem;
}

.settings-section {
  background: var(--nav-bg-dark);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
}

.settings-section h3 {
  margin-top: 0;
  margin-bottom: 1.5rem;
  color: var(--text-light);
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--text-light);
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--nav-bg-light);
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-light);
}

.input-disabled {
  background: rgba(255, 255, 255, 0.05) !important;
  cursor: not-allowed;
  opacity: 0.7;
}

.form-group textarea {
  min-height: 100px;
  resize: vertical;
}

.form-actions {
  text-align: right;
}

.update-btn {
  padding: 0.5rem 1.5rem;
  background: var(--primary-color);
  border: none;
  border-radius: 4px;
  color: var(--bg-dark);
  cursor: pointer;
  font-weight: bold;
}

.update-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
</style>
