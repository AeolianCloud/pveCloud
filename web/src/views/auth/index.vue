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
const { loginCaptchaEnabled, siteConfigLoaded } = storeToRefs(appStore)

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

const highlights = [
  '登录和控制台使用同一套入口',
  '刷新页面后会自动恢复登录态',
  '退出后会清理本地登录信息',
]

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
  <section class="page auth-page">
    <div class="auth-left">
      <div class="hero-badge" style="color:var(--c-primary); background:var(--c-primary-soft);">
        <span style="width:6px;height:6px;border-radius:50%;background:var(--c-primary);display:inline-block;"></span>
        用户登录
      </div>
      <h1>进入云资源控制台</h1>
      <p>
        登录后进入统一控制台入口。当前阶段先开放登录态、会话恢复和退出能力，实例、订单和工单仍按占位展示。
      </p>
      <div style="display:grid; gap:10px; margin-top:8px;">
        <div
          v-for="item in highlights"
          :key="item"
          style="display:flex; align-items:center; gap:8px; font-size:.92rem; color:var(--c-text-2);"
        >
          <span style="width:5px;height:5px;border-radius:50%;background:var(--c-primary);flex-shrink:0;"></span>
          {{ item }}
        </div>
      </div>
    </div>
    <div class="auth-right">
      <form class="auth-form" @submit.prevent="handleLogin">
        <h2>登录</h2>
        <label>
          <span>邮箱或用户名</span>
          <input v-model="account" type="text" placeholder="请输入邮箱或用户名" autocomplete="username" />
        </label>
        <label>
          <span>密码</span>
          <input v-model="password" type="password" placeholder="请输入密码" autocomplete="current-password" />
        </label>
        <div v-if="loginCaptchaEnabled" class="captcha-field">
          <label>
            <span>验证码</span>
            <input v-model="captchaCode" type="text" maxlength="8" placeholder="请输入验证码" autocomplete="off" />
          </label>
          <div class="captcha-row">
            <img v-if="captchaImage" class="captcha-image" :src="captchaImage" alt="登录验证码" />
            <div v-else class="captcha-image captcha-image--placeholder">
              {{ captchaLoading ? '加载中...' : '暂无验证码' }}
            </div>
            <button class="captcha-refresh" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
              {{ captchaLoading ? '刷新中...' : '换一张' }}
            </button>
          </div>
        </div>
        <p v-if="captchaError" class="hint error-text">{{ captchaError }}</p>
        <p v-if="errorMessage" class="hint error-text">{{ errorMessage }}</p>
        <button class="btn btn-primary" type="submit" :disabled="!canSubmit" style="width:100%">
          {{ loading ? '登录中...' : '登录' }}
        </button>
        <p class="hint">
          还没有账号？<RouterLink class="link" to="/register">立即注册</RouterLink>
          <span style="padding:0 6px;color:var(--c-text-3);">/</span>
          <RouterLink class="link" to="/forgot-password">忘记密码</RouterLink>
        </p>
      </form>
    </div>
  </section>
</template>
