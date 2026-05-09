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

const { captchaCode, captchaError, captchaId, captchaImage, captchaLoading, captchaReady, refreshCaptcha } =
  useAuthCaptcha(passwordResetRequestCaptchaEnabled, getPasswordResetRequestCaptcha)

const canSubmit = computed(() => (
  siteConfigLoaded.value &&
  email.value.trim() !== '' &&
  (!passwordResetRequestCaptchaEnabled.value || captchaCode.value.trim().length >= 4) &&
  captchaReady.value &&
  !loading.value
))

const submitHint = computed(() => {
  if (siteConfigLoading.value && !siteConfigLoaded.value) return '正在加载找回密码配置...'
  if (email.value.trim() === '') return '请输入注册邮箱'
  if (passwordResetRequestCaptchaEnabled.value && !captchaReady.value) return captchaError.value || '验证码加载中...'
  if (passwordResetRequestCaptchaEnabled.value && captchaCode.value.trim().length < 4) return '请输入验证码'
  if (siteConfigError.value) return siteConfigError.value
  return ''
})

function errorText(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (typeof error === 'object' && error !== null && 'request' in error) return '网络连接失败，请检查后重试'
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
    if (passwordResetRequestCaptchaEnabled.value) void refreshCaptcha()
  } catch (error) {
    errorMessage.value = errorText(error)
    if (passwordResetRequestCaptchaEnabled.value) void refreshCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void appStore.loadSiteConfig()
})
</script>

<template>
  <main class="auth-page">
    <section class="auth-card surface">
      <div class="auth-heading">
        <p class="section-label">Password Reset</p>
        <h1>找回密码</h1>
        <p>输入注册邮箱。无论账号是否存在，页面都会使用统一提示。</p>
      </div>

      <form class="form-grid" @submit.prevent="handleSubmit">
        <label class="field">
          <span>注册邮箱</span>
          <span class="field-control"><input v-model="email" type="email" autocomplete="email" placeholder="name@example.com" /></span>
        </label>

        <div v-if="passwordResetRequestCaptchaEnabled" class="field">
          <span>安全验证码</span>
          <div class="captcha-row">
            <span class="field-control"><input v-model="captchaCode" type="text" maxlength="8" autocomplete="off" placeholder="图中字符" /></span>
            <button class="captcha-box" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
              <img v-if="captchaImage" :src="captchaImage" alt="找回密码验证码" />
              <span v-else>{{ captchaLoading ? '加载中...' : '刷新' }}</span>
            </button>
          </div>
        </div>

        <p v-if="sent" class="notice success">如果邮箱对应有效账号，重置链接会发送到该邮箱。</p>
        <p v-else-if="errorMessage || captchaError || submitHint" class="notice" :class="errorMessage || captchaError ? 'error' : 'info'">
          {{ errorMessage || captchaError || submitHint }}
        </p>

        <button class="btn btn-primary btn-block" type="submit" :disabled="!canSubmit">
          <span v-if="loading" class="spinner-small"></span>
          {{ loading ? '发送中...' : '发送重置链接' }}
        </button>
      </form>

      <RouterLink class="back-link" to="/login">返回登录</RouterLink>
    </section>
  </main>
</template>

<style scoped>
.auth-page {
  min-height: calc(100vh - 144px);
  display: grid;
  place-items: center;
  padding: 34px 16px 72px;
}
.auth-card {
  width: min(560px, 100%);
  display: grid;
  gap: 24px;
  padding: clamp(26px, 5vw, 44px);
}
.auth-heading {
  display: grid;
  gap: 10px;
}
.auth-heading h1 {
  font-size: clamp(2rem, 5vw, 3rem);
  line-height: 1;
  letter-spacing: -0.06em;
}
.auth-heading p:last-child,
.back-link {
  color: var(--c-text-2);
}
.form-grid {
  display: grid;
  gap: 16px;
}
.captcha-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 132px;
  gap: 10px;
}
.captcha-box {
  min-height: 48px;
  display: grid;
  place-items: center;
  overflow: hidden;
  border: 1px solid var(--c-border);
  border-radius: 12px;
  color: var(--c-text-2);
  background: var(--c-surface-dim);
  cursor: pointer;
}
.captcha-box img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.back-link {
  justify-self: center;
  font-weight: 800;
}
@media (max-width: 520px) {
  .captcha-row {
    grid-template-columns: 1fr;
  }
}
</style>
