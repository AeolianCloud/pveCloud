import { computed, ref, watch, type Ref } from 'vue'

import type { CaptchaResponse } from '../api/auth'

function captchaErrorText(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (typeof error === 'object' && error !== null && 'request' in error) {
    return '验证码加载失败，请检查网络后重试'
  }
  return '验证码加载失败，请稍后再试'
}

export function useAuthCaptcha(enabled: Ref<boolean>, fetchCaptcha: () => Promise<CaptchaResponse>) {
  const captchaId = ref('')
  const captchaCode = ref('')
  const captchaImage = ref('')
  const captchaLoading = ref(false)
  const captchaError = ref('')

  const captchaReady = computed(() => {
    if (!enabled.value) return true
    return !captchaLoading.value && captchaId.value !== '' && captchaImage.value !== ''
  })

  function clearCaptcha() {
    captchaId.value = ''
    captchaCode.value = ''
    captchaImage.value = ''
    captchaError.value = ''
  }

  async function refreshCaptcha() {
    if (!enabled.value) {
      clearCaptcha()
      return
    }
    captchaLoading.value = true
    captchaError.value = ''
    try {
      const result = await fetchCaptcha()
      captchaId.value = result.captcha_id
      captchaCode.value = ''
      captchaImage.value = result.image
    } catch (error) {
      captchaId.value = ''
      captchaCode.value = ''
      captchaImage.value = ''
      captchaError.value = captchaErrorText(error)
    } finally {
      captchaLoading.value = false
    }
  }

  watch(
    enabled,
    (value) => {
      if (!value) {
        clearCaptcha()
        return
      }
      void refreshCaptcha()
    },
    { immediate: true },
  )

  return {
    captchaId,
    captchaCode,
    captchaImage,
    captchaLoading,
    captchaError,
    captchaReady,
    refreshCaptcha,
  }
}
