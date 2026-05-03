<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useWebAuthStore } from '../../store/modules/auth'

const route = useRoute()
const router = useRouter()
const authStore = useWebAuthStore()

const account = ref('')
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

const canSubmit = computed(() => account.value.trim() !== '' && password.value.length >= 6 && !loading.value)

async function handleLogin() {
  if (!canSubmit.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    await authStore.login({ account: account.value.trim(), password: password.value })
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/user'
    await router.replace(redirect)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="page auth-page">
    <div class="auth-left">
      <div class="hero-badge" style="color:var(--c-primary); background:var(--c-primary-soft);">
        <span style="width:6px;height:6px;border-radius:50%;background:var(--c-primary);display:inline-block;"></span>
        用户登录
      </div>
      <h1>进入云资源控制台</h1>
      <p>
        登录后进入统一控制台入口。当前阶段先开放登录态、会话恢复和退出能力，实例、订单和工单仍按占位展示。
      </p>
      <div style="display:grid; gap:10px; margin-top:8px;">
        <div v-for="item in ['登录和控制台使用同一入口', '刷新页面后自动恢复登录态', '退出后清理本地登录信息']" :key="item" style="display:flex; align-items:center; gap:8px; font-size:.92rem; color:var(--c-text-2);">
          <span style="width:5px;height:5px;border-radius:50%;background:var(--c-primary);flex-shrink:0;"></span>
          {{ item }}
        </div>
      </div>
    </div>
    <div class="auth-right">
      <form class="auth-form" @submit.prevent="handleLogin">
        <h2>登录</h2>
        <label>
          <span>邮箱或用户名</span>
          <input v-model="account" type="text" placeholder="请输入邮箱或用户名" autocomplete="username" />
        </label>
        <label>
          <span>密码</span>
          <input v-model="password" type="password" placeholder="请输入密码" autocomplete="current-password" />
        </label>
        <p v-if="errorMessage" class="hint error-text">{{ errorMessage }}</p>
        <button class="btn btn-primary" type="submit" :disabled="!canSubmit" style="width:100%">
          {{ loading ? '登录中...' : '登录' }}
        </button>
        <p class="hint">
          暂未开放注册，请联系管理员创建账号。
        </p>
      </form>
    </div>
  </section>
</template>
