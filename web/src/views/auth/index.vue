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

const { captchaCode, captchaError, captchaId, captchaImage, captchaLoading, captchaReady, refreshCaptcha } =
  useAuthCaptcha(loginCaptchaEnabled, getLoginCaptcha)

const canSubmit = computed(() => (
  siteConfigLoaded.value &&
  account.value.trim() !== '' &&
  password.value.length >= 6 &&
  (!loginCaptchaEnabled.value || captchaCode.value.trim().length >= 4) &&
  captchaReady.value &&
  !loading.value
))

const submitHint = computed(() => {
  if (siteConfigLoading.value && !siteConfigLoaded.value) return '正在同步登录配置...'
  if (loginCaptchaEnabled.value && !captchaReady.value) return captchaError.value || '验证码准备中...'
  if (loginCaptchaEnabled.value && captchaCode.value.trim().length < 4) return '请输入验证码后再登录'
  if (siteConfigError.value) return siteConfigError.value
  return ''
})

function loginErrorMessage(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { status?: number; data?: { message?: string } } }).response
    if (response?.status === 403 && response.data?.message) return response.data.message
  }
  return '账号或密码错误，请重新输入'
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
    if (loginCaptchaEnabled.value) void refreshCaptcha()
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
    <section class="auth-shell surface">
      <aside class="auth-aside">
        <p class="section-label">Sign In</p>
        <h1>登录用户中心</h1>
        <p>进入控制台后可以维护资料、查看实名状态，并浏览后续购买入口。</p>
      </aside>

      <form class="auth-card" @submit.prevent="handleLogin">
        <div class="auth-heading">
          <h2>欢迎回来</h2>
          <p>使用账号或邮箱登录。</p>
        </div>

        <label class="field">
          <span>注册邮箱或账号</span>
          <span class="field-control"><input v-model="account" type="text" autocomplete="username" placeholder="name@example.com" /></span>
        </label>

        <label class="field">
          <span class="field-row">密码 <RouterLink to="/forgot-password">忘记密码？</RouterLink></span>
          <span class="field-control"><input v-model="password" type="password" autocomplete="current-password" placeholder="输入登录密码" /></span>
        </label>

        <div v-if="loginCaptchaEnabled" class="field">
          <span>安全验证码</span>
          <div class="captcha-row">
            <span class="field-control"><input v-model="captchaCode" type="text" maxlength="8" autocomplete="off" placeholder="图中字符" /></span>
            <button class="captcha-box" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
              <img v-if="captchaImage" :src="captchaImage" alt="登录验证码" />
              <span v-else>{{ captchaLoading ? '加载中...' : '刷新' }}</span>
            </button>
          </div>
        </div>

        <p v-if="errorMessage || captchaError || submitHint" class="notice error">{{ errorMessage || captchaError || submitHint }}</p>

        <button class="btn btn-primary btn-block" type="submit" :disabled="!canSubmit">
          <span v-if="loading" class="spinner-small"></span>
          {{ loading ? '验证中...' : '登录' }}
        </button>

        <p class="switch-line">暂无账号？<RouterLink to="/register">注册新账户</RouterLink></p>
      </form>
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

.auth-shell {
  width: min(1040px, 100%);
  display: grid;
  grid-template-columns: minmax(280px, 0.9fr) minmax(360px, 1.1fr);
  overflow: hidden;
}

.auth-aside {
  display: grid;
  align-content: end;
  gap: 16px;
  min-height: 560px;
  padding: clamp(28px, 5vw, 48px);
  color: #fff;
  background:
    linear-gradient(135deg, rgba(29, 78, 216, 0.95), rgba(15, 23, 42, 0.92)),
    var(--c-primary);
}

.auth-aside h1 {
  font-size: clamp(2.2rem, 5vw, 4rem);
  line-height: 1;
  letter-spacing: -0.06em;
}

.auth-aside p:last-child {
  max-width: 460px;
  color: rgba(255, 255, 255, 0.78);
  line-height: 1.75;
}

.auth-card {
  display: grid;
  align-content: center;
  gap: 18px;
  padding: clamp(28px, 5vw, 52px);
  background: var(--c-surface-strong);
}

.auth-heading {
  display: grid;
  gap: 8px;
  margin-bottom: 8px;
}

.auth-heading h2 {
  font-size: 2rem;
  letter-spacing: -0.05em;
}

.auth-heading p,
.switch-line {
  color: var(--c-text-2);
}

.field-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.field-row a,
.switch-line a {
  color: var(--c-primary);
  font-weight: 800;
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

.switch-line {
  text-align: center;
}

@media (max-width: 820px) {
  .auth-shell {
    grid-template-columns: 1fr;
  }

  .auth-aside {
    min-height: 260px;
  }
}

@media (max-width: 520px) {
  .captcha-row {
    grid-template-columns: 1fr;
  }
}
</style>
