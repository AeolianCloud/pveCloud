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

const {
  captchaCode,
  captchaError,
  captchaId,
  captchaImage,
  captchaLoading,
  captchaReady,
  refreshCaptcha,
} = useAuthCaptcha(passwordResetConfirmCaptchaEnabled, getPasswordResetConfirmCaptcha)

const token = computed(() => {
  const value = route.query.token
  return typeof value === 'string' ? value : ''
})
const canSubmit = computed(() => {
  return (
    siteConfigLoaded.value &&
    token.value !== '' &&
    password.value.length >= 6 &&
    password.value === confirmPassword.value &&
    (!passwordResetConfirmCaptchaEnabled.value || captchaCode.value.trim().length >= 4) &&
    captchaReady.value &&
    !loading.value
  )
})

const submitHint = computed(() => {
  if (siteConfigLoading.value && !siteConfigLoaded.value) {
    return '正在加载重置密码配置，请稍候...'
  }
  if (token.value === '') {
    return '重置链接缺少 token，请重新申请密码找回'
  }
  if (password.value.length < 6) {
    return '请输入至少 6 位新密码'
  }
  if (password.value !== confirmPassword.value) {
    return '两次输入的新密码需要保持一致'
  }
  if (passwordResetConfirmCaptchaEnabled.value && !captchaReady.value) {
    return captchaError.value || '验证码加载中，请稍候...'
  }
  if (passwordResetConfirmCaptchaEnabled.value && captchaCode.value.trim().length < 4) {
    return '请输入验证码后再提交'
  }
  if (siteConfigError.value) {
    return siteConfigError.value
  }
  return ''
})

function errorText(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (typeof error === 'object' && error !== null && 'request' in error) {
    return '网络连接失败，请检查后重试'
  }
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
    if (passwordResetConfirmCaptchaEnabled.value) {
      void refreshCaptcha()
    }
    window.setTimeout(() => {
      void router.replace('/login')
    }, 1200)
  } catch (error) {
    errorMessage.value = errorText(error)
    if (passwordResetConfirmCaptchaEnabled.value) {
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
      <div class="hero-badge">
        <span></span>
        重置密码
      </div>
      <h1>设置新密码</h1>
      <p>重置成功后，旧登录会话会失效。请使用新密码重新登录。</p>
    </div>
    <div class="auth-right">
      <form class="auth-form" @submit.prevent="handleSubmit">
        <h2>重置密码</h2>
        <p v-if="!token" class="hint error-text">重置链接缺少 token，请重新申请密码找回。</p>
        <label>
          <span>新密码</span>
          <input v-model="password" type="password" placeholder="至少 6 个字符" autocomplete="new-password" />
        </label>
        <label>
          <span>确认新密码</span>
          <input v-model="confirmPassword" type="password" placeholder="再次输入新密码" autocomplete="new-password" />
        </label>
        <div v-if="passwordResetConfirmCaptchaEnabled" class="captcha-field">
          <label>
            <span>验证码</span>
            <input v-model="captchaCode" type="text" maxlength="8" placeholder="请输入验证码" autocomplete="off" />
          </label>
          <div class="captcha-row">
            <img v-if="captchaImage" class="captcha-image" :src="captchaImage" alt="重置密码验证码" />
            <div v-else class="captcha-image captcha-image--placeholder">
              {{ captchaLoading ? '加载中...' : '暂无验证码' }}
            </div>
            <button class="captcha-refresh" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
              {{ captchaLoading ? '刷新中...' : '换一张' }}
            </button>
          </div>
        </div>
        <p v-if="password && confirmPassword && password !== confirmPassword" class="hint error-text">两次输入的密码不一致</p>
        <p v-if="captchaError" class="hint error-text">{{ captchaError }}</p>
        <p v-else-if="submitHint && !canSubmit" class="hint">{{ submitHint }}</p>
        <p v-if="siteConfigError" class="hint">{{ siteConfigError }}</p>
        <p v-if="done" class="hint success-text">密码已重置，正在返回登录页。</p>
        <p v-if="errorMessage" class="hint error-text">{{ errorMessage }}</p>
        <button class="btn btn-primary" type="submit" :disabled="!canSubmit">
          {{ loading ? '提交中...' : '确认重置' }}
        </button>
        <p class="hint">
          链接失效？<RouterLink class="link" to="/forgot-password">重新申请</RouterLink>
        </p>
      </form>
    </div>
  </section>
</template>

<style scoped>
.auth-page {
  min-height: calc(100vh - 80px);
  display: grid;
  grid-template-columns: minmax(320px, 0.92fr) minmax(420px, 1.08fr);
  gap: clamp(20px, 4vw, 48px);
  align-items: stretch;
  width: min(1120px, calc(100% - 40px));
  margin: 0 auto;
  padding: clamp(28px, 5vw, 70px) 0;
}

.auth-left,
.auth-right {
  border: 1px solid var(--c-border);
  border-radius: 32px;
  box-shadow: var(--shadow);
}

.auth-left {
  position: relative;
  display: grid;
  align-content: end;
  gap: 18px;
  min-height: 560px;
  padding: clamp(28px, 4vw, 46px);
  overflow: hidden;
  background:
    radial-gradient(circle at 14% 12%, rgba(16, 185, 129, 0.26), transparent 30%),
    radial-gradient(circle at 90% 0%, rgba(59, 130, 246, 0.2), transparent 34%),
    linear-gradient(145deg, rgba(15, 23, 42, 0.94), rgba(19, 21, 31, 0.74));
}

[data-theme='light'] .auth-left {
  background:
    radial-gradient(circle at 14% 12%, rgba(16, 185, 129, 0.18), transparent 30%),
    radial-gradient(circle at 90% 0%, rgba(37, 99, 235, 0.14), transparent 34%),
    #fff;
}

.auth-left::after {
  content: '';
  position: absolute;
  right: -70px;
  top: -70px;
  width: 230px;
  height: 230px;
  border-radius: 56px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  transform: rotate(18deg);
}

.hero-badge {
  width: fit-content;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 999px;
  color: var(--c-success);
  background: var(--c-success-soft);
  font-size: 0.82rem;
  font-weight: 800;
}

.hero-badge span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--c-success);
}

.auth-left h1 {
  max-width: 460px;
  font-size: clamp(2.5rem, 5vw, 4.5rem);
  line-height: 0.98;
  letter-spacing: -0.065em;
}

.auth-left p {
  max-width: 460px;
  color: var(--c-text-2);
  font-size: 1.06rem;
  line-height: 1.8;
}

.auth-right {
  display: grid;
  place-items: center;
  padding: clamp(24px, 4vw, 48px);
  background: var(--c-card);
}

.auth-form {
  width: min(100%, 460px);
  display: grid;
  gap: 16px;
}

.auth-form h2 {
  margin-bottom: 6px;
  font-size: 2rem;
  letter-spacing: -0.05em;
}

.auth-form label {
  display: grid;
  gap: 8px;
  color: var(--c-text-2);
  font-weight: 700;
}

.auth-form input {
  min-height: 50px;
  padding: 0 15px;
  border: 1px solid var(--c-border);
  border-radius: 15px;
  color: var(--c-text);
  background: var(--c-surface-dim);
}

.captcha-field {
  display: grid;
  gap: 12px;
}

.captcha-row {
  display: grid;
  grid-template-columns: 1fr 104px;
  gap: 10px;
  align-items: center;
}

.captcha-image {
  width: 100%;
  height: 52px;
  object-fit: cover;
  border: 1px solid var(--c-border);
  border-radius: 14px;
  background: var(--c-surface-dim);
}

.captcha-image--placeholder {
  display: grid;
  place-items: center;
  color: var(--c-text-3);
}

.captcha-refresh {
  height: 52px;
  border: 1px solid var(--c-border);
  border-radius: 14px;
  color: var(--c-text);
  background: var(--c-surface-dim);
  cursor: pointer;
  font-weight: 800;
}

.hint {
  margin: 0;
  color: var(--c-text-2);
  line-height: 1.7;
}

.success-text {
  color: var(--c-success);
}

.error-text {
  color: var(--c-error);
}

.link {
  color: var(--c-primary-h);
  font-weight: 800;
}

@media (max-width: 900px) {
  .auth-page {
    grid-template-columns: 1fr;
  }

  .auth-left {
    min-height: 340px;
  }
}

@media (max-width: 620px) {
  .auth-page {
    width: min(100% - 28px, 1120px);
    padding: 22px 0 44px;
  }

  .auth-left,
  .auth-right {
    border-radius: 24px;
  }

  .captcha-row {
    grid-template-columns: 1fr;
  }
}
</style>
