import { defineStore } from 'pinia'

interface AppState {
  loading: boolean
  notification: {
    show: boolean
    type: 'success' | 'error' | 'info' | 'warning'
    message: string
  }
}

export const useAppStore = defineStore('app', {
  state: (): AppState => ({
    loading: false,
    notification: {
      show: false,
      type: 'info',
      message: '',
    },
  }),

  actions: {
    setLoading(status: boolean) {
      this.loading = status
    },

    showNotification(type: 'success' | 'error' | 'info' | 'warning', message: string) {
      this.notification = {
        show: true,
        type,
        message,
      }
      // 3秒后自动关闭
      setTimeout(() => {
        this.notification.show = false
      }, 3000)
    },
  },
})
