import { defineStore } from 'pinia'
import { ref } from 'vue'

import { request } from '../lib/http'

interface LoginPayload {
  token: string
  subject_id: number
  subject_type: string
}

interface RegisterPayload {
  token: string
  user_id: number
  user_no: string
  subject_type: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const subjectID = ref<number | null>(null)
  const subjectType = ref('')

  async function login(phone: string, password: string) {
    const payload = await request<LoginPayload>('/auth/login', {
      method: 'POST',
      bodyJson: { phone, password },
    })
    token.value = payload.token
    subjectID.value = payload.subject_id
    subjectType.value = payload.subject_type
  }

  async function register(phone: string, email: string, password: string) {
    const payload = await request<RegisterPayload>('/auth/register', {
      method: 'POST',
      bodyJson: { phone, email, password },
    })
    token.value = payload.token
    subjectID.value = payload.user_id
    subjectType.value = payload.subject_type
  }

  function logout() {
    token.value = ''
    subjectID.value = null
    subjectType.value = ''
  }

  return {
    token,
    subjectID,
    subjectType,
    login,
    register,
    logout,
  }
})
