<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { confirmPasswordReset, getPasswordResetConfirmCaptcha } from '../../api/auth'
import { getApiErrorMessage } from '../../api/request'
import { getSiteConfig } from '../../api/site-config'

const route = useRoute()
const router = useRouter()
const token = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const success = ref(false)
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
    const captcha = await getPasswordResetConfirmCaptcha()
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
  token.value = typeof route.query.token === 'string' ? route.query.token : ''
  try {
    const config = await getSiteConfig()
    captchaEnabled.value = config.password_reset_confirm_captcha_enabled
  } catch {
    captchaEnabled.value = false
  }

  if (token.value) {
    await refreshCaptcha()
  }
})

const handleSubmit = async () => {
  if (!token.value) {
    error.value = '无效的重置链接'
    return
  }

  if (!password.value || !confirmPassword.value) {
    error.value = '请填写所有字段'
    return
  }

  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }
  if (captchaEnabled.value && (!captchaId.value || !captchaCode.value)) {
    error.value = '请输入验证码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    await confirmPasswordReset({
      token: token.value,
      password: password.value,
      ...(captchaEnabled.value ? { captcha_id: captchaId.value, captcha_code: captchaCode.value } : {}),
    })
    success.value = true
    setTimeout(() => {
      router.push('/login')
    }, 3000)
  } catch (err) {
    error.value = getApiErrorMessage(err, '重置密码失败，请稍后重试')
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
        <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Reset Password</p>
        <h2 class="mt-3 text-3xl font-black text-neutral-950">重置密码</h2>
        <p class="mt-3 text-sm text-neutral-500">请输入您的新密码</p>

        <div class="mt-6">
        <div v-if="success" class="rounded-xl border border-neutral-300 bg-neutral-50 p-4">
          <h3 class="text-sm font-black text-neutral-950">密码重置成功</h3>
          <p class="mt-2 text-sm leading-6 text-neutral-600">密码已成功重置，即将跳转到登录页面。</p>
        </div>

        <div v-else-if="!token" class="rounded-xl border border-neutral-950 bg-neutral-50 p-4">
          <h3 class="text-sm font-black text-neutral-950">无效的重置链接</h3>
          <p class="mt-2 text-sm leading-6 text-neutral-600">重置链接无效或已过期，请重新申请重置密码。</p>
          <RouterLink to="/forgot-password" class="mt-4 inline-flex rounded-full border border-neutral-950 px-4 py-2 text-sm font-black text-neutral-950 hover:bg-neutral-950 hover:text-white">重新申请</RouterLink>
        </div>

        <form v-else class="space-y-5" @submit.prevent="handleSubmit">
          <div v-if="error" class="rounded-xl border border-neutral-950 bg-neutral-50 p-3 text-sm font-bold text-neutral-950">{{ error }}</div>
          <div>
            <label for="password" class="mb-2 block text-sm font-black text-neutral-800">新密码</label>
            <input id="password" v-model="password" name="password" type="password" required class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入新密码" />
          </div>
          <div>
            <label for="confirm-password" class="mb-2 block text-sm font-black text-neutral-800">确认新密码</label>
            <input id="confirm-password" v-model="confirmPassword" name="confirm-password" type="password" required class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请再次输入新密码" />
          </div>
          <div v-if="captchaEnabled">
            <div class="mb-2 flex items-center justify-between gap-3 text-sm font-black text-neutral-800">
              <label>验证码</label>
              <button type="button" class="text-neutral-500 hover:text-neutral-950" :disabled="captchaLoading" @click="refreshCaptcha">
                {{ captchaLoading ? '刷新中...' : '刷新验证码' }}
              </button>
            </div>
            <div class="grid gap-3 sm:grid-cols-[1fr_11rem]">
              <input v-model="captchaCode" type="text" class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入验证码" />
              <div class="flex h-12 items-center justify-center overflow-hidden rounded-xl border border-neutral-300 bg-neutral-50">
                <img v-if="captchaImage" :src="captchaImage" alt="验证码" class="h-full w-full object-cover" />
                <span v-else class="text-xs text-neutral-400">验证码加载中</span>
              </div>
            </div>
          </div>
          <button type="submit" :disabled="loading" class="btn-dark w-full rounded-full border py-3 text-sm font-black disabled:opacity-50">{{ loading ? '重置中...' : '重置密码' }}</button>
        </form>
        </div>
      </div>
    </div>
  </div>
</template>
