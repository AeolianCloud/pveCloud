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
      <div class="hero-badge" style="color:var(--c-green); background:var(--c-green-soft);">
        <span style="width:6px;height:6px;border-radius:50%;background:var(--c-green);display:inline-block;"></span>
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
