<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useWebAuthStore } from '../../store/modules/auth'

const route = useRoute()
const router = useRouter()
const authStore = useWebAuthStore()

const account = ref('')
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

const canSubmit = computed(() => account.value.trim() !== '' && password.value.length >= 6 && !loading.value)

function resolveRedirect(value: unknown) {
  if (typeof value !== 'string') return '/user'
  if (!value.startsWith('/') || value.startsWith('//')) return '/user'
  return value
}

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
    await authStore.login({ account: account.value.trim(), password: password.value })
    await router.replace(resolveRedirect(route.query.redirect))
  } catch (error) {
    errorMessage.value = loginErrorMessage(error)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-layout">
    <!-- Visual Left Side -->
    <div class="auth-visual">
      <div class="visual-content">
        <RouterLink to="/" class="logo-link">
          <div class="logo-mark"><svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor"/></svg></div>
          <span class="logo-text">PVE<span class="text-gradient">Cloud</span></span>
        </RouterLink>
        
        <h1 class="visual-title">极速、智能的<br/>云端控制中心</h1>
        <p class="visual-desc">一站式管理您的弹性计算实例、网络架构与存储资源。构建高可用架构从未如此简单。</p>
        
        <div class="visual-features">
          <div class="vf-item">
            <div class="vf-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg></div>
            <div>
              <h4>企业级安全认证</h4>
              <p>端到端数据加密与多因素认证支持</p>
            </div>
          </div>
          <div class="vf-item">
            <div class="vf-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg></div>
            <div>
              <h4>全球节点调度</h4>
              <p>实时监控全球可用区状态与资源分布</p>
            </div>
          </div>
        </div>
      </div>
      <div class="visual-bg"></div>
    </div>

    <!-- Login Form Right Side -->
    <div class="auth-form-container">
      <div class="auth-form-wrap glass-panel">
        <div class="mobile-logo">
          <RouterLink to="/" class="logo-link" style="justify-content: center; margin-bottom: 24px;">
            <div class="logo-mark"><svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor"/></svg></div>
          </RouterLink>
        </div>
        
        <h2 class="form-title">登录控制台</h2>
        <p class="form-subtitle">欢迎回来！请输入您的凭据以访问控制面板。</p>

        <form class="auth-form" @submit.prevent="handleLogin">
          <div class="input-group">
            <label>注册邮箱或账号</label>
            <div class="input-icon-wrap">
              <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path><polyline points="22,6 12,13 2,6"></polyline></svg>
              <input v-model="account" type="text" placeholder="name@company.com" autocomplete="username" />
            </div>
          </div>
          
          <div class="input-group">
            <div class="label-row">
              <label>密码</label>
              <a href="#" class="forgot-link">忘记密码？</a>
            </div>
            <div class="input-icon-wrap">
              <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
              <input v-model="password" type="password" placeholder="••••••••" autocomplete="current-password" />
            </div>
          </div>
          
          <div v-if="errorMessage" class="error-alert">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
            {{ errorMessage }}
          </div>
          
          <button type="submit" class="btn btn-primary btn-block submit-btn" :disabled="!canSubmit">
            <span v-if="loading" class="spinner-small"></span>
            {{ loading ? '验证身份中...' : '安全登录' }}
          </button>
        </form>

        <div class="auth-footer">
          <p>暂无账号？<a href="#" class="register-link">联系管理员获取邀请</a></p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 1fr 1fr;
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
  background: radial-gradient(circle at center, rgba(59, 130, 246, 0.15) 0%, transparent 50%);
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
.vf-icon { width: 48px; height: 48px; border-radius: 12px; background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.1); display: flex; align-items: center; justify-content: center; color: var(--c-primary); flex-shrink: 0; }
.vf-icon svg { width: 24px; height: 24px; }
.vf-item h4 { font-size: 1.1rem; font-weight: 700; margin-bottom: 6px; color: var(--c-text); }
.vf-item p { font-size: 0.95rem; color: var(--c-text-2); }

/* Right Form Area */
.auth-form-container {
  display: flex; align-items: center; justify-content: center;
  padding: 40px; position: relative;
}
.auth-form-wrap {
  width: 100%; max-width: 440px; padding: 48px; border-radius: var(--radius-xl);
}
.mobile-logo { display: none; }
.form-title { font-size: 2rem; font-weight: 800; margin-bottom: 8px; color: var(--c-text); }
.form-subtitle { font-size: 1rem; color: var(--c-text-2); margin-bottom: 32px; }

.input-group { margin-bottom: 24px; }
.label-row { display: flex; justify-content: space-between; align-items: baseline; }
.input-group label { display: block; font-size: 0.9rem; font-weight: 600; color: var(--c-text); margin-bottom: 8px; }
.forgot-link { font-size: 0.85rem; color: var(--c-primary); font-weight: 600; }
.forgot-link:hover { text-decoration: underline; }

.input-icon-wrap { position: relative; }
.input-icon { position: absolute; left: 16px; top: 50%; transform: translateY(-50%); width: 20px; height: 20px; color: var(--c-text-3); transition: color 0.3s; }
.input-icon-wrap input {
  width: 100%; background: rgba(0,0,0,0.2); border: 1px solid var(--c-border);
  padding: 16px 16px 16px 48px; border-radius: var(--radius); font-size: 1rem; color: var(--c-text);
  transition: all 0.3s;
}
.input-icon-wrap input:focus { border-color: var(--c-primary); box-shadow: 0 0 0 4px var(--c-primary-soft); background: rgba(0,0,0,0.4); }
.input-icon-wrap input:focus + .input-icon { color: var(--c-primary); } /* Note: this doesn't work perfectly in CSS unless icon is after input, handled visually by input outline */

.error-alert {
  display: flex; align-items: center; gap: 8px; padding: 12px 16px;
  background: var(--c-error-soft); border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: var(--radius-sm); color: #fca5a5; font-size: 0.9rem; margin-bottom: 24px;
}
.error-alert svg { width: 18px; height: 18px; flex-shrink: 0; }

.submit-btn { height: 52px; font-size: 1.1rem; border-radius: var(--radius); position: relative; }
.spinner-small {
  width: 18px; height: 18px; border: 2px solid rgba(255,255,255,0.3); border-top-color: #fff;
  border-radius: 50%; animation: spin 1s linear infinite; margin-right: 8px;
}
@keyframes spin { to { transform: rotate(360deg); } }

.auth-footer { text-align: center; margin-top: 32px; font-size: 0.95rem; color: var(--c-text-2); }
.register-link { color: var(--c-primary); font-weight: 600; margin-left: 4px; }
.register-link:hover { text-decoration: underline; }

@media (max-width: 992px) {
  .auth-layout { grid-template-columns: 1fr; }
  .auth-visual { display: none; }
  .mobile-logo { display: block; }
  .auth-form-wrap { padding: 32px 24px; border: none; background: transparent; box-shadow: none; }
}
</style>
