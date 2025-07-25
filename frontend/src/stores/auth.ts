import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'
import type { User } from '../types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => !!user.value)

  async function checkAuth() {
    try {
      const response = await api.get('/api/user/current')
      user.value = response.data.user
    } catch (error) {
      user.value = null
    }
  }

  async function createUser(name: string) {
    const response = await api.post('/api/user/create', { name })
    user.value = response.data
    return response.data
  }

  function logout() {
    user.value = null
    // In a real app, we'd also clear the session cookie
  }

  return {
    user,
    isAuthenticated,
    checkAuth,
    createUser,
    logout
  }
})