<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getPasswordResetRequestCaptcha, requestPasswordReset } from '../../api/auth'
import { getApiErrorMessage } from '../../api/request'
import { getSiteConfig } from '../../api/site-config'

const email = ref('')
const loading = ref(false)
const success = ref(false)
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
    const captcha = await getPasswordResetRequestCaptcha()
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
    captchaEnabled.value = config.password_reset_request_captcha_enabled
  } catch {
    captchaEnabled.value = false
  }

  await refreshCaptcha()
})

const handleSubmit = async () => {
  if (!email.value) {
    error.value = '请输入邮箱地址'
    return
  }
  if (captchaEnabled.value && (!captchaId.value || !captchaCode.value)) {
    error.value = '请输入验证码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    await requestPasswordReset({
      email: email.value,
      ...(captchaEnabled.value ? { captcha_id: captchaId.value, captcha_code: captchaCode.value } : {}),
    })
    success.value = true
  } catch (err) {
    error.value = getApiErrorMessage(err, '发送重置邮件失败，请稍后重试')
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
        <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Password Reset</p>
        <h2 class="mt-3 text-3xl font-black text-neutral-950">忘记密码</h2>
        <p class="mt-3 text-sm text-neutral-500">输入您的邮箱地址，我们将发送重置链接</p>

        <div class="mt-6">
        <div v-if="success" class="rounded-xl border border-neutral-300 bg-neutral-50 p-4">
          <h3 class="text-sm font-black text-neutral-950">重置邮件已发送</h3>
          <p class="mt-2 text-sm leading-6 text-neutral-600">请检查您的邮箱，点击重置链接修改密码。</p>
          <RouterLink to="/login" class="action-pill mt-4 border border-neutral-950 px-4 py-2 text-sm font-black text-neutral-950 hover:bg-neutral-950 hover:text-white">返回登录</RouterLink>
        </div>

        <form v-else class="space-y-5" @submit.prevent="handleSubmit">
          <div v-if="error" class="rounded-xl border border-neutral-950 bg-neutral-50 p-3 text-sm font-bold text-neutral-950">{{ error }}</div>
          <div>
            <label for="email" class="mb-2 block text-sm font-black text-neutral-800">邮箱地址</label>
            <input id="email" v-model="email" name="email" type="email" required class="field-focus w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入邮箱地址" />
          </div>
          <div v-if="captchaEnabled">
            <div class="mb-2 flex items-center justify-between gap-3 text-sm font-black text-neutral-800">
              <label>验证码</label>
              <button type="button" class="action-pill border border-neutral-300 px-3 py-1 text-xs text-neutral-600 hover:border-neutral-950 hover:text-neutral-950 disabled:opacity-50" :disabled="captchaLoading" @click="refreshCaptcha">
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
          <button type="submit" :disabled="loading" class="btn-dark action-pill w-full border py-3 text-sm font-black disabled:opacity-50">{{ loading ? '发送中...' : '发送重置链接' }}</button>
          <div class="text-center text-sm">
            <RouterLink to="/login" class="link-underline font-black text-neutral-950">返回登录</RouterLink>
          </div>
        </form>
        </div>
      </div>
    </div>
  </div>
</template>
