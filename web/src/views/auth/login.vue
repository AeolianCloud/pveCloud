<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getSiteConfig } from '../../api/site-config'
import { getApiErrorMessage } from '../../api/request'
import { getLoginCaptcha } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const account = ref('')
const password = ref('')
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
    const captcha = await getLoginCaptcha()
    captchaId.value = captcha.captcha_id
    captchaImage.value = captcha.image
    captchaCode.value = ''
  } catch (err) {
    error.value = getApiErrorMessage(err, '验证码加载失败，请稍后重试')
  } finally {
    captchaLoading.value = false
  }
}

const normalizeRedirect = (value: unknown) => {
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//')) {
    return '/user'
  }
  return value
}

onMounted(async () => {
  try {
    const config = await getSiteConfig()
    captchaEnabled.value = config.login_captcha_enabled
  } catch {
    captchaEnabled.value = false
  }

  await refreshCaptcha()
})

const handleLogin = async () => {
  if (!account.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }
  if (captchaEnabled.value && (!captchaId.value || !captchaCode.value)) {
    error.value = '请输入验证码'
    return
  }

  loading.value = true
  error.value = ''
  try {
    await authStore.loginWithPassword({
      account: account.value,
      password: password.value,
      ...(captchaEnabled.value ? { captcha_id: captchaId.value, captcha_code: captchaCode.value } : {}),
    })
    await router.push(normalizeRedirect(route.query.redirect))
  } catch (err) {
    error.value = getApiErrorMessage(err, '登录失败')
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
        <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Account Login</p>
        <h2 class="mt-3 text-3xl font-black text-neutral-950">登录</h2>

        <form class="mt-6 space-y-5" @submit.prevent="handleLogin">
          <div v-if="error" class="rounded-xl border border-neutral-950 bg-neutral-50 p-3 text-sm font-semibold text-neutral-950">
            {{ error }}
          </div>

          <div>
            <label class="mb-2 block text-sm font-black text-neutral-800">用户名或邮箱</label>
            <input
              v-model="account"
              type="text"
              class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950"
              placeholder="请输入用户名或邮箱"
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

          <div v-if="captchaEnabled">
            <div class="mb-2 flex items-center justify-between gap-3 text-sm font-black text-neutral-800">
              <label>验证码</label>
              <button type="button" class="text-neutral-500 hover:text-neutral-950" :disabled="captchaLoading" @click="refreshCaptcha">
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

          <div class="flex items-center justify-between text-sm">
            <label class="flex items-center gap-2 text-neutral-600">
              <input type="checkbox" class="rounded border-neutral-300 text-neutral-950" />
              记住我
            </label>
            <RouterLink to="/forgot-password" class="link-underline font-bold text-neutral-950">
              忘记密码？
            </RouterLink>
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="btn-dark w-full rounded-full border py-3 text-sm font-black disabled:opacity-50"
          >
            {{ loading ? '登录中...' : '登录' }}
          </button>

          <div class="text-center text-sm text-neutral-500">
            还没有账号？
            <RouterLink to="/register" class="link-underline font-bold text-neutral-950">免费注册</RouterLink>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
