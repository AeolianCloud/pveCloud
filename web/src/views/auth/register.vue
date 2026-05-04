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
const { registerCaptchaEnabled, siteConfigError, siteConfigLoaded, siteConfigLoading, theme } = storeToRefs(appStore)

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
    return '正在同步注册配置...'
  }
  if (username.value.trim().length > 0 && username.value.trim().length < 3) {
    return '用户名至少需要 3 个字符'
  }
  if (password.value.length > 0 && password.value.length < 6) {
    return '密码至少需要 6 位'
  }
  if (confirmPassword.value && password.value !== confirmPassword.value) {
    return '两次输入的密码不一致'
  }
  if (registerCaptchaEnabled.value && !captchaReady.value) {
    return captchaError.value || '验证码准备中...'
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
  <main class="account-page" :data-mode="theme">
    <section class="account-shell">
      <aside class="brand-panel">
        <div class="brand-topbar">
          <RouterLink to="/" class="brand-link">
            <span class="brand-mark"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 2 2 7l10 5 10-5-10-5Z"/><path d="m2 12 10 5 10-5"/><path d="m2 17 10 5 10-5"/></svg></span>
            <span class="brand-name">PVECloud</span>
          </RouterLink>
          <button class="theme-switch" type="button" @click="appStore.toggleTheme()">{{ theme === 'dark' ? '浅色' : '深色' }}</button>
        </div>
        <div class="brand-copy">
          <p class="eyebrow">CREATE ACCOUNT</p>
          <h1>创建用户端账号</h1>
          <p>注册后可进入用户中心，维护账号资料，并查看当前开放的服务器产品展示入口。</p>
        </div>
        <div class="brand-card"><span>账号能力</span><strong>注册 / 登录 / 找回密码</strong></div>
      </aside>

      <section class="form-panel">
        <div class="form-card form-card--wide">
          <div class="form-heading form-heading--row">
            <div><p class="eyebrow">SIGN UP</p><h2>注册新账户</h2><p>填写基础信息以创建用户端账号。</p></div>
            <RouterLink class="helper-pill" to="/login">返回登录</RouterLink>
          </div>

          <form class="auth-form" @submit.prevent="handleRegister">
            <div class="field-grid">
              <label class="field"><span>用户名 <b>*</b></span><span class="field-control"><input v-model="username" type="text" placeholder="至少 3 个字符" autocomplete="username" /></span></label>
              <label class="field"><span>显示名称</span><span class="field-control"><input v-model="displayName" type="text" placeholder="可选展示名称" autocomplete="name" /></span></label>
            </div>
            <label class="field"><span>注册邮箱 <b>*</b></span><span class="field-control"><input v-model="email" type="email" placeholder="name@company.com" autocomplete="email" /></span></label>
            <div class="field-grid">
              <label class="field"><span>密码 <b>*</b></span><span class="field-control"><input v-model="password" type="password" placeholder="至少 6 位" autocomplete="new-password" /></span></label>
              <label class="field"><span>确认密码 <b>*</b></span><span class="field-control"><input v-model="confirmPassword" type="password" placeholder="重复密码" autocomplete="new-password" /></span></label>
            </div>
            <div v-if="registerCaptchaEnabled" class="field"><span>安全验证码 <b>*</b></span><div class="captcha-row"><span class="field-control"><input v-model="captchaCode" type="text" maxlength="8" placeholder="图中字符" autocomplete="off" /></span><button class="captcha-box" type="button" @click="refreshCaptcha"><img v-if="captchaImage" :src="captchaImage" alt="验证码"/><span v-else>{{ captchaLoading ? '加载中...' : '点击刷新' }}</span></button></div></div>
            <div v-if="errorMessage || submitHint" class="alert-message"><span>{{ errorMessage || submitHint }}</span></div>
            <button type="submit" class="submit-button" :disabled="!canSubmit"><span v-if="loading" class="spinner-small"></span>{{ loading ? '正在创建账户...' : '注册并进入控制台' }}</button>
          </form>
        </div>
      </section>
    </section>
  </main>
</template>

<style scoped>
.account-page { --page-bg: radial-gradient(circle at 12% 8%, rgba(16,185,129,.2), transparent 30%), #080a12; --shell-bg: rgba(15,23,42,.72); --shell-border: rgba(255,255,255,.1); --panel-bg: linear-gradient(145deg, rgba(20,83,45,.88), rgba(15,23,42,.96)); --panel-text: #f8fafc; --panel-muted: #bbf7d0; --card-bg: rgba(255,255,255,.08); --form-bg: #101522; --field-bg: rgba(255,255,255,.055); --field-border: rgba(255,255,255,.11); --field-text: #f8fafc; --field-muted: #94a3b8; --shadow: 0 30px 90px rgba(0,0,0,.38); --accent: #10b981; --accent-strong: #059669; min-height: 100vh; display: grid; place-items: center; padding: 40px; background: var(--page-bg); }
.account-page[data-mode="light"] { --page-bg: radial-gradient(circle at 12% 8%, rgba(16,185,129,.12), transparent 30%), #f6f8fc; --shell-bg: #fff; --shell-border: rgba(15,23,42,.08); --panel-bg: linear-gradient(145deg, #ecfdf5, #fff 58%, #f7fffb); --panel-text: #0f172a; --panel-muted: #475569; --card-bg: rgba(255,255,255,.74); --form-bg: #fff; --field-bg: #f8fafc; --field-border: rgba(15,23,42,.11); --field-text: #0f172a; --field-muted: #64748b; --shadow: 0 30px 90px rgba(15,23,42,.12); --accent: #059669; --accent-strong: #047857; }
.account-page[data-mode="light"] .form-card { animation: card-slide-light 360ms cubic-bezier(.22,1,.36,1) both; }
.account-page[data-mode="dark"] .form-card { animation: card-slide-dark 360ms cubic-bezier(.22,1,.36,1) both; }
.account-shell { width: min(1180px, 100%); min-height: 720px; display: grid; grid-template-columns: minmax(340px,.86fr) minmax(520px,1.14fr); overflow: hidden; border: 1px solid var(--shell-border); border-radius: 32px; background: var(--shell-bg); box-shadow: var(--shadow); }
.brand-panel { position: relative; display: flex; flex-direction: column; justify-content: space-between; gap: 48px; padding: 48px; color: var(--panel-text); background: var(--panel-bg); }
.brand-panel::after { content: ''; position: absolute; right: -90px; bottom: -110px; width: 290px; height: 290px; border-radius: 50%; background: rgba(16,185,129,.22); filter: blur(18px); }
.brand-topbar,.brand-copy,.brand-card { position: relative; z-index: 1; }
.brand-topbar { display: flex; align-items: center; justify-content: space-between; gap: 20px; }
.brand-link { display: inline-flex; align-items: center; gap: 12px; width: fit-content; }
.brand-mark { width: 42px; height: 42px; display: grid; place-items: center; border-radius: 14px; color: #fff; background: linear-gradient(135deg, var(--accent), #2563eb); box-shadow: 0 16px 35px rgba(16,185,129,.28); }
.brand-mark svg { width: 24px; height: 24px; fill: none; stroke: currentColor; stroke-width: 1.8; stroke-linejoin: round; }
.brand-name { font-size: 1.25rem; font-weight: 800; letter-spacing: -.03em; }
.theme-switch,.helper-pill { height: 36px; padding: 0 14px; border: 1px solid var(--shell-border); border-radius: 999px; color: var(--panel-text); background: var(--card-bg); cursor: pointer; font-size: .86rem; font-weight: 800; }
.brand-copy h1 { margin-bottom: 22px; font-size: clamp(2.35rem, 4vw, 3.8rem); line-height: 1.04; letter-spacing: -.055em; }
.brand-copy p { color: var(--panel-muted); font-size: 1.05rem; line-height: 1.75; }
.eyebrow { margin-bottom: 14px; color: var(--accent); font-size: .78rem; font-weight: 800; letter-spacing: .14em; }
.brand-card { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 20px; border: 1px solid rgba(255,255,255,.12); border-radius: 22px; color: var(--panel-muted); background: var(--card-bg); }
.brand-card strong { color: var(--panel-text); }
.form-panel { display: grid; place-items: center; padding: 48px; background: var(--form-bg); }
.form-card { width: 100%; max-width: 560px; }
.form-heading { margin-bottom: 30px; }
.form-heading--row { display: flex; justify-content: space-between; gap: 20px; align-items: flex-start; }
.form-heading h2 { margin-bottom: 10px; color: var(--field-text); font-size: 2rem; line-height: 1.15; letter-spacing: -.04em; }
.form-heading p:last-child { color: var(--field-muted); }
.auth-form { display: grid; gap: 18px; }
.field-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.field { display: grid; gap: 9px; color: var(--field-text); font-size: .9rem; font-weight: 700; }
.field b { color: #ef4444; }
.field-control { min-height: 52px; display: flex; align-items: center; padding: 0 16px; border: 1px solid var(--field-border); border-radius: 16px; color: var(--field-muted); background: var(--field-bg); transition: border-color var(--transition-fast), box-shadow var(--transition-fast); }
.field-control:focus-within { border-color: var(--accent); box-shadow: 0 0 0 4px rgba(16,185,129,.16); }
.field-control input { width: 100%; color: var(--field-text); font-size: 1rem; }
.field-control input::placeholder { color: var(--field-muted); }
.captcha-row { display: grid; grid-template-columns: minmax(0,1fr) 132px; gap: 12px; }
.captcha-box { height: 52px; display: grid; place-items: center; overflow: hidden; border: 1px solid var(--field-border); border-radius: 16px; color: var(--field-muted); background: var(--field-bg); cursor: pointer; }
.captcha-box img { width: 100%; height: 100%; object-fit: cover; }
.alert-message { padding: 13px 15px; border: 1px solid rgba(239,68,68,.3); border-radius: 16px; color: #fca5a5; background: rgba(239,68,68,.12); font-size: .9rem; }
.account-page[data-mode="light"] .alert-message { color: #b91c1c; background: #fef2f2; border-color: #fecaca; }
.submit-button { height: 56px; display: inline-flex; align-items: center; justify-content: center; gap: 10px; border-radius: 16px; color: #fff; background: linear-gradient(135deg, var(--accent), var(--accent-strong)); box-shadow: 0 18px 35px rgba(16,185,129,.26); cursor: pointer; font-weight: 800; }
.submit-button:disabled { cursor: not-allowed; opacity: .58; }
.spinner-small { width: 18px; height: 18px; border: 2px solid rgba(255,255,255,.35); border-top-color: #fff; border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.account-page,.account-shell,.brand-panel,.brand-panel::after,.brand-mark,.brand-card,.form-panel,.field-control,.captcha-box,.alert-message,.theme-switch,.helper-pill,.submit-button { transition: background-color 320ms ease, border-color 320ms ease, box-shadow 320ms ease, color 320ms ease; }
.brand-name,.brand-copy h1,.brand-copy p,.eyebrow,.brand-card strong,.form-heading h2,.form-heading p,.field,.field-control input,.field-control input::placeholder { transition: color 360ms ease; }
.theme-switch:hover,.helper-pill:hover { transform: translateY(-1px); }
@keyframes card-slide-light { from { opacity: .84; transform: translateX(18px); filter: brightness(.98); } to { opacity: 1; transform: translateX(0); filter: brightness(1); } }
@keyframes card-slide-dark { from { opacity: .84; transform: translateX(-18px); filter: brightness(1.08); } to { opacity: 1; transform: translateX(0); filter: brightness(1); } }
@media (max-width: 980px) { .account-page { padding: 20px; } .account-shell { min-height: auto; grid-template-columns: 1fr; border-radius: 24px; } .brand-panel { min-height: 320px; padding: 32px; } .brand-card { display: none; } .form-panel { padding: 34px 24px 38px; } }
@media (max-width: 620px) { .account-page { padding: 0; } .account-shell { min-height: 100vh; border: 0; border-radius: 0; } .brand-panel { min-height: 270px; padding: 26px; } .field-grid,.captcha-row { grid-template-columns: 1fr; } .captcha-box { width: 132px; } .form-heading--row { display: grid; } }
</style>
