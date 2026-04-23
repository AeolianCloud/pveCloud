import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { login as loginRequest, register as registerRequest } from '../api/auth'
import { readStoredToken, writeStoredToken } from '../lib/http'

const authStorageKey = 'pvecloud-web-auth-meta'

interface AuthMeta {
  subjectID: number | null
  subjectType: string
}

function readStoredMeta(): AuthMeta {
  if (typeof window === 'undefined') {
    return { subjectID: null, subjectType: '' }
  }

  const raw = window.localStorage.getItem(authStorageKey)
  if (!raw) {
    return { subjectID: null, subjectType: '' }
  }

  try {
    return JSON.parse(raw) as AuthMeta
  } catch {
    return { subjectID: null, subjectType: '' }
  }
}

function writeStoredMeta(meta: AuthMeta) {
  if (typeof window === 'undefined') {
    return
  }

  if (meta.subjectID === null) {
    window.localStorage.removeItem(authStorageKey)
    return
  }

  window.localStorage.setItem(authStorageKey, JSON.stringify(meta))
}

export const useAuthStore = defineStore('auth', () => {
  const initialMeta = readStoredMeta()
  const token = ref(readStoredToken())
  const subjectID = ref<number | null>(initialMeta.subjectID)
  const subjectType = ref(initialMeta.subjectType)
  const isAuthenticated = computed(() => Boolean(token.value))

  async function login(phone: string, password: string) {
    const payload = await loginRequest(phone, password)
    token.value = payload.token
    subjectID.value = payload.subject_id
    subjectType.value = payload.subject_type
    writeStoredToken(payload.token)
    writeStoredMeta({ subjectID: payload.subject_id, subjectType: payload.subject_type })
  }

  async function register(phone: string, email: string, password: string) {
    const payload = await registerRequest(phone, email, password)
    token.value = payload.token
    subjectID.value = payload.user_id
    subjectType.value = payload.subject_type
    writeStoredToken(payload.token)
    writeStoredMeta({ subjectID: payload.user_id, subjectType: payload.subject_type })
  }

  function logout() {
    token.value = ''
    subjectID.value = null
    subjectType.value = ''
    writeStoredToken('')
    writeStoredMeta({ subjectID: null, subjectType: '' })
  }

  return {
    token,
    subjectID,
    subjectType,
    isAuthenticated,
    login,
    register,
    logout,
  }
})
