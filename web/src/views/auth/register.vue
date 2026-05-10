<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getRegisterCaptcha } from '../../api/auth'
import { getApiErrorMessage } from '../../api/request'
import { getSiteConfig } from '../../api/site-config'
import { useAuthStore } from '../../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const captchaEnabled = ref(false)
const captchaId = ref('')
const captchaCode = ref('')
const captchaImage = ref('')
const captchaLoading = ref(false)

const refreshCaptcha = async () => {
  if (!captchaEnabled.value) {
    captchaId.value = ''
    captchaCode.value = ''
    captchaImage.value = ''
    return
  }

  captchaLoading.value = true
  try {
    const captcha = await getRegisterCaptcha()
    captchaId.value = captcha.captcha_id
    captchaImage.value = captcha.image
    captchaCode.value = ''
  } catch (err) {
    error.value = getApiErrorMessage(err, '验证码加载失败，请稍后重试')
  } finally {
    captchaLoading.value = false
  }
}

onMounted(async () => {
  try {
    const config = await getSiteConfig()
    captchaEnabled.value = config.register_captcha_enabled
  } catch {
    captchaEnabled.value = false
  }

  await refreshCaptcha()
})

const handleRegister = async () => {
  if (!username.value || !email.value || !password.value || !confirmPassword.value) {
    error.value = '请填写所有字段'
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = '密码不一致'
    return
  }
  if (captchaEnabled.value && (!captchaId.value || !captchaCode.value)) {
    error.value = '请输入验证码'
    return
  }

  loading.value = true
  error.value = ''
  try {
    await authStore.registerAccount({
      username: username.value,
      email: email.value,
      password: password.value,
      ...(captchaEnabled.value ? { captcha_id: captchaId.value, captcha_code: captchaCode.value } : {}),
    })
    await router.push('/user')
  } catch (err) {
    error.value = getApiErrorMessage(err, '注册失败')
    await refreshCaptcha()
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex justify-center bg-white px-4 pt-8 pb-4 sm:pt-10 lg:pt-12">
    <div class="w-full max-w-md">
      <div class="surface-pop rounded-[1.5rem] border border-neutral-950 bg-white p-6 shadow-[8px_8px_0_#111] sm:p-7">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Create Account</p>
        <h2 class="mt-3 text-3xl font-black text-neutral-950">注册</h2>

        <form class="mt-6 space-y-5" @submit.prevent="handleRegister">
          <div v-if="error" class="rounded-xl border border-neutral-950 bg-neutral-50 p-3 text-sm font-semibold text-neutral-950">
            {{ error }}
          </div>

          <div>
            <label class="mb-2 block text-sm font-black text-neutral-800">用户名</label>
            <input
              v-model="username"
              type="text"
              class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950"
              placeholder="请输入用户名"
            />
          </div>

          <div>
            <label class="mb-2 block text-sm font-black text-neutral-800">邮箱</label>
            <input
              v-model="email"
              type="email"
              class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950"
              placeholder="请输入邮箱"
            />
          </div>

          <div>
            <label class="mb-2 block text-sm font-black text-neutral-800">密码</label>
            <input
              v-model="password"
              type="password"
              class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950"
              placeholder="请输入密码"
            />
          </div>

          <div>
            <label class="mb-2 block text-sm font-black text-neutral-800">确认密码</label>
            <input
              v-model="confirmPassword"
              type="password"
              class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950"
              placeholder="请再次输入密码"
            />
          </div>

          <div v-if="captchaEnabled">
            <div class="mb-2 flex items-center justify-between gap-3 text-sm font-black text-neutral-800">
              <label>验证码</label>
              <button type="button" class="action-pill border border-neutral-300 px-3 py-1 text-xs text-neutral-600 hover:border-neutral-950 hover:text-neutral-950 disabled:opacity-50" :disabled="captchaLoading" @click="refreshCaptcha">
                {{ captchaLoading ? '刷新中...' : '刷新验证码' }}
              </button>
            </div>
            <div class="grid gap-3 sm:grid-cols-[1fr_11rem]">
              <input
                v-model="captchaCode"
                type="text"
                class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950"
                placeholder="请输入验证码"
              />
              <div class="flex h-12 items-center justify-center overflow-hidden rounded-xl border border-neutral-300 bg-neutral-50">
                <img v-if="captchaImage" :src="captchaImage" alt="验证码" class="h-full w-full object-cover" />
                <span v-else class="text-xs text-neutral-400">验证码加载中</span>
              </div>
            </div>
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="btn-dark action-pill w-full border py-3 text-sm font-black disabled:opacity-50"
          >
            {{ loading ? '注册中...' : '注册' }}
          </button>

          <div class="text-center text-sm text-neutral-500">
            已有账号？
            <RouterLink to="/login" class="link-underline font-bold text-neutral-950">立即登录</RouterLink>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
