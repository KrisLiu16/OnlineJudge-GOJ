import { http } from '@/utils/http'

export interface WebsiteSettings {
  title: string
  subtitle: string
  about: string
  email: string
  github: string
  icp: string
  icpLink: string
  feature1: string
  feature2: string
  feature3: string
}

// 获取公开的网站设置
export const getPublicWebsiteSettings = () => {
  return http.get<WebsiteSettings>('/api/website/settings')
}

// 获取管理员的网站设置
export const getAdminWebsiteSettings = () => {
  return http.get<WebsiteSettings>('/api/admin/website/settings')
}

// 更新网站设置
export const updateWebsiteSettings = (settings: WebsiteSettings) => {
  return http.post('/api/admin/website/settings', settings)
}
