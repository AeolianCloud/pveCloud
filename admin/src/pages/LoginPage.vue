<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  AlertCircle,
  Eye,
  EyeOff,
  FileText,
  KeyRound,
  LockKeyhole,
  MonitorCheck,
  RefreshCw,
  ServerCog,
  ShieldCheck,
  UserRound,
  X,
} from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getAdminLoginCaptcha } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
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
const showPassword = ref(false)
const captcha = reactive({
  id: '',
  image: '',
})

const featureItems = [
  {
    icon: MonitorCheck,
    title: '资源集中管控',
    description: '统一管理节点、套餐、订单与实例运行状态',
  },
  {
    icon: FileText,
    title: '业务流程清晰',
    description: '订单、客户、工单、财务数据集中呈现',
  },
  {
    icon: ShieldCheck,
    title: '安全权限隔离',
    description: '验证码登录与角色权限共同守住后台入口',
  },
]

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
    await auth.login({
      username: form.username.trim(),
      password: form.password,
      captcha_id: captcha.id,
      captcha_code: form.captchaCode.trim(),
    })
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard'
    await router.replace(redirect)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败，请检查账号、密码或验证码'
    await loadCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(loadCaptcha)
</script>

<template>
  <main class="pc-login-page">
    <div class="pc-login-bg pc-login-bg-a" aria-hidden="true"></div>
    <div class="pc-login-bg pc-login-bg-b" aria-hidden="true"></div>
    <div class="pc-login-bg pc-login-bg-c" aria-hidden="true"></div>

    <header class="pc-login-logo" aria-label="pveCloud">
      <span class="pc-login-logo-mark">
        <ServerCog :size="26" aria-hidden="true" />
      </span>
      <span>pveCloud</span>
    </header>

    <section class="pc-login-intro" aria-label="pveCloud 后台管理系统">
      <div>
        <span class="pc-login-eyebrow">pveCloud Admin</span>
        <h1><strong>pveCloud</strong> 后台管理系统</h1>
        <p>面向 IDC 云服务器销售业务的统一运营控制台，集中处理产品、订单、客户和资源。</p>
      </div>

      <div class="pc-login-features">
        <article v-for="item in featureItems" :key="item.title" class="pc-login-feature">
          <span>
            <component :is="item.icon" :size="28" aria-hidden="true" />
          </span>
          <div>
            <strong>{{ item.title }}</strong>
            <small>{{ item.description }}</small>
          </div>
        </article>
      </div>
    </section>

    <Transition name="pc-login-toast">
      <div v-if="errorMessage" class="pc-login-toast" role="alert">
        <AlertCircle :size="20" aria-hidden="true" />
        <span>{{ errorMessage }}</span>
        <button type="button" aria-label="关闭错误提示" title="关闭错误提示" @click="errorMessage = ''">
          <X :size="16" aria-hidden="true" />
        </button>
      </div>
    </Transition>

    <section class="pc-login-card" aria-label="后台登录">
      <div class="pc-login-heading">
        <span>管理员登录</span>
        <h2>欢迎回来</h2>
        <p>请输入账号信息和验证码，进入 pveCloud 管理后台。</p>
      </div>

      <form class="pc-login-form" @submit.prevent="submit">
        <label>
          <span>管理员账号</span>
          <div class="pc-login-input">
            <UserRound :size="21" aria-hidden="true" />
            <input
              v-model="form.username"
              autocomplete="username"
              name="username"
              placeholder="请输入用户名 / 邮箱 / 手机号"
              type="text"
            />
          </div>
        </label>

        <label>
          <span>登录密码</span>
          <div class="pc-login-input">
            <KeyRound :size="21" aria-hidden="true" />
            <input
              v-model="form.password"
              autocomplete="current-password"
              name="password"
              placeholder="请输入登录密码"
              :type="showPassword ? 'text' : 'password'"
            />
            <button
              class="pc-login-eye"
              type="button"
              :aria-label="showPassword ? '隐藏密码' : '显示密码'"
              :title="showPassword ? '隐藏密码' : '显示密码'"
              @click="showPassword = !showPassword"
            >
              <EyeOff v-if="showPassword" :size="20" aria-hidden="true" />
              <Eye v-else :size="20" aria-hidden="true" />
            </button>
          </div>
        </label>

        <label>
          <span>安全验证</span>
          <div class="pc-login-captcha-row">
            <div class="pc-login-input">
              <ShieldCheck :size="21" aria-hidden="true" />
              <input
                v-model="form.captchaCode"
                autocomplete="off"
                maxlength="8"
                name="captcha_code"
                placeholder="请输入验证码"
                type="text"
              />
            </div>
            <button
              class="pc-login-captcha"
              type="button"
              :disabled="captchaLoading"
              aria-label="刷新验证码"
              title="刷新验证码"
              @click="loadCaptcha"
            >
              <img v-if="captcha.image" :src="captcha.image" alt="登录验证码" />
              <span v-else>{{ captchaLoading ? '加载中' : '刷新' }}</span>
              <RefreshCw :size="16" aria-hidden="true" />
            </button>
          </div>
        </label>

        <button class="pc-login-submit" type="submit" :disabled="!canSubmit || loading">
          <LockKeyhole :size="21" aria-hidden="true" />
          <span>{{ loading ? '登录中...' : '安全登录' }}</span>
        </button>
      </form>

      <div class="pc-login-note">
        <ShieldCheck :size="19" aria-hidden="true" />
        <span>登录会话受后台安全策略保护</span>
      </div>
    </section>
  </main>
</template>

<style scoped>
.pc-login-page {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background: #f5f8ff;
  color: #001845;
  font-family:
    "Microsoft YaHei", "PingFang SC", Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif;
}

.pc-login-page *,
.pc-login-page *::before,
.pc-login-page *::after {
  box-sizing: border-box;
}

.pc-login-page button,
.pc-login-page input {
  font: inherit;
}

.pc-login-page button {
  border: 0;
}

.pc-login-bg {
  position: absolute;
  pointer-events: none;
}

.pc-login-bg-a {
  top: -160px;
  left: -70px;
  width: 640px;
  height: 980px;
  background: #dbe6ff;
  clip-path: polygon(0 0, 100% 0, 22% 100%, 0 100%);
}

.pc-login-bg-b {
  bottom: -230px;
  left: -80px;
  width: 700px;
  height: 620px;
  background: linear-gradient(42deg, rgba(83, 131, 245, 0.25), rgba(83, 131, 245, 0.03));
  clip-path: polygon(0 0, 100% 55%, 100% 100%, 0 100%);
}

.pc-login-bg-c {
  left: 0;
  bottom: 0;
  width: 610px;
  height: 380px;
  background: rgba(82, 126, 244, 0.1);
  clip-path: polygon(0 26%, 100% 100%, 0 100%);
}

.pc-login-logo {
  position: absolute;
  top: 36px;
  left: 275px;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 14px;
  color: #001845;
  font-size: 29px;
  font-weight: 850;
}

.pc-login-logo-mark {
  width: 44px;
  height: 44px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  color: #ffffff;
  background: linear-gradient(135deg, #3978ff, #67c4ff);
  box-shadow: 0 18px 34px rgba(47, 104, 245, 0.28);
}

.pc-login-intro {
  position: absolute;
  left: 350px;
  top: 298px;
  z-index: 2;
  width: 520px;
}

.pc-login-eyebrow {
  color: #2f68f5;
  font-size: 15px;
  font-weight: 850;
}

.pc-login-intro h1 {
  margin: 14px 0 0;
  color: #001845;
  font-size: 50px;
  line-height: 1.15;
  font-weight: 850;
  letter-spacing: 0;
}

.pc-login-intro h1 strong {
  color: #2f68f5;
}

.pc-login-intro p {
  margin: 14px 0 0;
  color: #506489;
  font-size: 22px;
  line-height: 1.55;
}

.pc-login-features {
  display: grid;
  gap: 28px;
  margin-top: 58px;
}

.pc-login-feature {
  display: grid;
  grid-template-columns: 68px minmax(0, 1fr);
  align-items: center;
  gap: 26px;
}

.pc-login-feature > span {
  width: 68px;
  height: 68px;
  display: grid;
  place-items: center;
  border: 1px solid #b9cdfb;
  border-radius: 8px;
  color: #2f68f5;
  background: rgba(239, 244, 255, 0.72);
}

.pc-login-feature strong {
  display: block;
  color: #001845;
  font-size: 25px;
  line-height: 1.2;
}

.pc-login-feature small {
  display: block;
  margin-top: 8px;
  color: #526489;
  font-size: 17px;
  line-height: 1.5;
}

.pc-login-card {
  position: absolute;
  top: 130px;
  right: 350px;
  z-index: 2;
  width: 620px;
  padding: 48px 50px 46px;
  border: 1px solid #d3def0;
  border-radius: 16px;
  background: #ffffff;
  box-shadow: 0 28px 78px rgba(19, 40, 87, 0.12);
}

.pc-login-heading span {
  color: #2f68f5;
  font-size: 18px;
  font-weight: 850;
}

.pc-login-heading h2 {
  margin: 18px 0 0;
  color: #001845;
  font-size: 44px;
  line-height: 1.05;
  font-weight: 900;
  letter-spacing: 0;
}

.pc-login-heading p {
  margin: 14px 0 0;
  color: #4f6289;
  font-size: 17px;
}

.pc-login-form {
  display: grid;
  gap: 22px;
  margin-top: 34px;
}

.pc-login-form label {
  display: grid;
  gap: 11px;
  color: #001845;
  font-size: 14px;
  font-weight: 850;
}

.pc-login-input {
  height: 58px;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 0 18px;
  border: 1px solid #d5e0f1;
  border-radius: 8px;
  color: #68799e;
  background: #ffffff;
  transition:
    border-color 160ms ease,
    box-shadow 160ms ease;
}

.pc-login-input:focus-within {
  border-color: #2f68f5;
  box-shadow: 0 0 0 4px rgba(47, 104, 245, 0.1);
}

.pc-login-input input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  color: #001845;
  background: transparent;
  font-size: 15px;
  font-weight: 650;
}

.pc-login-input input::placeholder {
  color: #64769b;
}

.pc-login-eye {
  width: 34px;
  height: 34px;
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 7px;
  color: #68799e;
  background: transparent;
  cursor: pointer;
}

.pc-login-eye:hover,
.pc-login-eye:focus-visible {
  color: #2f68f5;
  background: rgba(47, 104, 245, 0.08);
  outline: 0;
}

.pc-login-captcha-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 148px;
  gap: 12px;
}

.pc-login-captcha {
  height: 58px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 18px;
  align-items: center;
  gap: 8px;
  padding: 0 10px;
  border: 1px solid #d5e0f1;
  border-radius: 8px;
  color: #2f68f5;
  background: #f2f6ff;
  cursor: pointer;
}

.pc-login-captcha:disabled {
  cursor: wait;
}

.pc-login-captcha img {
  width: 100%;
  height: 42px;
  display: block;
  border-radius: 6px;
  object-fit: cover;
}

.pc-login-captcha span {
  color: #2f68f5;
  font-size: 14px;
  font-weight: 850;
}

.pc-login-submit {
  min-height: 61px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-top: 8px;
  border-radius: 8px;
  color: #ffffff;
  background: #4f73ce;
  box-shadow: 0 12px 24px rgba(47, 104, 245, 0.2);
  cursor: pointer;
  font-size: 22px;
  font-weight: 850;
}

.pc-login-submit:hover:not(:disabled) {
  background: #3f65c3;
}

.pc-login-submit:disabled {
  cursor: not-allowed;
  opacity: 0.56;
}

.pc-login-note {
  min-height: 59px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 22px;
  border-radius: 8px;
  color: #385286;
  background: #f1f5fd;
  font-size: 17px;
  font-weight: 750;
}

.pc-login-note svg {
  color: #2f68f5;
}

.pc-login-toast {
  position: fixed;
  top: 26px;
  left: 50%;
  z-index: 20;
  width: min(520px, calc(100vw - 32px));
  min-height: 48px;
  display: grid;
  grid-template-columns: 20px minmax(0, 1fr) 28px;
  align-items: center;
  gap: 12px;
  padding: 12px 12px 12px 16px;
  border: 1px solid #ffd3ce;
  border-radius: 8px;
  color: #b42318;
  background: #fff2f0;
  box-shadow: 0 18px 46px rgba(27, 43, 85, 0.16);
  transform: translateX(-50%);
}

.pc-login-toast span {
  min-width: 0;
  font-size: 14px;
}

.pc-login-toast button {
  width: 28px;
  height: 28px;
  display: grid;
  place-items: center;
  padding: 0;
  border-radius: 6px;
  color: #b42318;
  background: transparent;
  cursor: pointer;
}

.pc-login-toast-enter-active,
.pc-login-toast-leave-active {
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.pc-login-toast-enter-from,
.pc-login-toast-leave-to {
  opacity: 0;
  transform: translate(-50%, -10px);
}

@media (max-width: 1460px) {
  .pc-login-logo {
    left: 110px;
  }

  .pc-login-intro {
    left: 150px;
  }

  .pc-login-card {
    right: 120px;
  }
}

@media (max-width: 1120px) {
  .pc-login-page {
    min-height: auto;
    display: grid;
    gap: 28px;
    padding: 28px 22px 44px;
  }

  .pc-login-bg-a {
    width: 420px;
  }

  .pc-login-logo,
  .pc-login-intro,
  .pc-login-card {
    position: relative;
    inset: auto;
    width: min(100%, 680px);
    margin: 0 auto;
  }

  .pc-login-intro {
    margin-top: 54px;
  }

  .pc-login-card {
    padding: 34px;
  }
}

@media (max-width: 720px) {
  .pc-login-intro h1 {
    font-size: 38px;
  }

  .pc-login-intro p {
    font-size: 18px;
  }

  .pc-login-feature {
    grid-template-columns: 54px minmax(0, 1fr);
  }

  .pc-login-feature > span {
    width: 54px;
    height: 54px;
  }

  .pc-login-captcha-row {
    grid-template-columns: 1fr;
  }

  .pc-login-captcha {
    width: 170px;
  }
}
</style>
