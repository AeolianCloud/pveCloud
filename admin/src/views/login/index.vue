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
.login-page {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  padding: 32px;
}

.login-page__halo {
  position: absolute;
  width: 360px;
  aspect-ratio: 1;
  border-radius: 999px;
  filter: blur(18px);
  opacity: 0.52;
}

.login-page__halo--left {
  top: -40px;
  left: -60px;
  background: rgba(14, 165, 233, 0.18);
}

.login-page__halo--right {
  right: -20px;
  bottom: 0;
  background: rgba(245, 158, 11, 0.16);
}

.login-page__shell {
  position: relative;
  z-index: 1;
  width: min(1220px, 100%);
  min-height: calc(100vh - 64px);
  margin: 0 auto;
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(420px, 0.92fr);
  gap: 26px;
  align-items: stretch;
}

.login-page__hero,
.login-page__card {
  border-radius: 30px;
  box-shadow: var(--pc-card-shadow);
  backdrop-filter: blur(18px);
}

.login-page__hero {
  display: grid;
  align-content: space-between;
  gap: 28px;
  padding: 34px;
  color: var(--pc-brand-text);
  background:
    radial-gradient(circle at top right, rgba(245, 158, 11, 0.24), transparent 24%),
    linear-gradient(160deg, #07111d 0%, #0b2740 58%, #0e4f76 100%);
}

.login-page__eyebrow,
.login-page__card-label {
  width: fit-content;
  padding: 7px 12px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.login-page__eyebrow {
  color: rgba(240, 249, 255, 0.92);
  background: rgba(255, 255, 255, 0.1);
}

.login-page__hero h1 {
  margin: 0;
  font-family: var(--pc-display-font);
  font-size: clamp(42px, 6vw, 72px);
  line-height: 0.94;
}

.login-page__hero p {
  max-width: 620px;
  margin: 0;
  color: rgba(226, 232, 240, 0.82);
  line-height: 1.7;
  font-size: 16px;
}

.login-page__highlight-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.login-page__highlight-card {
  padding: 18px;
  border-radius: 22px;
  display: grid;
  gap: 12px;
  background: rgba(255, 255, 255, 0.08);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.08);
}

.login-page__highlight-icon {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.12);
}

.login-page__highlight-card strong {
  line-height: 1.5;
}

.login-page__meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.login-page__meta div {
  padding: 18px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.08);
}

.login-page__meta span {
  display: block;
  margin-bottom: 8px;
  color: rgba(191, 219, 254, 0.72);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.login-page__meta strong {
  font-size: 15px;
}

.login-page__card {
  align-self: center;
  border: none;
  background: rgba(255, 255, 255, 0.9);
}

.login-page__card :deep(.el-card__header) {
  padding-bottom: 10px;
  border-bottom-color: rgba(148, 163, 184, 0.16);
}

.login-page__card :deep(.el-card__body) {
  padding-top: 20px;
}

.login-page__card-header strong {
  display: block;
  margin-top: 10px;
  color: var(--pc-title-text);
  font-family: var(--pc-display-font);
  font-size: 28px;
}

.login-page__card-header p {
  margin: 8px 0 0;
  color: var(--pc-muted-text);
  line-height: 1.65;
}

.login-page__card-label {
  color: var(--pc-accent-strong);
  background: rgba(14, 165, 233, 0.12);
}

.login-page__alert,
.login-page__form {
  margin-top: 2px;
}

.login-page__form :deep(.el-input__wrapper) {
  min-height: 50px;
  border-radius: 16px;
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.14);
}

.login-page__captcha {
  width: 100%;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 168px;
  gap: 12px;
}

.login-page__captcha-button {
  justify-content: space-between;
  border-radius: 16px;
}

.login-page__captcha-button img {
  width: 100%;
  height: 36px;
  object-fit: cover;
}

.login-page__captcha-icon {
  margin-left: 6px;
}

.login-page__submit {
  width: 100%;
  min-height: 52px;
  margin-top: 8px;
  border-radius: 16px;
  font-weight: 700;
}

@media (max-width: 1080px) {
  .login-page__shell {
    grid-template-columns: 1fr;
  }

  .login-page__card {
    align-self: stretch;
  }
}

@media (max-width: 767px) {
  .login-page {
    padding: 16px;
  }

  .login-page__shell,
  .login-page__hero {
    min-height: auto;
  }

  .login-page__hero,
  .login-page__card {
    padding-left: 0;
    padding-right: 0;
  }

  .login-page__hero {
    padding: 24px;
  }

  .login-page__highlight-grid,
  .login-page__meta {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .login-page__captcha {
    grid-template-columns: 1fr;
  }
}
</style>
