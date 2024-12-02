import axios from 'axios'
import type { ApiResponse } from '@/types/user'
import { useUserStore } from '@/stores/modules/user'

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    // 确保 config.headers 存在
    config.headers = config.headers || {}

    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }


    return config
  },
  (error) => {
    console.error('Request Error:', error)
    return Promise.reject(error)
  },
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    return Promise.reject(error)
  },
)

export default request
