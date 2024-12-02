import { defineStore } from 'pinia'

interface UserState {
  user: {
    id: number
    username: string
    email: string
    avatar: string
    role: string
  } | null
  token: string | null
  ws: WebSocket | null
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    user: null,
    token: localStorage.getItem('token') || '',
    ws: null
  }),
  getters: {
    isLoggedIn: (state) => state.user !== null,
    userInfo: (state) => state.user,
  },
  actions: {
    setUser(user: UserState['user']) {
      this.user = user
      if (user) {
        this.connectWebSocket()
      }
    },
    setToken(token: string) {
      this.token = token
      localStorage.setItem('token', token)
    },
    logout() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
      this.user = null
      this.token = null
      localStorage.removeItem('token')
    },
    clearToken() {
      // console.log('[UserStore] 清除token')
      this.token = ''
      localStorage.removeItem('token')
    },
    connectWebSocket() {
      const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost/api'}/ws?token=${this.token}`
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        // console.log('[WebSocket] 连接已建立')
      }

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          // console.log('[WebSocket] 收到消息:', message)

          switch (message.type) {
            case 'connected':
              // console.log('[WebSocket] 连接成功:', message.data)
              break
            case 'judge_result':
              window.dispatchEvent(new CustomEvent('judge_result', { detail: message.data }))
              break
            case 'judge_status':
              window.dispatchEvent(new CustomEvent('judge_status', { detail: message.data }))
              break
          }
        } catch (error) {
          // console.error('[WebSocket] 消息解析错误:', error)
        }
      }

      this.ws.onclose = () => {
        // console.log('[WebSocket] 连接已关闭')
        this.ws = null
        setTimeout(() => {
          if (this.user) {
            this.connectWebSocket()
          }
        }, 5000)
      }

      this.ws.onerror = (error) => {
        // console.error('[WebSocket] 连接错误:', error)
      }

      const heartbeat = setInterval(() => {
        if (this.ws?.readyState === WebSocket.OPEN) {
          this.ws.send('ping')
        } else if (!this.user) {
          clearInterval(heartbeat)
        }
      }, 30000)
    }
  },
  persist: true,
})
