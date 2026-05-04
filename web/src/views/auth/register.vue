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
  <div class="auth-layout">
    <!-- Visual Left Side -->
    <div class="auth-visual" style="background: radial-gradient(circle at 100% 0%, rgba(16, 185, 129, 0.15) 0%, transparent 60%);">
      <div class="visual-bg"></div>
      <div class="visual-content">
        <RouterLink to="/" class="logo-link">
          <div class="logo-mark"><svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor"/></svg></div>
          <span class="logo-text">PVE<span class="text-gradient">Cloud</span></span>
        </RouterLink>
        
        <h1 class="visual-title">开启您的<br/>云端基础设施之旅</h1>
        <p class="visual-desc">创建一个账户，即刻拥有全球领先的弹性计算资源。几分钟内部署您的第一个实例。</p>
        
        <div class="visual-features">
          <div class="vf-item">
            <div class="vf-icon" style="color: var(--c-success); background: rgba(16, 185, 129, 0.1); border-color: rgba(16, 185, 129, 0.2);"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg></div>
            <div>
              <h4>按需付费，透明公开</h4>
              <p>没有任何隐藏费用，随时可随业务规模平滑升级。</p>
            </div>
          </div>
          <div class="vf-item">
            <div class="vf-icon" style="color: var(--c-success); background: rgba(16, 185, 129, 0.1); border-color: rgba(16, 185, 129, 0.2);"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg></div>
            <div>
              <h4>7x24 小时技术支持</h4>
              <p>专业的云架构师团队全天候为您排忧解难。</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Register Form Right Side -->
    <div class="auth-form-container">
      <div class="auth-form-wrap glass-panel">
        <div class="mobile-logo">
          <RouterLink to="/" class="logo-link" style="justify-content: center; margin-bottom: 24px;">
            <div class="logo-mark"><svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor"/></svg></div>
          </RouterLink>
        </div>
        
        <div class="form-header">
          <div>
            <h2 class="form-title">注册新账户</h2>
            <p class="form-subtitle">创建账户以访问控制台</p>
          </div>
          <RouterLink to="/login" class="login-link">返回登录</RouterLink>
        </div>

        <form class="auth-form" @submit.prevent="handleRegister">
          <div class="form-grid">
            <div class="input-group">
              <label>用户名 <span class="required">*</span></label>
              <div class="input-icon-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
                <input v-model="username" type="text" placeholder="至少3个字符" autocomplete="username" />
              </div>
            </div>

            <div class="input-group">
              <label>显示名称</label>
              <div class="input-icon-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
                <input v-model="displayName" type="text" placeholder="可选展示名称" autocomplete="name" />
              </div>
            </div>
          </div>

          <div class="input-group">
            <label>注册邮箱 <span class="required">*</span></label>
            <div class="input-icon-wrap">
              <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path><polyline points="22,6 12,13 2,6"></polyline></svg>
              <input v-model="email" type="email" placeholder="name@company.com" autocomplete="email" />
            </div>
          </div>
          
          <div class="form-grid">
            <div class="input-group">
              <label>密码 <span class="required">*</span></label>
              <div class="input-icon-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
                <input v-model="password" type="password" placeholder="至少 6 位" autocomplete="new-password" />
              </div>
            </div>

            <div class="input-group">
              <label>确认密码 <span class="required">*</span></label>
              <div class="input-icon-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
                <input v-model="confirmPassword" type="password" placeholder="重复密码" autocomplete="new-password" />
              </div>
            </div>
          </div>

          <!-- Captcha -->
          <div v-if="registerCaptchaEnabled" class="input-group captcha-group">
            <label>安全验证码 <span class="required">*</span></label>
            <div class="captcha-flex">
              <div class="input-icon-wrap" style="flex: 1;">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
                <input v-model="captchaCode" type="text" maxlength="8" placeholder="图中字符" autocomplete="off" />
              </div>
              <div class="captcha-img-box">
                <img v-if="captchaImage" :src="captchaImage" alt="验证码" @click="refreshCaptcha" class="captcha-img" />
                <div v-else class="captcha-placeholder" @click="refreshCaptcha">
                  {{ captchaLoading ? '加载中...' : '点击刷新' }}
                </div>
              </div>
            </div>
          </div>
          
          <!-- Alerts -->
          <div v-if="errorMessage || submitHint" class="alert-box" :class="{ 'is-error': errorMessage || (submitHint && !siteConfigLoading) }">
            <svg v-if="errorMessage || (submitHint && !siteConfigLoading)" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
            {{ errorMessage || submitHint }}
          </div>
          
          <button type="submit" class="btn btn-primary btn-block submit-btn" :disabled="!canSubmit">
            <span v-if="loading" class="spinner-small"></span>
            {{ loading ? '正在创建账户...' : '注册并进入控制台' }}
          </button>
        </form>

      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 45% 55%;
  background: var(--c-bg);
}

/* Left Visual Area */
.auth-visual {
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px;
}
.visual-bg {
  position: absolute; inset: 0; z-index: 0;
  background: 
    linear-gradient(135deg, rgba(9, 10, 15, 0.4) 0%, rgba(9, 10, 15, 0.9) 100%),
    url('data:image/svg+xml;utf8,<svg width="40" height="40" xmlns="http://www.w3.org/2000/svg"><rect width="40" height="40" fill="none"/><circle cx="20" cy="20" r="1" fill="rgba(255,255,255,0.1)"/></svg>') repeat;
  background-size: cover, 40px 40px;
}
.visual-bg::before {
  content: ''; position: absolute; inset: -50%;
  background: radial-gradient(circle at center, rgba(16, 185, 129, 0.1) 0%, transparent 50%);
  animation: pulseBg 10s ease-in-out infinite alternate;
}
@keyframes pulseBg { 0% { transform: scale(0.8) translate(-10%, -10%); } 100% { transform: scale(1.2) translate(10%, 10%); } }

.visual-content { position: relative; z-index: 1; max-width: 500px; }
.logo-link { display: inline-flex; align-items: center; gap: 12px; margin-bottom: 60px; }
.logo-mark { width: 32px; height: 32px; color: var(--c-primary); }
.logo-text { font-size: 1.5rem; font-weight: 800; color: var(--c-text); }

.visual-title { font-size: 3rem; font-weight: 800; line-height: 1.2; margin-bottom: 24px; letter-spacing: -0.02em; }
.visual-desc { font-size: 1.15rem; color: var(--c-text-2); line-height: 1.6; margin-bottom: 48px; }

.visual-features { display: grid; gap: 32px; }
.vf-item { display: flex; gap: 16px; align-items: flex-start; }
.vf-icon { width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; border: 1px solid; }
.vf-icon svg { width: 24px; height: 24px; }
.vf-item h4 { font-size: 1.1rem; font-weight: 700; margin-bottom: 6px; color: var(--c-text); }
.vf-item p { font-size: 0.95rem; color: var(--c-text-2); }

/* Right Form Area */
.auth-form-container {
  display: flex; align-items: center; justify-content: center;
  padding: 40px; position: relative;
}
.auth-form-wrap {
  width: 100%; max-width: 540px; padding: 48px; border-radius: var(--radius-xl);
}
.mobile-logo { display: none; }

.form-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 32px; }
.form-title { font-size: 2rem; font-weight: 800; margin-bottom: 8px; color: var(--c-text); }
.form-subtitle { font-size: 1rem; color: var(--c-text-2); }
.login-link { font-size: 0.9rem; font-weight: 600; color: var(--c-primary); padding: 8px 16px; background: var(--c-primary-soft); border-radius: 99px; }

.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.input-group { margin-bottom: 20px; }
.input-group label { display: block; font-size: 0.9rem; font-weight: 600; color: var(--c-text); margin-bottom: 8px; }
.required { color: var(--c-error); }

.input-icon-wrap { position: relative; }
.input-icon { position: absolute; left: 16px; top: 50%; transform: translateY(-50%); width: 18px; height: 18px; color: var(--c-text-3); transition: color 0.3s; }
.input-icon-wrap input {
  width: 100%; background: rgba(0,0,0,0.2); border: 1px solid var(--c-border);
  padding: 14px 16px 14px 44px; border-radius: var(--radius); font-size: 0.95rem; color: var(--c-text);
  transition: all 0.3s;
}
.input-icon-wrap input:focus { border-color: var(--c-primary); box-shadow: 0 0 0 4px var(--c-primary-soft); background: rgba(0,0,0,0.4); }

.captcha-flex { display: flex; gap: 12px; }
.captcha-img-box { width: 130px; height: 48px; border-radius: var(--radius); overflow: hidden; border: 1px solid var(--c-border); background: var(--c-surface-dim); cursor: pointer; }
.captcha-img { width: 100%; height: 100%; object-fit: cover; }
.captcha-placeholder { width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; font-size: 0.8rem; color: var(--c-text-3); }

.alert-box {
  display: flex; align-items: center; gap: 8px; padding: 12px 16px;
  background: var(--c-surface-dim); border: 1px solid var(--c-border);
  border-radius: var(--radius-sm); color: var(--c-text-2); font-size: 0.9rem; margin-bottom: 24px;
}
.alert-box.is-error { background: var(--c-error-soft); border-color: rgba(239, 68, 68, 0.3); color: #fca5a5; }
.alert-box svg { width: 18px; height: 18px; flex-shrink: 0; }

.submit-btn { height: 52px; font-size: 1.1rem; border-radius: var(--radius); position: relative; }
.spinner-small {
  width: 18px; height: 18px; border: 2px solid rgba(255,255,255,0.3); border-top-color: #fff;
  border-radius: 50%; animation: spin 1s linear infinite; margin-right: 8px;
}
@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 992px) {
  .auth-layout { grid-template-columns: 1fr; }
  .auth-visual { display: none; }
  .mobile-logo { display: block; }
  .auth-form-wrap { padding: 32px 24px; border: none; background: transparent; box-shadow: none; }
  .form-grid { grid-template-columns: 1fr; gap: 0; }
}
</style>
