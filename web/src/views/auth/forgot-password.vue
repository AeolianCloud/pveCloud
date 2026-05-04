<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'

import { getPasswordResetRequestCaptcha, requestPasswordReset } from '../../api/auth'
import { useAuthCaptcha } from '../../composables/use-auth-captcha'
import { useWebAppStore } from '../../store/modules/app'

const appStore = useWebAppStore()
const { passwordResetRequestCaptchaEnabled, siteConfigError, siteConfigLoaded, siteConfigLoading } = storeToRefs(appStore)

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

const submitHint = computed(() => {
  if (siteConfigLoading.value && !siteConfigLoaded.value) {
    return '正在加载找回密码配置，请稍候...'
  }
  if (email.value.trim() === '') {
    return '请输入注册邮箱'
  }
  if (passwordResetRequestCaptchaEnabled.value && !captchaReady.value) {
    return captchaError.value || '验证码加载中，请稍候...'
  }
  if (passwordResetRequestCaptchaEnabled.value && captchaCode.value.trim().length < 4) {
    return '请输入验证码后再发送重置链接'
  }
  if (siteConfigError.value) {
    return siteConfigError.value
  }
  return ''
})

const statusTone = computed(() => {
  if (captchaError.value || errorMessage.value) return 'danger'
  if (sent.value) return 'success'
  if (!canSubmit.value) return 'muted'
  return 'success'
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
  <section class="page auth-page auth-page--forgot">
    <div class="auth-stage auth-stage--forgot">
      <aside class="auth-panel auth-panel--story">
        <div class="auth-kicker auth-kicker--warning">pveCloud</div>
        <h1 class="auth-display">通过邮箱重置密码</h1>
        <p class="auth-copy">
          如果邮箱对应有效账号，系统会发送一次性重置链接。为了账号安全，页面不会暴露邮箱是否已经注册。
        </p>
      </aside>

      <div class="auth-panel auth-panel--form-shell">
        <div class="auth-form-card">
          <div class="auth-form-card__header">
            <div>
              <p class="auth-eyebrow">Password reset</p>
              <h2>找回密码</h2>
            </div>
            <RouterLink class="auth-mini-link" to="/login">返回登录</RouterLink>
          </div>

          <form class="auth-form auth-form--stacked" @submit.prevent="handleSubmit">
            <label class="auth-field">
              <span>注册邮箱</span>
              <input v-model="email" type="email" placeholder="输入注册邮箱" autocomplete="email" />
            </label>

            <div v-if="passwordResetRequestCaptchaEnabled" class="captcha-field auth-fieldset">
              <div class="auth-fieldset__legend">
                <span>安全校验</span>
                <small>找回密码验证码已开启</small>
              </div>
              <label class="auth-field">
                <span>验证码</span>
                <input v-model="captchaCode" type="text" maxlength="8" placeholder="输入图中字符" autocomplete="off" />
              </label>
              <div class="captcha-row captcha-row--panel">
                <img v-if="captchaImage" class="captcha-image" :src="captchaImage" alt="找回密码验证码" />
                <div v-else class="captcha-image captcha-image--placeholder">
                  {{ captchaLoading ? '加载中...' : '暂无验证码' }}
                </div>
                <button class="captcha-refresh" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
                  {{ captchaLoading ? '刷新中...' : '刷新验证码' }}
                </button>
              </div>
            </div>

            <div class="auth-status" :data-tone="statusTone">
              <p v-if="sent" class="hint success-text">如果邮箱对应有效账号，重置链接会发送到该邮箱。</p>
              <p v-else-if="captchaError" class="hint error-text">{{ captchaError }}</p>
              <p v-else-if="errorMessage" class="hint error-text">{{ errorMessage }}</p>
              <p v-else class="hint">{{ submitHint || '确认邮箱后发送重置链接' }}</p>
            </div>

            <button class="btn btn-primary auth-submit" type="submit" :disabled="!canSubmit">
              {{ loading ? '发送中...' : '发送重置链接' }}
            </button>

            <div class="auth-secondary-row">
              <span class="hint">想起密码了？</span>
              <RouterLink class="link" to="/login">返回登录</RouterLink>
            </div>
          </form>
        </div>
      </div>
    </div>
  </section>
</template>
