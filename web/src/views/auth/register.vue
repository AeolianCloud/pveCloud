<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useWebAuthStore } from '../../store/modules/auth'

const router = useRouter()
const authStore = useWebAuthStore()

const username = ref('')
const email = ref('')
const displayName = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMessage = ref('')

const canSubmit = computed(() => {
  return (
    username.value.trim().length >= 3 &&
    email.value.trim() !== '' &&
    password.value.length >= 6 &&
    password.value === confirmPassword.value &&
    !loading.value
  )
})

function errorText(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { status?: number; data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (typeof error === 'object' && error !== null && 'request' in error) {
    return '网络连接失败，请检查后重试'
  }
  return '注册失败，请稍后再试'
}

async function handleRegister() {
  if (!canSubmit.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    await authStore.register({
      username: username.value.trim(),
      email: email.value.trim(),
      password: password.value,
      display_name: displayName.value.trim() || null,
    })
    await router.replace('/user')
  } catch (error) {
    errorMessage.value = errorText(error)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="page auth-page">
    <div class="auth-left">
      <div class="hero-badge" style="color:var(--c-green); background:var(--c-green-soft);">
        <span style="width:6px;height:6px;border-radius:50%;background:var(--c-green);display:inline-block;"></span>
        用户注册
      </div>
      <h1>创建云资源账号</h1>
      <p>注册后可以进入控制台入口，当前阶段仅开放账号自助能力，不创建订单、实例或钱包数据。</p>
    </div>
    <div class="auth-right">
      <form class="auth-form" @submit.prevent="handleRegister">
        <h2>注册</h2>
        <label>
          <span>用户名</span>
          <input v-model="username" type="text" placeholder="至少 3 个字符" autocomplete="username" />
        </label>
        <label>
          <span>邮箱</span>
          <input v-model="email" type="email" placeholder="用于登录和密码找回" autocomplete="email" />
        </label>
        <label>
          <span>显示名称</span>
          <input v-model="displayName" type="text" placeholder="可选" autocomplete="name" />
        </label>
        <label>
          <span>密码</span>
          <input v-model="password" type="password" placeholder="至少 6 个字符" autocomplete="new-password" />
        </label>
        <label>
          <span>确认密码</span>
          <input v-model="confirmPassword" type="password" placeholder="再次输入密码" autocomplete="new-password" />
        </label>
        <p v-if="password && confirmPassword && password !== confirmPassword" class="hint error-text">两次输入的密码不一致</p>
        <p v-if="errorMessage" class="hint error-text">{{ errorMessage }}</p>
        <button class="btn btn-primary" type="submit" :disabled="!canSubmit">
          {{ loading ? '注册中...' : '注册并进入控制台' }}
        </button>
        <p class="hint">
          已有账号？<RouterLink class="link" to="/login">返回登录</RouterLink>
        </p>
      </form>
    </div>
  </section>
</template>
