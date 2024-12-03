import axios from 'axios'
import type { AxiosInstance, AxiosResponse } from 'axios'

// 创建 axios 实例
const http: AxiosInstance = axios.create({
  baseURL: '/api', // API 的基础路径
  timeout: 10000,  // 请求超时时间
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
http.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
http.interceptors.response.use(
  (response: AxiosResponse) => {
    // 如果响应成功，直接返回数据
    return response.data
  },
  (error) => {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          // 未授权，清除 token 并跳转到登录页
          localStorage.removeItem('token')
          window.location.href = '/login'
          break
        case 403:
          // 权限不足
          console.error('权限不足')
          break
        case 404:
          // 请求的资源不存在
          console.error('请求的资源不存在')
          break
        case 500:
          // 服务器错误
          console.error('服务器错误')
          break
        default:
          console.error('发生错误:', error.response.data)
      }
    }
    return Promise.reject(error)
  }
)

export { http }
