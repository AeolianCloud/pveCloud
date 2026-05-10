import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<any>(null)

  const setToken = (newToken: string | null) => {
    token.value = newToken
    if (newToken) {
      localStorage.setItem('token', newToken)
    } else {
      localStorage.removeItem('token')
    }
  }

  const setUser = (newUser: any) => {
    user.value = newUser
  }

  const logout = () => {
    setToken(null)
    setUser(null)
  }

  return {
    token,
    user,
    setToken,
    setUser,
    logout,
  }
})