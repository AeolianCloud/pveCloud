<script setup lang="ts">
import { computed, reactive, ref, watchEffect } from 'vue'
import { storeToRefs } from 'pinia'

import { changePassword, updateProfile } from '../../api/user-profile'
import { useWebAuthStore } from '../../store/modules/auth'

const authStore = useWebAuthStore()
const { user } = storeToRefs(authStore)

const profileForm = reactive({
  email: '',
  displayName: '',
})
const passwordForm = reactive({
  currentPassword: '',
  password: '',
  confirmPassword: '',
})
const profileLoading = ref(false)
const passwordLoading = ref(false)
const profileMessage = ref('')
const passwordMessage = ref('')
const profileError = ref('')
const passwordError = ref('')

watchEffect(() => {
  profileForm.email = user.value?.email ?? ''
  profileForm.displayName = user.value?.display_name ?? ''
})

const canSaveProfile = computed(() => profileForm.email.trim() !== '' && !profileLoading.value)
const canChangePassword = computed(() => {
  return (
    passwordForm.currentPassword.length >= 6 &&
    passwordForm.password.length >= 6 &&
    passwordForm.password === passwordForm.confirmPassword &&
    !passwordLoading.value
  )
})

function errorText(error: unknown, fallback: string) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (typeof error === 'object' && error !== null && 'request' in error) {
    return '网络连接失败，请检查后重试'
  }
  return fallback
}

async function handleProfileSubmit() {
  if (!canSaveProfile.value) return
  profileLoading.value = true
  profileError.value = ''
  profileMessage.value = ''
  try {
    const result = await updateProfile({
      email: profileForm.email.trim(),
      display_name: profileForm.displayName.trim() || null,
    })
    authStore.setAuthState(result)
    profileMessage.value = '资料已保存'
  } catch (error) {
    profileError.value = errorText(error, '资料保存失败，请稍后再试')
  } finally {
    profileLoading.value = false
  }
}

async function handlePasswordSubmit() {
  if (!canChangePassword.value) return
  passwordLoading.value = true
  passwordError.value = ''
  passwordMessage.value = ''
  try {
    await changePassword({
      current_password: passwordForm.currentPassword,
      password: passwordForm.password,
    })
    passwordForm.currentPassword = ''
    passwordForm.password = ''
    passwordForm.confirmPassword = ''
    passwordMessage.value = '密码已更新，其它登录会话已失效'
  } catch (error) {
    passwordError.value = errorText(error, '密码修改失败，请稍后再试')
  } finally {
    passwordLoading.value = false
  }
}
</script>

<template>
  <section class="page content-page">
    <div class="section-pad">
      <div class="sec-header" style="margin-bottom: clamp(28px, 4vw, 48px);">
        <p class="label">账号资料</p>
        <h2>管理你的登录资料</h2>
        <p>当前阶段只开放邮箱、显示名称和密码修改。用户名保持只读。</p>
      </div>

      <div class="account-grid">
        <form class="auth-form account-panel" @submit.prevent="handleProfileSubmit">
          <h2>基础资料</h2>
          <label>
            <span>用户名</span>
            <input :value="user?.username" type="text" disabled />
          </label>
          <label>
            <span>邮箱</span>
            <input v-model="profileForm.email" type="email" autocomplete="email" />
          </label>
          <label>
            <span>显示名称</span>
            <input v-model="profileForm.displayName" type="text" autocomplete="name" />
          </label>
          <p v-if="profileMessage" class="hint success-text">{{ profileMessage }}</p>
          <p v-if="profileError" class="hint error-text">{{ profileError }}</p>
          <button class="btn btn-primary" type="submit" :disabled="!canSaveProfile">
            {{ profileLoading ? '保存中...' : '保存资料' }}
          </button>
        </form>

        <form class="auth-form account-panel" @submit.prevent="handlePasswordSubmit">
          <h2>修改密码</h2>
          <label>
            <span>当前密码</span>
            <input v-model="passwordForm.currentPassword" type="password" autocomplete="current-password" />
          </label>
          <label>
            <span>新密码</span>
            <input v-model="passwordForm.password" type="password" autocomplete="new-password" />
          </label>
          <label>
            <span>确认新密码</span>
            <input v-model="passwordForm.confirmPassword" type="password" autocomplete="new-password" />
          </label>
          <p v-if="passwordForm.password && passwordForm.confirmPassword && passwordForm.password !== passwordForm.confirmPassword" class="hint error-text">两次输入的密码不一致</p>
          <p v-if="passwordMessage" class="hint success-text">{{ passwordMessage }}</p>
          <p v-if="passwordError" class="hint error-text">{{ passwordError }}</p>
          <button class="btn btn-primary" type="submit" :disabled="!canChangePassword">
            {{ passwordLoading ? '更新中...' : '更新密码' }}
          </button>
        </form>
      </div>
    </div>
  </section>
</template>
