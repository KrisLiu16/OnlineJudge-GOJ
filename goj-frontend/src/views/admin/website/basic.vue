<template>
  <div class="website-settings">
    <el-card class="settings-card">
      <template #header>
        <div class="card-header">
          <span>网站基本设置</span>
        </div>
      </template>

      <el-form :model="form" label-width="120px" v-loading="loading">
        <el-form-item label="网站标题">
          <el-input v-model="form.title" placeholder="例如：GO! Judge" />
        </el-form-item>

        <el-form-item label="副标题">
          <el-input v-model="form.subtitle" placeholder="例如：快速、智能的在线评测系统" />
        </el-form-item>

        <el-form-item label="ICP备案号">
          <el-input v-model="form.icp" placeholder="例如：京ICP备XXXXXXXX号" />
        </el-form-item>

        <el-form-item label="ICP备案链接">
          <el-input v-model="form.icpLink" placeholder="例如：https://beian.miit.gov.cn/" />
        </el-form-item>

        <el-form-item label="关于我们">
          <el-input
            v-model="form.about"
            type="textarea"
            :rows="4"
            placeholder="网站简介"
          />
        </el-form-item>

        <el-form-item label="联系邮箱">
          <el-input v-model="form.email" placeholder="support@example.com" />
        </el-form-item>

        <el-form-item label="GitHub链接">
          <el-input v-model="form.github" placeholder="https://github.com/yourusername" />
        </el-form-item>

        <el-divider>首页特性块设置</el-divider>

        <el-form-item label="特性块1">
          <el-input
            v-model="form.feature1"
            type="textarea"
            :rows="6"
            :placeholder="defaultFeatures.feature1"
          />
        </el-form-item>

        <el-form-item label="特性块2">
          <el-input
            v-model="form.feature2"
            type="textarea"
            :rows="6"
            :placeholder="defaultFeatures.feature2"
          />
        </el-form-item>

        <el-form-item label="特性块3">
          <el-input
            v-model="form.feature3"
            type="textarea"
            :rows="6"
            :placeholder="defaultFeatures.feature3"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="saveSettings" :loading="saving">
            保存设置
          </el-button>
          <el-button @click="resetFeatures">重置特性块为默认值</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/modules/user'

const userStore = useUserStore()
const loading = ref(false)
const saving = ref(false)

const defaultFeatures = {
  feature1: '<div class="feature-icon"><span class="icon-wrapper">📚</span></div><h3>丰富的题库</h3><p>包含各种难度的编程题目，从入门到进阶</p>',
  feature2: '<div class="feature-icon"><span class="icon-wrapper">🚀</span></div><h3>实时评测</h3><p>快速的代码执行和结果反馈</p>',
  feature3: '<div class="feature-icon"><span class="icon-wrapper">👥</span></div><h3>社区讨论</h3><p>与其他同学交流学习心得</p>'
}

const form = ref({
  title: 'GO! Judge',
  subtitle: '快速、智能的在线评测系统',
  icp: '',
  icpLink: '',
  about: 'GOJ是一个高性能在线评测的平台，致力于提供快速、稳定的评测服务。',
  email: 'support@example.com',
  github: 'https://github.com/yourusername',
  feature1: defaultFeatures.feature1,
  feature2: defaultFeatures.feature2,
  feature3: defaultFeatures.feature3
})

const resetFeatures = () => {
  form.value.feature1 = defaultFeatures.feature1
  form.value.feature2 = defaultFeatures.feature2
  form.value.feature3 = defaultFeatures.feature3
  ElMessage.success('已重置特性块为默认值')
}

const loadSettings = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/admin/settings', {
      headers: {
        Authorization: `Bearer ${userStore.token}`
      }
    })
    if (!response.ok) throw new Error('加载设置失败')
    const data = await response.json()
    if (data.code === 200) {
      // 合并默认值和服务器返回的值
      form.value = {
        ...form.value,
        ...data.data,
        feature1: data.data.feature1 || defaultFeatures.feature1,
        feature2: data.data.feature2 || defaultFeatures.feature2,
        feature3: data.data.feature3 || defaultFeatures.feature3
      }
    }
  } catch (error) {
    ElMessage.error('加载设置失败')
  } finally {
    loading.value = false
  }
}

const saveSettings = async () => {
  saving.value = true
  try {
    const response = await fetch('/api/admin/settings', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userStore.token}`
      },
      body: JSON.stringify(form.value)
    })
    if (!response.ok) throw new Error('保存设置失败')
    const data = await response.json()
    if (data.code === 200) {
      ElMessage.success('设置保存成功')
      localStorage.setItem('websiteSettings', JSON.stringify(form.value))
    } else {
      throw new Error(data.message || '保存设置失败')
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存设置失败')
  } finally {
    saving.value = false
  }
}

onMounted(loadSettings)
</script>

<style scoped>
.website-settings {
  padding: 20px;
}

.settings-card {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
