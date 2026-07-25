import { defineStore } from 'pinia'
import { api, ApiError } from '@/lib/api'

export interface User {
  id: string
  name: string
  email: string
}

interface AuthState {
  user: User | null
  initialized: boolean
}

interface UserResponse {
  user: User
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
    initialized: false,
  }),

  getters: {
    isAuthenticated: (state): boolean => state.user !== null,
  },

  actions: {
    // Hydrate the current user from the session cookie. Called once on startup;
    // a 401 simply means "not logged in" and is not an error here.
    async fetchMe(): Promise<void> {
      try {
        const { user } = await api.get<UserResponse>('/auth/me')
        this.user = user
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) {
          this.user = null
        } else {
          throw err
        }
      } finally {
        this.initialized = true
      }
    },

    async login(email: string, password: string): Promise<void> {
      const { user } = await api.post<UserResponse>('/auth/login', { email, password })
      this.user = user
    },

    async register(payload: {
      name: string
      email: string
      password: string
      password_confirmation: string
    }): Promise<void> {
      const { user } = await api.post<UserResponse>('/auth/register', payload)
      this.user = user
    },

    async logout(): Promise<void> {
      await api.post('/auth/logout')
      this.user = null
    },

    async updateProfile(name: string): Promise<void> {
      const { user } = await api.post<UserResponse>('/auth/profile', { name })
      this.user = user
    },

    async changePassword(payload: {
      current_password: string
      password: string
      password_confirmation: string
    }): Promise<string> {
      const { message } = await api.post<{ message: string }>('/auth/password/change', payload)
      return message
    },

    async changeEmail(payload: { new_email: string; current_password: string }): Promise<string> {
      const { message } = await api.post<{ message: string }>('/auth/email/change', payload)
      return message
    },

    async confirmEmailChange(token: string): Promise<string> {
      const { message, user } = await api.post<{ message: string; user?: User }>(
        '/auth/email/confirm',
        { token },
      )
      // If this browser is signed in as the same user, reflect the new email.
      if (user) this.user = user
      return message
    },

    async forgotPassword(email: string): Promise<string> {
      const { message } = await api.post<{ message: string }>('/auth/password/forgot', { email })
      return message
    },

    async resetPassword(payload: {
      token: string
      password: string
      password_confirmation: string
    }): Promise<string> {
      const { message } = await api.post<{ message: string }>('/auth/password/reset', payload)
      return message
    },
  },
})
