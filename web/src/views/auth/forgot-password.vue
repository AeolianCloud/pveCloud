<script setup lang="ts">
import { computed, ref } from 'vue'

import { requestPasswordReset } from '../../api/auth'

const email = ref('')
const loading = ref(false)
const errorMessage = ref('')
const sent = ref(false)

const canSubmit = computed(() => email.value.trim() !== '' && !loading.value)

function errorText(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (typeof error === 'object' && error !== null && 'request' in error) {
    return '网络连接失败，请检查后重试'
  }
  return '密码找回服务暂不可用，请稍后再试'
}

async function handleSubmit() {
  if (!canSubmit.value) return
  loading.value = true
  errorMessage.value = ''
  sent.value = false
  try {
    await requestPasswordReset({ email: email.value.trim() })
    sent.value = true
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
      <div class="hero-badge" style="color:var(--c-orange); background:var(--c-orange-soft);">
        <span style="width:6px;height:6px;border-radius:50%;background:var(--c-orange);display:inline-block;"></span>
        密码找回
      </div>
      <h1>通过邮箱重置密码</h1>
      <p>如果邮箱对应有效账号，系统会发送一次性重置链接。为了账号安全，页面不会暴露邮箱是否已经注册。</p>
    </div>
    <div class="auth-right">
      <form class="auth-form" @submit.prevent="handleSubmit">
        <h2>找回密码</h2>
        <label>
          <span>注册邮箱</span>
          <input v-model="email" type="email" placeholder="请输入邮箱" autocomplete="email" />
        </label>
        <p v-if="sent" class="hint success-text">如果邮箱对应有效账号，重置链接会发送到该邮箱。</p>
        <p v-if="errorMessage" class="hint error-text">{{ errorMessage }}</p>
        <button class="btn btn-primary" type="submit" :disabled="!canSubmit">
          {{ loading ? '发送中...' : '发送重置链接' }}
        </button>
        <p class="hint">
          想起密码了？<RouterLink class="link" to="/login">返回登录</RouterLink>
        </p>
      </form>
    </div>
  </section>
</template>
