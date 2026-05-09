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
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMessage = ref('')

const { captchaCode, captchaError, captchaId, captchaImage, captchaLoading, captchaReady, refreshCaptcha } =
  useAuthCaptcha(registerCaptchaEnabled, getRegisterCaptcha)

const canSubmit = computed(() => (
  siteConfigLoaded.value &&
  username.value.trim().length >= 3 &&
  email.value.trim() !== '' &&
  password.value.length >= 6 &&
  password.value === confirmPassword.value &&
  (!registerCaptchaEnabled.value || captchaCode.value.trim().length >= 4) &&
  captchaReady.value &&
  !loading.value
))

const submitHint = computed(() => {
  if (siteConfigLoading.value && !siteConfigLoaded.value) return '正在同步注册配置...'
  if (username.value.trim().length > 0 && username.value.trim().length < 3) return '用户名至少需要 3 个字符'
  if (password.value.length > 0 && password.value.length < 6) return '密码至少需要 6 位'
  if (confirmPassword.value && password.value !== confirmPassword.value) return '两次输入的密码不一致'
  if (registerCaptchaEnabled.value && !captchaReady.value) return captchaError.value || '验证码准备中...'
  if (siteConfigError.value) return siteConfigError.value
  return ''
})

function errorText(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (typeof error === 'object' && error !== null && 'request' in error) return '网络连接失败，请检查后重试'
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
      captcha_id: registerCaptchaEnabled.value ? captchaId.value : undefined,
      captcha_code: registerCaptchaEnabled.value ? captchaCode.value.trim() : undefined,
    })
    await router.replace('/user')
  } catch (error) {
    errorMessage.value = errorText(error)
    if (registerCaptchaEnabled.value) void refreshCaptcha()
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
        <p class="section-label">Register</p>
        <h1>创建用户端账号</h1>
        <p>注册成功后直接进入用户中心。显示名称可在账号资料页维护。</p>
      </aside>

      <form class="auth-card" @submit.prevent="handleRegister">
        <div class="auth-heading">
          <h2>注册新账户</h2>
          <p>填写用户名、邮箱和登录密码。</p>
        </div>

        <div class="field-grid">
          <label class="field"><span>用户名</span><span class="field-control"><input v-model="username" type="text" autocomplete="username" placeholder="至少 3 个字符" /></span></label>
          <label class="field"><span>邮箱</span><span class="field-control"><input v-model="email" type="email" autocomplete="email" placeholder="name@example.com" /></span></label>
        </div>

        <div class="field-grid">
          <label class="field"><span>密码</span><span class="field-control"><input v-model="password" type="password" autocomplete="new-password" placeholder="至少 6 位" /></span></label>
          <label class="field"><span>确认密码</span><span class="field-control"><input v-model="confirmPassword" type="password" autocomplete="new-password" placeholder="重复密码" /></span></label>
        </div>

        <div v-if="registerCaptchaEnabled" class="field">
          <span>安全验证码</span>
          <div class="captcha-row">
            <span class="field-control"><input v-model="captchaCode" type="text" maxlength="8" autocomplete="off" placeholder="图中字符" /></span>
            <button class="captcha-box" type="button" :disabled="captchaLoading" @click="refreshCaptcha">
              <img v-if="captchaImage" :src="captchaImage" alt="注册验证码" />
              <span v-else>{{ captchaLoading ? '加载中...' : '刷新' }}</span>
            </button>
          </div>
        </div>

        <p v-if="errorMessage || captchaError || submitHint" class="notice error">{{ errorMessage || captchaError || submitHint }}</p>

        <button class="btn btn-primary btn-block" type="submit" :disabled="!canSubmit">
          <span v-if="loading" class="spinner-small"></span>
          {{ loading ? '创建中...' : '注册并进入控制台' }}
        </button>

        <p class="switch-line">已有账号？<RouterLink to="/login">返回登录</RouterLink></p>
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
  width: min(1080px, 100%);
  display: grid;
  grid-template-columns: minmax(280px, 0.86fr) minmax(420px, 1.14fr);
  overflow: hidden;
}
.auth-aside {
  display: grid;
  align-content: end;
  gap: 16px;
  min-height: 560px;
  padding: clamp(28px, 5vw, 48px);
  color: #fff;
  background: linear-gradient(135deg, rgba(15, 118, 110, 0.95), rgba(15, 23, 42, 0.92));
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
.auth-heading h2 {
  font-size: 2rem;
  letter-spacing: -0.05em;
}
.auth-heading p,
.switch-line {
  margin-top: 8px;
  color: var(--c-text-2);
}
.field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
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
.switch-line a {
  color: var(--c-primary);
  font-weight: 800;
}
@media (max-width: 860px) {
  .auth-shell,
  .field-grid {
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
