<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'

import { getPasswordResetRequestCaptcha, requestPasswordReset } from '../../api/auth'
import { useAuthCaptcha } from '../../composables/use-auth-captcha'
import { useWebAppStore } from '../../store/modules/app'

const appStore = useWebAppStore()
const { passwordResetRequestCaptchaEnabled, siteConfigLoaded } = storeToRefs(appStore)

const email = ref('')
const loading = ref(false)
const errorMessage = ref('')
const sent = ref(false)

const {
  captchaCode,
  captchaError,
  captchaId,
  captchaImage,
  captchaLoading,
  captchaReady,
  refreshCaptcha,
} = useAuthCaptcha(passwordResetRequestCaptchaEnabled, getPasswordResetRequestCaptcha)

const canSubmit = computed(() => {
  return (
    siteConfigLoaded.value &&
    email.value.trim() !== '' &&
    (!passwordResetRequestCaptchaEnabled.value || captchaCode.value.trim().length >= 4) &&
    captchaReady.value &&
    !loading.value
  )
})

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
    await requestPasswordReset({
      email: email.value.trim(),
      captcha_id: passwordResetRequestCaptchaEnabled.value ? captchaId.value : undefined,
      captcha_code: passwordResetRequestCaptchaEnabled.value ? captchaCode.value.trim() : undefined,
    })
    sent.value = true
    if (passwordResetRequestCaptchaEnabled.value) {
      void refreshCaptcha()
    }
  } catch (error) {
    errorMessage.value = errorText(error)
    if (passwordResetRequestCaptchaEnabled.value) {
      void refreshCaptcha()
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void appStore.loadSiteConfig()
})
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
        <div v-if="passwordResetRequestCaptchaEnabled" class="captcha-field">
          <label>
            <span>验证码</span>
            <input v-model="captchaCode" type="text" maxlength="8" placeholder="请输入验证码" autocomplete="off" />
          </label>
          <div class="captcha-row">
            <img v-if="captchaImage" class="captcha-image" :src="captchaImage" alt="找回密码验证码" />
            <div v-else class="captcha-image captcha-image--placeholder">
              {{ captchaLoading ? '加载中...' : '暂无验证码' }}
            </div>
            <button class="captcha-refresh" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
              {{ captchaLoading ? '刷新中...' : '换一张' }}
            </button>
          </div>
        </div>
        <p v-if="sent" class="hint success-text">如果邮箱对应有效账号，重置链接会发送到该邮箱。</p>
        <p v-if="captchaError" class="hint error-text">{{ captchaError }}</p>
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
