import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AdminUser } from '@/types'
import { getProfile, logout as logoutApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const refreshToken = ref<string>(localStorage.getItem('refresh_token') || '')
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

  // 设置 access + refresh token（登录/刷新时使用）
  function setTokens(accessToken: string, newRefreshToken: string) {
    token.value = accessToken
    refreshToken.value = newRefreshToken
    localStorage.setItem('token', accessToken)
    localStorage.setItem('refresh_token', newRefreshToken)
  }

  // 拉取当前用户信息
  async function fetchUser() {
    const res = await getProfile()
    user.value = res.data.data
  }

  // logoutLocal 仅清理本地登录态（不请求后端）。
  // 用途：Refresh 失败 / Token 彻底失效时快速回到登录页。
  function logoutLocal() {
    token.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')
  }

  // logout 退出登录：优先请求后端撤销会话，再清理本地登录态。
  // 注意：后端接口幂等；请求失败也会清本地，避免用户卡住。
  async function logout() {
    try {
      if (token.value) {
        await logoutApi()
      }
    } catch {
      // 忽略错误：后端可能已过期/已撤销
    } finally {
      logoutLocal()
    }
  }

  return {
    token,
    refreshToken,
    user,
    isLoggedIn,
    isSuperAdmin,
    permissionSet,
    hasPermission,
    setToken,
    setTokens,
    fetchUser,
    logoutLocal,
    logout,
  }
})
