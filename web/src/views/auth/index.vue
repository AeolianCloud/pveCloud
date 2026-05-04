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
const { loginCaptchaEnabled, siteConfigError, siteConfigLoaded, siteConfigLoading, theme } = storeToRefs(appStore)

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
  if (siteConfigLoading.value && !siteConfigLoaded.value) return '正在同步登录配置...'
  if (loginCaptchaEnabled.value && !captchaReady.value) return captchaError.value || '验证码准备中...'
  if (loginCaptchaEnabled.value && captchaCode.value.trim().length < 4) return '请输入验证码后再登录'
  if (siteConfigError.value) return siteConfigError.value
  return ''
})

function loginErrorMessage(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { status?: number; data?: { message?: string } } }).response
    if (response?.status === 403 && response.data?.message) {
      return response.data.message
    }
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
  <main class="login-page" :data-mode="theme">
    <section class="login-shell" aria-label="用户登录">
      <aside class="brand-panel">
        <div class="brand-topbar">
          <RouterLink to="/" class="brand-link" aria-label="返回首页">
            <span class="brand-mark">
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M12 2 2 7l10 5 10-5-10-5Z" />
                <path d="m2 12 10 5 10-5" />
                <path d="m2 17 10 5 10-5" />
              </svg>
            </span>
            <span class="brand-name">PVECloud</span>
          </RouterLink>
          <button class="theme-switch" type="button" @click="appStore.toggleTheme()">
            {{ theme === 'dark' ? '浅色' : '深色' }}
          </button>
        </div>

        <div class="brand-copy">
          <p class="eyebrow">USER CONSOLE</p>
          <h1>登录用户中心</h1>
          <p>管理账号资料，查看服务器产品目录，并保持用户端会话同步。</p>
        </div>

        <div class="brand-card">
          <div class="metric-row">
            <span>当前入口</span>
            <strong>用户端</strong>
          </div>
          <div class="metric-row">
            <span>会话状态</span>
            <strong>安全认证</strong>
          </div>
        </div>
      </aside>

      <section class="form-panel">
        <div class="form-card">
          <div class="form-heading">
            <p class="eyebrow">SIGN IN</p>
            <h2>欢迎回来</h2>
            <p>请输入账号信息继续访问控制台。</p>
          </div>

          <form class="auth-form" @submit.prevent="handleLogin">
            <label class="field">
              <span>注册邮箱或账号</span>
              <span class="field-control">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                  <path d="M4 4h16v16H4z" />
                  <path d="m22 6-10 7L2 6" />
                </svg>
                <input v-model="account" type="text" placeholder="name@company.com" autocomplete="username" />
              </span>
            </label>

            <label class="field">
              <span class="field-label-row">
                <span>密码</span>
                <RouterLink to="/forgot-password" class="helper-link">忘记密码？</RouterLink>
              </span>
              <span class="field-control">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                  <rect x="3" y="11" width="18" height="10" rx="2" />
                  <path d="M7 11V8a5 5 0 0 1 10 0v3" />
                </svg>
                <input v-model="password" type="password" placeholder="输入登录密码" autocomplete="current-password" />
              </span>
            </label>

            <div v-if="loginCaptchaEnabled" class="field captcha-field">
              <span>安全验证码</span>
              <div class="captcha-row">
                <span class="field-control captcha-input">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z" />
                  </svg>
                  <input v-model="captchaCode" type="text" maxlength="8" placeholder="图中字符" autocomplete="off" />
                </span>
                <button class="captcha-box" type="button" @click="refreshCaptcha">
                  <img v-if="captchaImage" :src="captchaImage" alt="登录验证码" />
                  <span v-else>{{ captchaLoading ? '加载中...' : '点击刷新' }}</span>
                </button>
              </div>
            </div>

            <div v-if="errorMessage || captchaError || submitHint" class="alert-message">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                <circle cx="12" cy="12" r="10" />
                <path d="M12 8v4" />
                <path d="M12 16h.01" />
              </svg>
              <span>{{ errorMessage || captchaError || submitHint }}</span>
            </div>

            <button type="submit" class="submit-button" :disabled="!canSubmit">
              <span v-if="loading" class="spinner-small"></span>
              {{ loading ? '验证身份中...' : '登录控制台' }}
            </button>
          </form>

          <p class="switch-auth">暂无账号？<RouterLink to="/register">注册新账户</RouterLink></p>
        </div>
      </section>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  --login-bg: radial-gradient(circle at 12% 8%, rgba(59, 130, 246, 0.22), transparent 30%), #080a12;
  --shell-bg: rgba(15, 23, 42, 0.72);
  --shell-border: rgba(255, 255, 255, 0.1);
  --panel-bg: linear-gradient(145deg, rgba(30, 41, 59, 0.92), rgba(15, 23, 42, 0.96));
  --panel-text: #f8fafc;
  --panel-muted: #a8b3c7;
  --panel-card-bg: rgba(255, 255, 255, 0.08);
  --form-bg: #101522;
  --field-bg: rgba(255, 255, 255, 0.055);
  --field-border: rgba(255, 255, 255, 0.11);
  --field-text: #f8fafc;
  --field-muted: #94a3b8;
  --shadow: 0 30px 90px rgba(0, 0, 0, 0.38);
  --accent: #3b82f6;
  --accent-strong: #2563eb;

  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 40px;
  background: var(--login-bg);
}

.login-shell {
  width: min(1120px, 100%);
  min-height: 680px;
  display: grid;
  grid-template-columns: minmax(360px, 0.9fr) minmax(420px, 1fr);
  overflow: hidden;
  border: 1px solid var(--shell-border);
  border-radius: 32px;
  background: var(--shell-bg);
  box-shadow: var(--shadow);
  transition: background-color 360ms ease, border-color 360ms ease, box-shadow 360ms ease;
}

.login-page[data-mode="light"] .login-shell { animation: theme-shift-light 360ms ease; }
.login-page[data-mode="dark"] .login-shell { animation: theme-shift-dark 360ms ease; }

.brand-panel {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px;
  color: var(--panel-text);
  background: var(--panel-bg);
  transition: background 360ms ease, color 360ms ease;
}

.brand-panel::after {
  content: '';
  position: absolute;
  right: -80px;
  bottom: -100px;
  width: 280px;
  height: 280px;
  border-radius: 50%;
  background: rgba(59, 130, 246, 0.24);
  filter: blur(18px);
  transition: background-color 360ms ease, opacity 360ms ease;
}

.brand-topbar {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.brand-link {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  width: fit-content;
}

.theme-switch {
  height: 36px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 999px;
  color: var(--panel-text);
  background: var(--panel-card-bg);
  cursor: pointer;
  font-size: 0.86rem;
  font-weight: 800;
}

.brand-mark {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 14px;
  color: #fff;
  background: linear-gradient(135deg, var(--accent), #7c3aed);
  box-shadow: 0 16px 35px rgba(37, 99, 235, 0.35);
  transition: background 360ms ease, box-shadow 360ms ease, transform 220ms ease;
}

.brand-link:hover .brand-mark { transform: translateY(-1px); }

.brand-mark svg {
  width: 24px;
  height: 24px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linejoin: round;
}

.brand-name {
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  transition: color 360ms ease;
}

.brand-copy {
  position: relative;
  z-index: 1;
  max-width: 420px;
}

.eyebrow {
  margin-bottom: 14px;
  color: var(--accent);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.14em;
}

.brand-copy h1 {
  margin-bottom: 22px;
  font-size: clamp(2.5rem, 4vw, 4rem);
  line-height: 1.04;
  letter-spacing: -0.055em;
}

.brand-copy p {
  color: var(--panel-muted);
  font-size: 1.05rem;
  line-height: 1.75;
}

.brand-card {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 14px;
  padding: 20px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 22px;
  background: var(--panel-card-bg);
  transition: background-color 360ms ease, border-color 360ms ease, box-shadow 360ms ease;
}

.metric-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  color: var(--panel-muted);
  font-size: 0.92rem;
  transition: color 360ms ease;
}

.metric-row strong {
  color: var(--panel-text);
  font-size: 0.96rem;
  transition: color 360ms ease;
}

.form-panel {
  display: grid;
  place-items: center;
  padding: 56px;
  background: var(--form-bg);
  transition: background-color 360ms ease;
}

.form-card {
  width: 100%;
  max-width: 430px;
}

.form-heading {
  margin-bottom: 34px;
}

.form-heading h2 {
  margin-bottom: 10px;
  color: var(--field-text);
  font-size: 2.15rem;
  line-height: 1.15;
  letter-spacing: -0.04em;
  transition: color 360ms ease;
}

.form-heading p:last-child {
  color: var(--field-muted);
  transition: color 360ms ease;
}

.auth-form {
  display: grid;
  gap: 20px;
}

.field {
  display: grid;
  gap: 9px;
  color: var(--field-text);
  font-size: 0.9rem;
  font-weight: 700;
  transition: color 360ms ease;
}

.field-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.helper-link,
.switch-auth a {
  color: var(--accent);
  font-weight: 800;
  transition: color 220ms ease;
}

.helper-link:hover,
.switch-auth a:hover {
  color: var(--accent-strong);
}

.field-control {
  min-height: 54px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
  border: 1px solid var(--field-border);
  border-radius: 16px;
  color: var(--field-muted);
  background: var(--field-bg);
  transition: background-color 360ms ease, border-color 360ms ease, box-shadow var(--transition-fast), color 360ms ease;
}

.field-control:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.16);
}

.field-control svg {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.field-control input {
  width: 100%;
  color: var(--field-text);
  font-size: 1rem;
  transition: color 360ms ease;
}

.field-control input::placeholder {
  color: var(--field-muted);
  transition: color 360ms ease;
}

.captcha-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 132px;
  gap: 12px;
}

.captcha-box {
  height: 54px;
  display: grid;
  place-items: center;
  overflow: hidden;
  border: 1px solid var(--field-border);
  border-radius: 16px;
  color: var(--field-muted);
  background: var(--field-bg);
  cursor: pointer;
  transition: background-color 360ms ease, border-color 360ms ease, color 360ms ease;
}

.captcha-box img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.alert-message {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 13px 15px;
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 16px;
  color: #fca5a5;
  background: rgba(239, 68, 68, 0.12);
  font-size: 0.9rem;
  line-height: 1.5;
  transition: background-color 360ms ease, border-color 360ms ease, color 360ms ease;
}

.alert-message svg {
  width: 19px;
  height: 19px;
  flex-shrink: 0;
  margin-top: 1px;
}

.submit-button {
  height: 56px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-radius: 16px;
  color: #fff;
  background: linear-gradient(135deg, var(--accent), var(--accent-strong));
  box-shadow: 0 18px 35px rgba(37, 99, 235, 0.28);
  cursor: pointer;
  font-weight: 800;
  transition: transform var(--transition-fast), box-shadow var(--transition-fast), opacity var(--transition-fast);
}

.submit-button:not(:disabled):hover {
  transform: translateY(-1px);
  box-shadow: 0 22px 42px rgba(37, 99, 235, 0.34);
}

.submit-button:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.spinner-small {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.switch-auth {
  margin-top: 26px;
  color: var(--field-muted);
  text-align: center;
  transition: color 360ms ease;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes theme-shift-light {
  from { opacity: 0.82; transform: translateY(6px) scale(0.992); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

@keyframes theme-shift-dark {
  from { opacity: 0.82; transform: translateY(-6px) scale(0.992); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.login-page[data-mode="light"] {
  --login-bg: radial-gradient(circle at 12% 8%, rgba(37, 99, 235, 0.12), transparent 30%), #f6f8fc;
  --shell-bg: #ffffff;
  --shell-border: rgba(15, 23, 42, 0.08);
  --panel-bg: linear-gradient(145deg, #eef5ff, #ffffff 58%, #f8fbff);
  --panel-text: #0f172a;
  --panel-muted: #475569;
  --panel-card-bg: rgba(255, 255, 255, 0.74);
  --form-bg: #ffffff;
  --field-bg: #f8fafc;
  --field-border: rgba(15, 23, 42, 0.11);
  --field-text: #0f172a;
  --field-muted: #64748b;
  --shadow: 0 30px 90px rgba(15, 23, 42, 0.12);
  --accent: #2563eb;
  --accent-strong: #1d4ed8;
}

.login-page[data-mode="light"] .brand-card {
  border-color: rgba(37, 99, 235, 0.12);
  box-shadow: 0 18px 50px rgba(15, 23, 42, 0.08);
}

.login-page[data-mode="light"] .brand-panel::after {
  background: rgba(37, 99, 235, 0.12);
}

.login-page[data-mode="light"] .theme-switch {
  border-color: rgba(37, 99, 235, 0.14);
}

.login-page[data-mode="light"] .alert-message {
  color: #b91c1c;
  background: #fef2f2;
  border-color: #fecaca;
}

@media (max-width: 960px) {
  .login-page {
    padding: 20px;
  }

  .login-shell {
    min-height: auto;
    grid-template-columns: 1fr;
    border-radius: 24px;
  }

  .brand-panel {
    min-height: 340px;
    padding: 32px;
  }

  .brand-card {
    display: none;
  }

  .form-panel {
    padding: 34px 24px 38px;
  }
}

@media (max-width: 520px) {
  .login-page {
    padding: 0;
  }

  .login-shell {
    min-height: 100vh;
    border: 0;
    border-radius: 0;
  }

  .brand-panel {
    min-height: 280px;
    padding: 26px;
  }

  .brand-copy h1 {
    font-size: 2.25rem;
  }

  .captcha-row {
    grid-template-columns: 1fr;
  }

  .captcha-box {
    width: 132px;
  }
}
</style>
