import { defineStore, PiniaPluginContext } from 'pinia'
import { userApi } from '@/api/user'
import type { User, LoginResponse, PasswordUpdateResponse } from '@/types/user'

interface UserState {
  token: string | null
  userInfo: User | null
  isAuthenticated: boolean
  userAvatar: string
}

interface UserStoreActions {
  setUserState(userData: LoginResponse): void
  login(account: string, password: string): Promise<LoginResponse | undefined>
  register(username: string, email: string, password: string): Promise<LoginResponse | undefined>
  clearUserState(): void
  logout(): void
  updateProfile(data: Partial<User>): Promise<User | undefined>
  fetchUserProfile(): Promise<User | undefined>
  updateAvatar(formData: FormData): Promise<{ avatar: string } | undefined>
  updatePassword(
    oldPassword: string,
    newPassword: string,
  ): Promise<PasswordUpdateResponse | undefined>
}

export const useUserStore = defineStore<
  'user',
  UserState,
  {
    isLoggedIn: (state: UserState) => boolean
    user: (state: UserState) => User | null
    role: (state: UserState) => string
    userID: (state: UserState) => number | undefined
  },
  UserStoreActions
>('user', {
  state: (): UserState => ({
    token: null,
    userInfo: null,
    isAuthenticated: false,
    userAvatar: '/images/avatars/default-avatar.png',
  }),

  getters: {
    isLoggedIn: (state) => state.isAuthenticated,
    user: (state) => state.userInfo,
    role: (state) => state.userInfo?.role || '',
    userID: (state) => state.userInfo?.id,
  },

  actions: {
    setUserState(userData: LoginResponse) {
      this.token = userData.token
      this.userInfo = userData.user
      this.isAuthenticated = true
      this.userAvatar = userData.user.avatar || '/images/avatars/default-avatar.png'
    },

    async login(account: string, password: string) {
      try {
        const res = await userApi.login({ account, password })
        if (res.data) {
          this.setUserState(res.data)
          return res.data
        }
      } catch (error) {
        this.clearUserState()
        throw error
      }
    },

    async register(username: string, email: string, password: string) {
      try {
        await userApi.register({
          username,
          email,
          password,
          avatar: '/images/avatars/default-avatar.png',
        })
        return await this.login(email, password)
      } catch (error) {
        throw error
      }
    },

    clearUserState() {
      this.token = null
      this.userInfo = null
      this.isAuthenticated = false
      this.userAvatar = '/images/avatars/default-avatar.png'
    },

    logout() {
      this.clearUserState()
    },

    async updateProfile(data: Partial<User>) {
      try {
        const res = await userApi.updateProfile(data)
        if (res.data) {
          if (this.userInfo) {
            this.userInfo = { ...this.userInfo, ...res.data }
            this.userAvatar = res.data.avatar || '/images/avatars/default-avatar.png'
          }
          return res.data
        }
      } catch (error) {
        throw error
      }
    },

    async fetchUserProfile() {
      try {
        const res = await userApi.getProfile()
        if (res.data) {
          this.userInfo = res.data
          this.userAvatar = res.data.avatar || '/images/avatars/default-avatar.png'
          this.isAuthenticated = true
          return res.data
        }
      } catch (error) {
        this.clearUserState()
        throw error
      }
    },

    async updateAvatar(formData: FormData) {
      try {
        const res = await userApi.updateAvatar(formData)
        if (res.data) {
          this.userAvatar = res.data.avatar
          if (this.userInfo) {
            this.userInfo.avatar = res.data.avatar
          }
          return res.data
        }
      } catch (error) {
        throw error
      }
    },

    async updatePassword(oldPassword: string, newPassword: string) {
      try {
        const res = await userApi.updatePassword({ oldPassword, newPassword })
        return res.data
      } catch (error) {
        throw error
      }
    },
  },

  persist: {
    key: 'user-store',
    paths: ['token', 'userInfo', 'isAuthenticated', 'userAvatar'],
  } as PiniaPluginContext['options']['persist'],
})
