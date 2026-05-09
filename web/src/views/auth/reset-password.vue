<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'

import { confirmPasswordReset, getPasswordResetConfirmCaptcha } from '../../api/auth'
import { useAuthCaptcha } from '../../composables/use-auth-captcha'
import { useWebAppStore } from '../../store/modules/app'

const route = useRoute()
const router = useRouter()
const appStore = useWebAppStore()
const { passwordResetConfirmCaptchaEnabled, siteConfigError, siteConfigLoaded, siteConfigLoading } = storeToRefs(appStore)

const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMessage = ref('')
const done = ref(false)

const { captchaCode, captchaError, captchaId, captchaImage, captchaLoading, captchaReady, refreshCaptcha } =
  useAuthCaptcha(passwordResetConfirmCaptchaEnabled, getPasswordResetConfirmCaptcha)

const token = computed(() => {
  const value = route.query.token
  return typeof value === 'string' ? value : ''
})

const canSubmit = computed(() => (
  siteConfigLoaded.value &&
  token.value !== '' &&
  password.value.length >= 6 &&
  password.value === confirmPassword.value &&
  (!passwordResetConfirmCaptchaEnabled.value || captchaCode.value.trim().length >= 4) &&
  captchaReady.value &&
  !loading.value
))

const submitHint = computed(() => {
  if (siteConfigLoading.value && !siteConfigLoaded.value) return '正在加载重置密码配置...'
  if (token.value === '') return '重置链接缺少 token，请重新申请密码找回'
  if (password.value.length < 6) return '请输入至少 6 位新密码'
  if (password.value !== confirmPassword.value) return '两次输入的新密码需要保持一致'
  if (passwordResetConfirmCaptchaEnabled.value && !captchaReady.value) return captchaError.value || '验证码加载中...'
  if (passwordResetConfirmCaptchaEnabled.value && captchaCode.value.trim().length < 4) return '请输入验证码'
  if (siteConfigError.value) return siteConfigError.value
  return ''
})

function errorText(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (typeof error === 'object' && error !== null && 'request' in error) return '网络连接失败，请检查后重试'
  return '密码重置失败，请重新申请重置链接'
}

async function handleSubmit() {
  if (!canSubmit.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    await confirmPasswordReset({
      token: token.value,
      password: password.value,
      captcha_id: passwordResetConfirmCaptchaEnabled.value ? captchaId.value : undefined,
      captcha_code: passwordResetConfirmCaptchaEnabled.value ? captchaCode.value.trim() : undefined,
    })
    done.value = true
    if (passwordResetConfirmCaptchaEnabled.value) void refreshCaptcha()
    window.setTimeout(() => {
      void router.replace('/login')
    }, 1200)
  } catch (error) {
    errorMessage.value = errorText(error)
    if (passwordResetConfirmCaptchaEnabled.value) void refreshCaptcha()
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
        <p class="section-label">Reset Password</p>
        <h1>设置新密码</h1>
        <p>重置成功后旧会话会失效，请使用新密码重新登录。</p>
      </div>

      <form class="form-grid" @submit.prevent="handleSubmit">
        <label class="field"><span>新密码</span><span class="field-control"><input v-model="password" type="password" autocomplete="new-password" placeholder="至少 6 位" /></span></label>
        <label class="field"><span>确认新密码</span><span class="field-control"><input v-model="confirmPassword" type="password" autocomplete="new-password" placeholder="重复新密码" /></span></label>

        <div v-if="passwordResetConfirmCaptchaEnabled" class="field">
          <span>安全验证码</span>
          <div class="captcha-row">
            <span class="field-control"><input v-model="captchaCode" type="text" maxlength="8" autocomplete="off" placeholder="图中字符" /></span>
            <button class="captcha-box" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
              <img v-if="captchaImage" :src="captchaImage" alt="重置密码验证码" />
              <span v-else>{{ captchaLoading ? '加载中...' : '刷新' }}</span>
            </button>
          </div>
        </div>

        <p v-if="done" class="notice success">密码已重置，正在返回登录页。</p>
        <p v-else-if="errorMessage || captchaError || submitHint" class="notice" :class="errorMessage || captchaError ? 'error' : 'info'">
          {{ errorMessage || captchaError || submitHint }}
        </p>

        <button class="btn btn-primary btn-block" type="submit" :disabled="!canSubmit">
          <span v-if="loading" class="spinner-small"></span>
          {{ loading ? '提交中...' : '确认重置' }}
        </button>
      </form>

      <RouterLink class="back-link" to="/forgot-password">重新申请重置链接</RouterLink>
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
