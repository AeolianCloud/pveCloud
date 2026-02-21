import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AdminUser } from '@/types'
import { getProfile } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<AdminUser | null>(null)

  // 是否已登录
  const isLoggedIn = computed(() => !!token.value)

  // 是否超级管理员（拥有 super_admin 角色名）
  const isSuperAdmin = computed(() =>
    user.value?.roles?.some(r => r.name === 'super_admin') ?? false
  )

  // 当前用户所有权限 name 的扁平集合
  const permissionSet = computed<Set<string>>(() => {
    const names: string[] = []
    user.value?.roles?.forEach(role => {
      role.permissions?.forEach(p => names.push(p.name))
    })
    return new Set(names)
  })

  // 判断是否拥有指定权限：超管直接放行
  function hasPermission(name: string): boolean {
    return isSuperAdmin.value || permissionSet.value.has(name)
  }

  // 设置 token，同步写入 localStorage
  function setToken(val: string) {
    token.value = val
    localStorage.setItem('token', val)
  }

  // 拉取当前用户信息
  async function fetchUser() {
    const res = await getProfile()
    user.value = res.data.data
  }

  // 登出：清除 token 和用户信息
  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  return { token, user, isLoggedIn, isSuperAdmin, permissionSet, hasPermission, setToken, fetchUser, logout }
})
