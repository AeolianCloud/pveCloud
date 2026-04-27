<script setup lang="ts">
import { Key, Lock, User } from '@element-plus/icons-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getAdminLoginCaptcha } from '../../api/auth'
import { ADMIN_ROUTE_PATH, normalizeAdminRedirect } from '../../router/constants'
import { useAuthStore } from '../../store/modules/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const form = reactive({
  username: '',
  password: '',
  captchaCode: '',
})

const loading = ref(false)
const captchaLoading = ref(false)
const errorMessage = ref('')
const captcha = reactive({
  id: '',
  image: '',
})

const canSubmit = computed(
  () =>
    form.username.trim().length > 0 &&
    form.password.length >= 6 &&
    captcha.id.length > 0 &&
    form.captchaCode.trim().length >= 4,
)

async function loadCaptcha() {
  captchaLoading.value = true
  try {
    const result = await getAdminLoginCaptcha()
    captcha.id = result.captcha_id
    captcha.image = result.image
    form.captchaCode = ''
  } catch (error) {
    captcha.id = ''
    captcha.image = ''
    errorMessage.value = error instanceof Error ? error.message : '验证码加载失败，请稍后重试'
  } finally {
    captchaLoading.value = false
  }
}

async function submit() {
  if (!canSubmit.value || loading.value) {
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    await authStore.login({
      username: form.username.trim(),
      password: form.password,
      captcha_id: captcha.id,
      captcha_code: form.captchaCode.trim(),
    })
    const redirect = normalizeAdminRedirect(route.query.redirect, ADMIN_ROUTE_PATH.dashboard)
    await router.replace(redirect)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败，请检查账号、密码或验证码'
    await loadCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadCaptcha()
})
</script>

<template>
  <main class="login-page">
    <!-- 装饰光晕 -->
    <div class="login-page__glow login-page__glow--top" />
    <div class="login-page__glow login-page__glow--bottom" />
    <div class="login-page__glow login-page__glow--side" />

    <!-- 背景网格（纯 CSS） -->
    <div class="login-page__grid" />

    <div class="login-page__container">
      <!-- 左侧品牌区域 -->
      <div class="login-page__brand">
        <div class="login-page__brand-content">
          <div class="login-page__logo">
            <svg class="login-page__logo-icon" viewBox="0 0 40 40" fill="none">
              <rect width="40" height="40" rx="10" fill="currentColor" opacity="0.2"/>
              <path d="M12 28V12h6a6 6 0 0 1 0 12h-2m-4-6h4a2 2 0 0 1 0 4h-4v-4Z" fill="currentColor"/>
              <circle cx="28" cy="20" r="6" fill="currentColor" opacity="0.4"/>
            </svg>
            <span class="login-page__logo-text">PVE Cloud</span>
          </div>
          <h1 class="login-page__title">管理控制台</h1>
          <p class="login-page__desc">
            Proxmox VE 云管理平台，高效管理您的虚拟化基础设施
          </p>
          <div class="login-page__features">
            <div class="login-page__feature-item">
              <span class="login-page__feature-dot" />
              统一资源监控
            </div>
            <div class="login-page__feature-item">
              <span class="login-page__feature-dot" />
              快捷虚拟机部署
            </div>
            <div class="login-page__feature-item">
              <span class="login-page__feature-dot" />
              多集群管理
            </div>
          </div>
        </div>
        <div class="login-page__footer-text">
          &copy; 2024 PVE Cloud. All rights reserved.
        </div>
      </div>

      <!-- 右侧登录卡片 -->
      <div class="login-page__card">
        <div class="login-page__card-inner">
          <div class="login-page__card-header">
            <div class="login-page__card-avatar">
              <svg viewBox="0 0 40 40" fill="none">
                <rect width="40" height="40" rx="10" fill="currentColor" opacity="0.15"/>
                <path d="M12 28V12h6a6 6 0 0 1 0 12h-2m-4-6h4a2 2 0 0 1 0 4h-4v-4Z" fill="currentColor"/>
                <circle cx="28" cy="20" r="6" fill="currentColor" opacity="0.3"/>
              </svg>
            </div>
            <h2 class="login-page__card-title">欢迎回来</h2>
            <p class="login-page__card-subtitle">请登录您的管理账号</p>
          </div>

          <el-alert
            v-if="errorMessage"
            :title="errorMessage"
            type="error"
            :closable="false"
            show-icon
            class="login-page__alert"
          />

          <el-form
            class="login-page__form"
            label-position="top"
            @submit.prevent="submit"
          >
            <el-form-item label="账号">
              <el-input
                v-model="form.username"
                placeholder="用户名 / 邮箱"
                :prefix-icon="User"
                size="large"
              />
            </el-form-item>

            <el-form-item label="密码">
              <el-input
                v-model="form.password"
                placeholder="请输入密码"
                show-password
                type="password"
                :prefix-icon="Lock"
                size="large"
              />
            </el-form-item>

            <el-form-item label="验证码">
              <div class="login-page__captcha">
                <el-input
                  v-model="form.captchaCode"
                  maxlength="8"
                  placeholder="输入验证码"
                  :prefix-icon="Key"
                  size="large"
                />
                <button
                  class="login-page__captcha-btn"
                  type="button"
                  :disabled="captchaLoading"
                  @click="loadCaptcha"
                >
                  <img
                    v-if="captcha.image"
                    :src="captcha.image"
                    alt="验证码"
                    class="login-page__captcha-img"
                  />
                  <span v-else class="login-page__captcha-placeholder">加载中...</span>
                </button>
              </div>
            </el-form-item>

            <el-button
              class="login-page__submit"
              :disabled="!canSubmit"
              :loading="loading"
              native-type="submit"
              round
              size="large"
              type="primary"
            >
              登 录
            </el-button>
          </el-form>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
/* ===== 页面容器 ===== */
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: #0b1120;
  padding: 24px;
}

/* ===== 装饰光晕 ===== */
.login-page__glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(120px);
  pointer-events: none;
}
.login-page__glow--top {
  width: 600px;
  height: 600px;
  top: -200px;
  left: -100px;
  background: radial-gradient(circle, rgba(14, 165, 233, 0.25), transparent 70%);
}
.login-page__glow--bottom {
  width: 500px;
  height: 500px;
  bottom: -150px;
  right: -80px;
  background: radial-gradient(circle, rgba(168, 85, 247, 0.2), transparent 70%);
}
.login-page__glow--side {
  width: 400px;
  height: 400px;
  top: 50%;
  right: -100px;
  transform: translateY(-50%);
  background: radial-gradient(circle, rgba(14, 165, 233, 0.12), transparent 70%);
}

/* ===== 背景网格 ===== */
.login-page__grid {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image:
    linear-gradient(to right, rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
  background-size: 80px 80px;
  mask-image: radial-gradient(ellipse at center, rgba(0, 0, 0, 0.6), transparent 75%);
  -webkit-mask-image: radial-gradient(ellipse at center, rgba(0, 0, 0, 0.6), transparent 75%);
}

/* ===== 主容器 ===== */
.login-page__container {
  position: relative;
  z-index: 1;
  width: min(1100px, 100%);
  display: grid;
  grid-template-columns: 1fr 420px;
  gap: 40px;
  align-items: center;
}

/* ===== 左侧品牌区域 ===== */
.login-page__brand {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 520px;
  padding: 0 0 0 8px;
}

.login-page__brand-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.login-page__logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.login-page__logo-icon {
  width: 36px;
  height: 36px;
  color: #0ea5e9;
}

.login-page__logo-text {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #f1f5f9;
  font-family: var(--pc-display-font, 'Inter', sans-serif);
}

.login-page__title {
  margin: 0;
  font-size: clamp(36px, 5vw, 52px);
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.03em;
  color: #f8fafc;
  font-family: var(--pc-display-font, 'Inter', sans-serif);
}

.login-page__desc {
  margin: 0;
  font-size: 16px;
  line-height: 1.7;
  color: rgba(148, 163, 184, 0.85);
  max-width: 480px;
}

.login-page__features {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 8px;
}

.login-page__feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  color: rgba(226, 232, 240, 0.8);
}

.login-page__feature-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #0ea5e9;
  box-shadow: 0 0 8px rgba(14, 165, 233, 0.5);
  flex-shrink: 0;
}

.login-page__footer-text {
  font-size: 13px;
  color: rgba(100, 116, 139, 0.6);
}

/* ===== 右侧登录卡片 ===== */
.login-page__card {
  background: rgba(255, 255, 255, 0.04);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-radius: 24px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.04),
    0 24px 80px rgba(0, 0, 0, 0.4);
}

.login-page__card-inner {
  padding: 40px 32px;
}

.login-page__card-header {
  text-align: center;
  margin-bottom: 28px;
}

.login-page__card-avatar {
  width: 56px;
  height: 56px;
  margin: 0 auto 16px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0ea5e9;
  background: rgba(14, 165, 233, 0.1);
  border: 1px solid rgba(14, 165, 233, 0.15);
}

.login-page__card-title {
  margin: 0 0 8px;
  font-size: 22px;
  font-weight: 700;
  color: #f1f5f9;
  font-family: var(--pc-display-font, 'Inter', sans-serif);
}

.login-page__card-subtitle {
  margin: 0;
  font-size: 14px;
  color: rgba(148, 163, 184, 0.8);
}

/* ===== 表单错误提示 ===== */
.login-page__alert {
  margin-bottom: 20px;
  border-radius: 12px;
}

/* ===== 表单 ===== */
.login-page__form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.login-page__form :deep(.el-form-item__label) {
  color: rgba(203, 213, 225, 0.9);
  font-weight: 600;
  font-size: 13px;
  padding-bottom: 6px;
}

.login-page__form :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  box-shadow: none;
  padding: 4px 12px;
  transition: border-color 0.2s, background 0.2s;
}

.login-page__form :deep(.el-input__wrapper:hover) {
  border-color: rgba(255, 255, 255, 0.15);
}

.login-page__form :deep(.el-input__wrapper.is-focus) {
  border-color: #0ea5e9;
  background: rgba(14, 165, 233, 0.06);
  box-shadow: 0 0 0 3px rgba(14, 165, 233, 0.1);
}

.login-page__form :deep(.el-input__inner) {
  color: #f1f5f9;
  height: 46px;
}

.login-page__form :deep(.el-input__inner::placeholder) {
  color: rgba(100, 116, 139, 0.5);
}

.login-page__form :deep(.el-input__prefix) {
  color: rgba(100, 116, 139, 0.6);
  margin-right: 8px;
}

/* ===== 验证码区域 ===== */
.login-page__captcha {
  display: grid;
  grid-template-columns: 1fr 140px;
  gap: 12px;
  align-items: stretch;
}

.login-page__captcha-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  cursor: pointer;
  padding: 0;
  overflow: hidden;
  transition: border-color 0.2s, background 0.2s;
  min-height: 46px;
}

.login-page__captcha-btn:hover {
  border-color: rgba(255, 255, 255, 0.15);
  background: rgba(255, 255, 255, 0.06);
}

.login-page__captcha-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.login-page__captcha-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.login-page__captcha-placeholder {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.6);
}

/* ===== 提交按钮 ===== */
.login-page__submit {
  width: 100%;
  height: 48px;
  margin-top: 4px;
  border-radius: 12px;
  font-weight: 700;
  font-size: 15px;
  letter-spacing: 0.04em;
  background: linear-gradient(135deg, #0ea5e9, #2563eb);
  border: none;
  color: #fff;
  transition: opacity 0.2s, transform 0.15s;
}

.login-page__submit:hover {
  opacity: 0.92;
  transform: translateY(-1px);
}

.login-page__submit:active {
  transform: translateY(0);
}

.login-page__submit.is-disabled {
  background: rgba(255, 255, 255, 0.06);
  color: rgba(100, 116, 139, 0.4);
}

/* ===== 响应式 ===== */
@media (max-width: 1024px) {
  .login-page__container {
    grid-template-columns: 1fr;
    max-width: 460px;
  }

  .login-page__brand {
    display: none;
  }
}

@media (max-width: 480px) {
  .login-page {
    padding: 12px;
  }

  .login-page__card-inner {
    padding: 28px 20px;
  }

  .login-page__captcha {
    grid-template-columns: 1fr 120px;
  }
}
</style>
