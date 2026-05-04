<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'

import { getRegisterCaptcha } from '../../api/auth'
import { useAuthCaptcha } from '../../composables/use-auth-captcha'
import { useWebAppStore } from '../../store/modules/app'
import { useWebAuthStore } from '../../store/modules/auth'

const router = useRouter()
const authStore = useWebAuthStore()
const appStore = useWebAppStore()
const { registerCaptchaEnabled, siteConfigError, siteConfigLoaded, siteConfigLoading } = storeToRefs(appStore)

const username = ref('')
const email = ref('')
const displayName = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMessage = ref('')

const {
  captchaCode,
  captchaError,
  captchaId,
  captchaImage,
  captchaLoading,
  captchaReady,
  refreshCaptcha,
} = useAuthCaptcha(registerCaptchaEnabled, getRegisterCaptcha)

const canSubmit = computed(() => {
  return (
    siteConfigLoaded.value &&
    username.value.trim().length >= 3 &&
    email.value.trim() !== '' &&
    password.value.length >= 6 &&
    password.value === confirmPassword.value &&
    (!registerCaptchaEnabled.value || captchaCode.value.trim().length >= 4) &&
    captchaReady.value &&
    !loading.value
  )
})

const submitHint = computed(() => {
  if (siteConfigLoading.value && !siteConfigLoaded.value) {
    return '正在同步注册配置，请稍候...'
  }
  if (username.value.trim().length < 3) {
    return '用户名至少需要 3 个字符'
  }
  if (email.value.trim() === '') {
    return '请输入邮箱地址'
  }
  if (password.value.length < 6) {
    return '密码至少需要 6 位'
  }
  if (password.value !== confirmPassword.value) {
    return '两次输入的密码需要保持一致'
  }
  if (registerCaptchaEnabled.value && !captchaReady.value) {
    return captchaError.value || '验证码正在准备中...'
  }
  if (registerCaptchaEnabled.value && captchaCode.value.trim().length < 4) {
    return '请输入验证码后再继续'
  }
  if (siteConfigError.value) {
    return siteConfigError.value
  }
  return '表单已就绪，可以创建账号'
})

const statusTone = computed(() => {
  if (captchaError.value || errorMessage.value || (password.value && confirmPassword.value && password.value !== confirmPassword.value)) {
    return 'danger'
  }
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
      captcha_id: registerCaptchaEnabled.value ? captchaId.value : undefined,
      captcha_code: registerCaptchaEnabled.value ? captchaCode.value.trim() : undefined,
    })
    await router.replace('/user')
  } catch (error) {
    errorMessage.value = errorText(error)
    if (registerCaptchaEnabled.value) {
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
  <section class="page auth-page auth-page--register">
    <div class="auth-ambient auth-ambient--emerald"></div>
    <div class="auth-ambient auth-ambient--sun"></div>

    <div class="auth-stage auth-stage--register">
      <aside class="auth-panel auth-panel--story auth-panel--story-compact">
        <div class="auth-kicker auth-kicker--success">pveCloud</div>
        <h1 class="auth-display">创建云资源账号</h1>
        <p class="auth-copy">
          注册成功后会直接进入控制台。当前阶段只创建用户账号，不自动生成订单、实例或钱包数据。
        </p>
      </aside>

      <div class="auth-panel auth-panel--form-shell">
        <div class="auth-form-card auth-form-card--wide">
          <div class="auth-form-card__header">
            <div>
              <p class="auth-eyebrow">Get started</p>
              <h2>创建用户账号</h2>
            </div>
            <RouterLink class="auth-mini-link" to="/login">返回登录</RouterLink>
          </div>

          <form class="auth-form auth-form--grid" @submit.prevent="handleRegister">
            <label class="auth-field">
              <span>用户名</span>
              <input v-model="username" type="text" placeholder="至少 3 个字符" autocomplete="username" />
            </label>

            <label class="auth-field">
              <span>显示名称</span>
              <input v-model="displayName" type="text" placeholder="可选，用于页面展示" autocomplete="name" />
            </label>

            <label class="auth-field auth-field--full">
              <span>邮箱</span>
              <input v-model="email" type="email" placeholder="用于登录、找回密码和通知" autocomplete="email" />
            </label>

            <label class="auth-field">
              <span>密码</span>
              <input v-model="password" type="password" placeholder="至少 6 个字符" autocomplete="new-password" />
            </label>

            <label class="auth-field">
              <span>确认密码</span>
              <input v-model="confirmPassword" type="password" placeholder="再次输入密码" autocomplete="new-password" />
            </label>

            <div v-if="registerCaptchaEnabled" class="captcha-field auth-fieldset auth-field--full">
              <div class="auth-fieldset__legend">
                <span>安全校验</span>
                <small>注册验证码已开启</small>
              </div>
              <div class="auth-captcha-grid">
                <label class="auth-field">
                  <span>验证码</span>
                  <input v-model="captchaCode" type="text" maxlength="8" placeholder="输入图中字符" autocomplete="off" />
                </label>
                <div class="captcha-row captcha-row--panel captcha-row--stretch">
                  <img v-if="captchaImage" class="captcha-image" :src="captchaImage" alt="注册验证码" />
                  <div v-else class="captcha-image captcha-image--placeholder">
                    {{ captchaLoading ? '加载中...' : '暂无验证码' }}
                  </div>
                  <button class="captcha-refresh" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
                    {{ captchaLoading ? '刷新中...' : '刷新验证码' }}
                  </button>
                </div>
              </div>
            </div>

            <div class="auth-status auth-field--full" :data-tone="statusTone">
              <p v-if="password && confirmPassword && password !== confirmPassword" class="hint error-text">两次输入的密码不一致</p>
              <p v-else-if="captchaError" class="hint error-text">{{ captchaError }}</p>
              <p v-else-if="errorMessage" class="hint error-text">{{ errorMessage }}</p>
              <p v-else class="hint">{{ submitHint }}</p>
            </div>

            <div class="auth-actions auth-field--full">
              <button class="btn btn-primary auth-submit" type="submit" :disabled="!canSubmit">
                {{ loading ? '注册中...' : '注册并进入控制台' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </section>
</template>
