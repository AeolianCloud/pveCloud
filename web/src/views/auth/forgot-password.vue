<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'

import { getPasswordResetRequestCaptcha, requestPasswordReset } from '../../api/auth'
import { useAuthCaptcha } from '../../composables/use-auth-captcha'
import { useWebAppStore } from '../../store/modules/app'

const appStore = useWebAppStore()
const { passwordResetRequestCaptchaEnabled, siteConfigError, siteConfigLoaded, siteConfigLoading, theme } = storeToRefs(appStore)

const email = ref('')
const loading = ref(false)
const errorMessage = ref('')
const sent = ref(false)

const {
  captchaCode,
  captchaError,
  captchaId,
  captchaImage,
  captchaLoading,
  captchaReady,
  refreshCaptcha,
} = useAuthCaptcha(passwordResetRequestCaptchaEnabled, getPasswordResetRequestCaptcha)

const canSubmit = computed(() => {
  return (
    siteConfigLoaded.value &&
    email.value.trim() !== '' &&
    (!passwordResetRequestCaptchaEnabled.value || captchaCode.value.trim().length >= 4) &&
    captchaReady.value &&
    !loading.value
  )
})

const submitHint = computed(() => {
  if (siteConfigLoading.value && !siteConfigLoaded.value) {
    return '正在加载找回密码配置，请稍候...'
  }
  if (email.value.trim() === '') {
    return '请输入注册邮箱'
  }
  if (passwordResetRequestCaptchaEnabled.value && !captchaReady.value) {
    return captchaError.value || '验证码加载中，请稍候...'
  }
  if (passwordResetRequestCaptchaEnabled.value && captchaCode.value.trim().length < 4) {
    return '请输入验证码后再发送重置链接'
  }
  if (siteConfigError.value) {
    return siteConfigError.value
  }
  return ''
})

const statusTone = computed(() => {
  if (captchaError.value || errorMessage.value) return 'danger'
  if (sent.value) return 'success'
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
    if (passwordResetRequestCaptchaEnabled.value) {
      void refreshCaptcha()
    }
  } catch (error) {
    errorMessage.value = errorText(error)
    if (passwordResetRequestCaptchaEnabled.value) {
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
          <p class="eyebrow">PASSWORD RESET</p>
          <h1>通过邮箱找回密码</h1>
          <p>如果邮箱对应有效账号，系统会发送一次性重置链接。页面不会暴露邮箱是否已经注册。</p>
        </div>
        <div class="brand-card"><span>安全策略</span><strong>统一提示，不暴露账号</strong></div>
      </aside>

      <section class="form-panel">
        <div class="form-card">
          <div class="form-heading form-heading--row">
            <div><p class="eyebrow">RESET</p><h2>找回密码</h2><p>输入注册邮箱，接收重置链接。</p></div>
            <RouterLink class="helper-pill" to="/login">返回登录</RouterLink>
          </div>
          <form class="auth-form" @submit.prevent="handleSubmit">
            <label class="field"><span>注册邮箱</span><span class="field-control"><input v-model="email" type="email" placeholder="输入注册邮箱" autocomplete="email" /></span></label>
            <div v-if="passwordResetRequestCaptchaEnabled" class="field"><span>安全验证码</span><div class="captcha-row"><span class="field-control"><input v-model="captchaCode" type="text" maxlength="8" placeholder="图中字符" autocomplete="off" /></span><button class="captcha-box" type="button" :disabled="captchaLoading" @click="refreshCaptcha"><img v-if="captchaImage" :src="captchaImage" alt="找回密码验证码"/><span v-else>{{ captchaLoading ? '加载中...' : '点击刷新' }}</span></button></div></div>
            <div class="status-message" :data-tone="statusTone"><span v-if="sent">如果邮箱对应有效账号，重置链接会发送到该邮箱。</span><span v-else-if="captchaError">{{ captchaError }}</span><span v-else-if="errorMessage">{{ errorMessage }}</span><span v-else>{{ submitHint || '确认邮箱后发送重置链接' }}</span></div>
            <button type="submit" class="submit-button" :disabled="!canSubmit"><span v-if="loading" class="spinner-small"></span>{{ loading ? '发送中...' : '发送重置链接' }}</button>
            <p class="switch-auth">想起密码了？<RouterLink to="/login">返回登录</RouterLink></p>
          </form>
        </div>
      </section>
    </section>
  </main>
</template>

<style scoped>
.account-page { --page-bg: radial-gradient(circle at 12% 8%, rgba(245,158,11,.18), transparent 30%), #080a12; --shell-bg: rgba(15,23,42,.72); --shell-border: rgba(255,255,255,.1); --panel-bg: linear-gradient(145deg, rgba(120,53,15,.88), rgba(15,23,42,.96)); --panel-text: #f8fafc; --panel-muted: #fde68a; --card-bg: rgba(255,255,255,.08); --form-bg: #101522; --field-bg: rgba(255,255,255,.055); --field-border: rgba(255,255,255,.11); --field-text: #f8fafc; --field-muted: #94a3b8; --shadow: 0 30px 90px rgba(0,0,0,.38); --accent: #f59e0b; --accent-strong: #d97706; min-height: 100vh; display: grid; place-items: center; padding: 40px; background: var(--page-bg); }
.account-page[data-mode="light"] { --page-bg: radial-gradient(circle at 12% 8%, rgba(245,158,11,.12), transparent 30%), #f6f8fc; --shell-bg: #fff; --shell-border: rgba(15,23,42,.08); --panel-bg: linear-gradient(145deg, #fffbeb, #fff 58%, #fffaf0); --panel-text: #0f172a; --panel-muted: #475569; --card-bg: rgba(255,255,255,.74); --form-bg: #fff; --field-bg: #f8fafc; --field-border: rgba(15,23,42,.11); --field-text: #0f172a; --field-muted: #64748b; --shadow: 0 30px 90px rgba(15,23,42,.12); --accent: #d97706; --accent-strong: #b45309; }
.account-page[data-mode="light"] .account-shell { animation: theme-shift-light 360ms ease; }
.account-page[data-mode="dark"] .account-shell { animation: theme-shift-dark 360ms ease; }
.account-shell { width: min(1120px, 100%); min-height: 680px; display: grid; grid-template-columns: minmax(340px,.9fr) minmax(440px,1.1fr); overflow: hidden; border: 1px solid var(--shell-border); border-radius: 32px; background: var(--shell-bg); box-shadow: var(--shadow); }
.brand-panel { position: relative; display: flex; flex-direction: column; justify-content: space-between; gap: 48px; padding: 48px; color: var(--panel-text); background: var(--panel-bg); }
.brand-panel::after { content: ''; position: absolute; right: -90px; bottom: -110px; width: 290px; height: 290px; border-radius: 50%; background: rgba(245,158,11,.2); filter: blur(18px); }
.brand-topbar,.brand-copy,.brand-card { position: relative; z-index: 1; }
.brand-topbar { display: flex; align-items: center; justify-content: space-between; gap: 20px; }
.brand-link { display: inline-flex; align-items: center; gap: 12px; width: fit-content; }
.brand-mark { width: 42px; height: 42px; display: grid; place-items: center; border-radius: 14px; color: #fff; background: linear-gradient(135deg, var(--accent), #2563eb); box-shadow: 0 16px 35px rgba(245,158,11,.25); }
.brand-mark svg { width: 24px; height: 24px; fill: none; stroke: currentColor; stroke-width: 1.8; stroke-linejoin: round; }
.brand-name { font-size: 1.25rem; font-weight: 800; letter-spacing: -.03em; }
.theme-switch,.helper-pill { height: 36px; padding: 0 14px; border: 1px solid var(--shell-border); border-radius: 999px; color: var(--panel-text); background: var(--card-bg); cursor: pointer; font-size: .86rem; font-weight: 800; }
.brand-copy h1 { margin-bottom: 22px; font-size: clamp(2.35rem, 4vw, 3.8rem); line-height: 1.04; letter-spacing: -.055em; }
.brand-copy p { color: var(--panel-muted); font-size: 1.05rem; line-height: 1.75; }
.eyebrow { margin-bottom: 14px; color: var(--accent); font-size: .78rem; font-weight: 800; letter-spacing: .14em; }
.brand-card { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 20px; border: 1px solid rgba(255,255,255,.12); border-radius: 22px; color: var(--panel-muted); background: var(--card-bg); }
.brand-card strong { color: var(--panel-text); }
.form-panel { display: grid; place-items: center; padding: 56px; background: var(--form-bg); }
.form-card { width: 100%; max-width: 430px; }
.form-heading { margin-bottom: 30px; }
.form-heading--row { display: flex; justify-content: space-between; gap: 20px; align-items: flex-start; }
.form-heading h2 { margin-bottom: 10px; color: var(--field-text); font-size: 2rem; line-height: 1.15; letter-spacing: -.04em; }
.form-heading p:last-child,.switch-auth { color: var(--field-muted); }
.auth-form { display: grid; gap: 20px; }
.field { display: grid; gap: 9px; color: var(--field-text); font-size: .9rem; font-weight: 700; }
.field-control { min-height: 54px; display: flex; align-items: center; padding: 0 16px; border: 1px solid var(--field-border); border-radius: 16px; background: var(--field-bg); }
.field-control:focus-within { border-color: var(--accent); box-shadow: 0 0 0 4px rgba(245,158,11,.16); }
.field-control input { width: 100%; color: var(--field-text); font-size: 1rem; }
.field-control input::placeholder { color: var(--field-muted); }
.captcha-row { display: grid; grid-template-columns: minmax(0,1fr) 132px; gap: 12px; }
.captcha-box { height: 54px; display: grid; place-items: center; overflow: hidden; border: 1px solid var(--field-border); border-radius: 16px; color: var(--field-muted); background: var(--field-bg); cursor: pointer; }
.captcha-box img { width: 100%; height: 100%; object-fit: cover; }
.status-message { padding: 13px 15px; border: 1px solid var(--field-border); border-radius: 16px; color: var(--field-muted); background: var(--field-bg); font-size: .9rem; line-height: 1.5; }
.status-message[data-tone="success"] { color: #34d399; border-color: rgba(16,185,129,.28); background: rgba(16,185,129,.1); }
.status-message[data-tone="danger"] { color: #fca5a5; border-color: rgba(239,68,68,.3); background: rgba(239,68,68,.12); }
.account-page[data-mode="light"] .status-message[data-tone="success"] { color: #047857; background: #ecfdf5; border-color: #a7f3d0; }
.account-page[data-mode="light"] .status-message[data-tone="danger"] { color: #b91c1c; background: #fef2f2; border-color: #fecaca; }
.submit-button { height: 56px; display: inline-flex; align-items: center; justify-content: center; gap: 10px; border-radius: 16px; color: #fff; background: linear-gradient(135deg, var(--accent), var(--accent-strong)); box-shadow: 0 18px 35px rgba(245,158,11,.24); cursor: pointer; font-weight: 800; }
.submit-button:disabled { cursor: not-allowed; opacity: .58; }
.switch-auth { margin-top: 6px; text-align: center; }.switch-auth a { color: var(--accent); font-weight: 800; margin-left: 4px; }
.spinner-small { width: 18px; height: 18px; border: 2px solid rgba(255,255,255,.35); border-top-color: #fff; border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes theme-shift-light { from { opacity: .82; transform: translateY(6px) scale(.992); } to { opacity: 1; transform: translateY(0) scale(1); } }
@keyframes theme-shift-dark { from { opacity: .82; transform: translateY(-6px) scale(.992); } to { opacity: 1; transform: translateY(0) scale(1); } }
.account-page,.account-shell,.brand-panel,.brand-panel::after,.brand-mark,.brand-card,.form-panel,.field-control,.captcha-box,.status-message,.theme-switch,.helper-pill,.submit-button { transition: background-color 360ms ease, border-color 360ms ease, box-shadow 360ms ease, color 360ms ease, opacity 360ms ease, transform 220ms ease; }
.brand-name,.brand-copy h1,.brand-copy p,.eyebrow,.brand-card strong,.form-heading h2,.form-heading p,.field,.field-control input,.field-control input::placeholder,.switch-auth { transition: color 360ms ease; }
.theme-switch:hover,.helper-pill:hover { transform: translateY(-1px); }
@media (max-width: 960px) { .account-page { padding: 20px; } .account-shell { min-height: auto; grid-template-columns: 1fr; border-radius: 24px; } .brand-panel { min-height: 320px; padding: 32px; } .brand-card { display: none; } .form-panel { padding: 34px 24px 38px; } }
@media (max-width: 520px) { .account-page { padding: 0; } .account-shell { min-height: 100vh; border: 0; border-radius: 0; } .brand-panel { min-height: 270px; padding: 26px; } .captcha-row { grid-template-columns: 1fr; } .captcha-box { width: 132px; } .form-heading--row { display: grid; } }
</style>
