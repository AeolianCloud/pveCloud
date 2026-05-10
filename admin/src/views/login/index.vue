<script setup lang="ts">
import { KeyOutline, LockClosedOutline, PersonOutline } from '@vicons/ionicons5'
import { NButton, NForm, NFormItem, NIcon, NInput, useNotification } from 'naive-ui'
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getAdminLoginCaptcha } from '../../api/auth'
import { ADMIN_ROUTE_PATH, normalizeAdminRedirect } from '../../router/constants'
import { useAuthStore } from '../../store/modules/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const notification = useNotification()

const form = reactive({
  username: '',
  password: '',
  captchaCode: '',
})

const loading = ref(false)
const captchaLoading = ref(false)
const captcha = reactive({
  id: '',
  image: '',
})

const touched = reactive({ username: false, password: false, captchaCode: false })

const canSubmit = computed(
  () =>
    form.username.trim().length > 0 &&
    form.password.length >= 6 &&
    captcha.id.length > 0 &&
    form.captchaCode.trim().length >= 4,
)

const usernameError = computed(() => {
  if (!touched.username) return ''
  if (form.username.trim().length === 0) return '请输入用户名'
  return ''
})

const passwordError = computed(() => {
  if (!touched.password) return ''
  if (form.password.length === 0) return '请输入密码'
  if (form.password.length < 6) return `还需 ${6 - form.password.length} 位`
  return ''
})

const captchaError = computed(() => {
  if (!touched.captchaCode) return ''
  if (form.captchaCode.trim().length === 0) return '请输入验证码'
  if (form.captchaCode.trim().length < 4) return `还需 ${4 - form.captchaCode.trim().length} 位`
  return ''
})

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
    notification.error({
      title: '验证码加载失败',
      content: error instanceof Error ? error.message : '请稍后重试',
      duration: 4000,
    })
  } finally {
    captchaLoading.value = false
  }
}

async function submit() {
  touched.username = true
  touched.password = true
  touched.captchaCode = true
  if (!canSubmit.value || loading.value) return

  loading.value = true

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
    notification.error({
      title: '登录失败',
      content: error instanceof Error ? error.message : '请检查账号、密码或验证码',
      duration: 4000,
    })
    await loadCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  document.body.classList.add('login-active')
  void loadCaptcha()
})

onUnmounted(() => {
  document.body.classList.remove('login-active')
})
</script>

<template>
  <main class="login-page">
    <div class="login-page__glow login-page__glow--top" />
    <div class="login-page__glow login-page__glow--bottom" />
    <div class="login-page__glow login-page__glow--side" />
    <div class="login-page__grid" />

    <div class="login-page__container">
      <div class="login-page__brand">
        <div class="login-page__brand-content">
          <div class="login-page__logo">
            <svg class="login-page__logo-icon" viewBox="0 0 40 40" fill="none">
              <rect width="40" height="40" rx="10" fill="currentColor" opacity="0.2" />
              <path d="M12 28V12h6a6 6 0 0 1 0 12h-2m-4-6h4a2 2 0 0 1 0 4h-4v-4Z" fill="currentColor" />
              <circle cx="28" cy="20" r="6" fill="currentColor" opacity="0.4" />
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
        <div class="login-page__footer-text">&copy; 2024 PVE Cloud. All rights reserved.</div>
      </div>

      <div class="login-page__card">
        <div class="login-page__card-inner">
          <div class="login-page__card-header">
            <div class="login-page__card-avatar">
              <svg viewBox="0 0 40 40" fill="none">
                <rect width="40" height="40" rx="10" fill="currentColor" opacity="0.15" />
                <path d="M12 28V12h6a6 6 0 0 1 0 12h-2m-4-6h4a2 2 0 0 1 0 4h-4v-4Z" fill="currentColor" />
                <circle cx="28" cy="20" r="6" fill="currentColor" opacity="0.3" />
              </svg>
            </div>
            <h2 class="login-page__card-title">欢迎回来</h2>
            <p class="login-page__card-subtitle">请登录您的管理账号</p>
          </div>

          <NForm class="login-page__form" label-placement="top" @submit.prevent="submit">
            <NFormItem label="账号" :feedback="usernameError" :validation-status="usernameError ? 'error' : undefined">
              <NInput
                v-model:value="form.username"
                placeholder="用户名 / 邮箱"
                size="large"
                @blur="touched.username = true"
              >
                <template #prefix>
                  <NIcon><PersonOutline /></NIcon>
                </template>
              </NInput>
            </NFormItem>

            <NFormItem label="密码" :feedback="passwordError" :validation-status="passwordError ? 'error' : undefined">
              <NInput
                v-model:value="form.password"
                placeholder="请输入密码"
                type="password"
                show-password-on="click"
                size="large"
                @blur="touched.password = true"
              >
                <template #prefix>
                  <NIcon><LockClosedOutline /></NIcon>
                </template>
              </NInput>
            </NFormItem>

            <NFormItem label="验证码" :feedback="captchaError" :validation-status="captchaError ? 'error' : undefined">
              <div class="login-page__captcha">
                <NInput
                  v-model:value="form.captchaCode"
                  :maxlength="8"
                  placeholder="输入验证码"
                  size="large"
                  @blur="touched.captchaCode = true"
                >
                  <template #prefix>
                    <NIcon><KeyOutline /></NIcon>
                  </template>
                </NInput>
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
            </NFormItem>

            <NButton
              class="login-page__submit"
              :disabled="!canSubmit"
              :loading="loading"
              attr-type="submit"
              round
              size="large"
              type="primary"
              block
              @click="submit"
            >
              登 录
            </NButton>
          </NForm>
        </div>
      </div>
    </div>
  </main>
</template>

<style>
body.login-active {
  background: #0b1120 !important;
}
</style>

<style scoped>
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

.login-page__container {
  position: relative;
  z-index: 1;
  width: min(1100px, 100%);
  display: grid;
  grid-template-columns: 1fr 420px;
  gap: 40px;
  align-items: center;
}

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
}

.login-page__title {
  margin: 0;
  font-size: clamp(36px, 5vw, 52px);
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.03em;
  color: #f8fafc;
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
}

.login-page__card-subtitle {
  margin: 0;
  font-size: 14px;
  color: rgba(148, 163, 184, 0.8);
}

.login-page__form :deep(.n-form-item-label__text) {
  color: rgba(203, 213, 225, 0.9);
  font-weight: 600;
  font-size: 13px;
}

.login-page__form :deep(.n-input) {
  background: rgba(255, 255, 255, 0.04);
  border-radius: 12px;
}

.login-page__form :deep(.n-input .n-input__input-el),
.login-page__form :deep(.n-input .n-input__textarea-el) {
  color: #f1f5f9;
}

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

.login-page__submit {
  width: 100%;
  margin-top: 8px;
}

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
