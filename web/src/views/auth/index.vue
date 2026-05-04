<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'

import { getLoginCaptcha } from '../../api/auth'
import { useAuthCaptcha } from '../../composables/use-auth-captcha'
import { useWebAppStore } from '../../store/modules/app'
import { useWebAuthStore } from '../../store/modules/auth'
import { resolveWebRedirect } from '../../utils/web-auth'

const route = useRoute()
const router = useRouter()
const authStore = useWebAuthStore()
const appStore = useWebAppStore()
const { loginCaptchaEnabled, siteConfigError, siteConfigLoaded, siteConfigLoading } = storeToRefs(appStore)

const account = ref('')
const password = ref('')
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
} = useAuthCaptcha(loginCaptchaEnabled, getLoginCaptcha)

const canSubmit = computed(() => {
  return (
    siteConfigLoaded.value &&
    account.value.trim() !== '' &&
    password.value.length >= 6 &&
    (!loginCaptchaEnabled.value || captchaCode.value.trim().length >= 4) &&
    captchaReady.value &&
    !loading.value
  )
})

const submitHint = computed(() => {
  if (siteConfigLoading.value && !siteConfigLoaded.value) {
    return '正在同步登录配置，请稍候...'
  }
  if (account.value.trim() === '') {
    return '先输入邮箱或用户名'
  }
  if (password.value.length < 6) {
    return '密码至少需要 6 位'
  }
  if (loginCaptchaEnabled.value && !captchaReady.value) {
    return captchaError.value || '验证码正在准备中...'
  }
  if (loginCaptchaEnabled.value && captchaCode.value.trim().length < 4) {
    return '请输入验证码后再继续'
  }
  if (siteConfigError.value) {
    return siteConfigError.value
  }
  return '表单已就绪，可以登录'
})

const statusTone = computed(() => {
  if (captchaError.value || errorMessage.value) return 'danger'
  if (!canSubmit.value) return 'muted'
  return 'success'
})

function loginErrorMessage(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { status?: number; data?: { message?: string } } }).response

    if (response?.status === 401) {
      return '账号或密码错误'
    }
    if ((response?.status === 400 || response?.status === 403 || response?.status === 429) && response.data?.message) {
      return response.data.message
    }
    if (response?.status && response.status >= 500) {
      return '登录服务暂时不可用，请稍后再试'
    }
    if (response?.data?.message) {
      return response.data.message
    }
  }

  if (typeof error === 'object' && error !== null && 'request' in error) {
    return '网络连接失败，请检查后重试'
  }

  return '登录失败，请稍后再试'
}

async function handleLogin() {
  if (!canSubmit.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    await authStore.login({
      account: account.value.trim(),
      password: password.value,
      captcha_id: loginCaptchaEnabled.value ? captchaId.value : undefined,
      captcha_code: loginCaptchaEnabled.value ? captchaCode.value.trim() : undefined,
    })
    await router.replace(resolveWebRedirect(route.query.redirect))
  } catch (error) {
    errorMessage.value = loginErrorMessage(error)
    if (loginCaptchaEnabled.value) {
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
  <section class="page auth-page auth-page--login">
    <div class="auth-ambient auth-ambient--blue"></div>
    <div class="auth-ambient auth-ambient--violet"></div>

    <div class="auth-stage">
      <aside class="auth-panel auth-panel--story">
        <div class="auth-kicker auth-kicker--primary">pveCloud</div>
        <h1 class="auth-display">登录云资源控制台</h1>
        <p class="auth-copy">
          使用邮箱或用户名进入控制台。登录态会在刷新后自动恢复，退出后清理本地凭据。
        </p>
      </aside>

      <div class="auth-panel auth-panel--form-shell">
        <div class="auth-form-card">
          <div class="auth-form-card__header">
            <div>
              <p class="auth-eyebrow">Welcome back</p>
              <h2>登录你的账号</h2>
            </div>
            <RouterLink class="auth-mini-link" to="/register">创建账号</RouterLink>
          </div>

          <form class="auth-form auth-form--stacked" @submit.prevent="handleLogin">
            <label class="auth-field">
              <span>邮箱或用户名</span>
              <input v-model="account" type="text" placeholder="例如 demo@example.com / demo-user" autocomplete="username" />
            </label>

            <label class="auth-field">
              <span>密码</span>
              <input v-model="password" type="password" placeholder="输入你的登录密码" autocomplete="current-password" />
            </label>

            <div v-if="loginCaptchaEnabled" class="captcha-field auth-fieldset">
              <div class="auth-fieldset__legend">
                <span>安全校验</span>
                <small>登录验证码已开启</small>
              </div>
              <label class="auth-field">
                <span>验证码</span>
                <input v-model="captchaCode" type="text" maxlength="8" placeholder="输入图中字符" autocomplete="off" />
              </label>
              <div class="captcha-row captcha-row--panel">
                <img v-if="captchaImage" class="captcha-image" :src="captchaImage" alt="登录验证码" />
                <div v-else class="captcha-image captcha-image--placeholder">
                  {{ captchaLoading ? '加载中...' : '暂无验证码' }}
                </div>
                <button class="captcha-refresh" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
                  {{ captchaLoading ? '刷新中...' : '刷新验证码' }}
                </button>
              </div>
            </div>

            <div class="auth-status" :data-tone="statusTone">
              <p v-if="captchaError" class="hint error-text">{{ captchaError }}</p>
              <p v-else-if="errorMessage" class="hint error-text">{{ errorMessage }}</p>
              <p v-else class="hint">{{ submitHint }}</p>
            </div>

            <button class="btn btn-primary auth-submit" type="submit" :disabled="!canSubmit">
              {{ loading ? '登录中...' : '进入控制台' }}
            </button>

            <div class="auth-secondary-row">
              <span class="hint">还没有账号？</span>
              <RouterLink class="link" to="/register">立即注册</RouterLink>
              <span class="auth-divider"></span>
              <RouterLink class="link" to="/forgot-password">忘记密码</RouterLink>
            </div>
          </form>
        </div>
      </div>
    </div>
  </section>
</template>
