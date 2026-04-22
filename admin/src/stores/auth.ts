import { defineStore } from 'pinia'
import { ref } from 'vue'

import { request } from '../lib/http'

interface LoginPayload {
  token: string
  subject_id: number
  subject_type: string
}

export const useAuthStore = defineStore('admin-auth', () => {
  const token = ref('')
  const subjectID = ref<number | null>(null)
  const subjectType = ref('')

  async function login(username: string, password: string) {
    const payload = await request<LoginPayload>('/auth/login', {
      method: 'POST',
      bodyJson: { username, password },
    })
    token.value = payload.token
    subjectID.value = payload.subject_id
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
    logout,
  }
})
