<script setup lang="ts">
import { computed, reactive, ref, watchEffect } from 'vue'
import { storeToRefs } from 'pinia'

import { changePassword, updateProfile } from '../../api/user-profile'
import { useWebAuthStore } from '../../store/modules/auth'

const authStore = useWebAuthStore()
const { user } = storeToRefs(authStore)

const profileForm = reactive({ email: '', displayName: '' })
const passwordForm = reactive({ currentPassword: '', password: '', confirmPassword: '' })
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
const canChangePassword = computed(() => (
  passwordForm.currentPassword.length >= 6 &&
  passwordForm.password.length >= 6 &&
  passwordForm.password === passwordForm.confirmPassword &&
  !passwordLoading.value
))

function errorText(error: unknown, fallback: string) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) return response.data.message
  }
  if (typeof error === 'object' && error !== null && 'request' in error) return '网络连接失败，请检查后重试'
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
    await changePassword({ current_password: passwordForm.currentPassword, password: passwordForm.password })
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
  <section class="profile-page page-shell">
    <div class="page-hero surface">
      <p class="section-label">Account Profile</p>
      <h1 class="page-title">账号资料</h1>
      <p class="page-copy">用户名只读。邮箱、显示名称和密码可在此维护。</p>
    </div>

    <div class="profile-grid">
      <form class="profile-card card" @submit.prevent="handleProfileSubmit">
        <div class="card-heading">
          <span class="section-label">Profile</span>
          <h2>基础资料</h2>
        </div>
        <label class="field"><span>用户名</span><span class="field-control readonly"><input :value="user?.username" type="text" disabled /></span></label>
        <label class="field"><span>邮箱</span><span class="field-control"><input v-model="profileForm.email" type="email" autocomplete="email" /></span></label>
        <label class="field"><span>显示名称</span><span class="field-control"><input v-model="profileForm.displayName" type="text" autocomplete="name" placeholder="可为空" /></span></label>
        <p v-if="profileMessage" class="notice success">{{ profileMessage }}</p>
        <p v-if="profileError" class="notice error">{{ profileError }}</p>
        <button class="btn btn-primary" type="submit" :disabled="!canSaveProfile">{{ profileLoading ? '保存中...' : '保存资料' }}</button>
      </form>

      <form class="profile-card card" @submit.prevent="handlePasswordSubmit">
        <div class="card-heading">
          <span class="section-label">Security</span>
          <h2>修改密码</h2>
        </div>
        <label class="field"><span>当前密码</span><span class="field-control"><input v-model="passwordForm.currentPassword" type="password" autocomplete="current-password" /></span></label>
        <label class="field"><span>新密码</span><span class="field-control"><input v-model="passwordForm.password" type="password" autocomplete="new-password" /></span></label>
        <label class="field"><span>确认新密码</span><span class="field-control"><input v-model="passwordForm.confirmPassword" type="password" autocomplete="new-password" /></span></label>
        <p v-if="passwordForm.password && passwordForm.confirmPassword && passwordForm.password !== passwordForm.confirmPassword" class="notice error">两次输入的密码不一致</p>
        <p v-if="passwordMessage" class="notice success">{{ passwordMessage }}</p>
        <p v-if="passwordError" class="notice error">{{ passwordError }}</p>
        <button class="btn btn-primary" type="submit" :disabled="!canChangePassword">{{ passwordLoading ? '更新中...' : '更新密码' }}</button>
      </form>
    </div>
  </section>
</template>

<style scoped>
.profile-page {
  display: grid;
  gap: 22px;
}

.page-hero {
  display: grid;
  gap: 12px;
  padding: clamp(24px, 4vw, 38px);
}

.profile-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.profile-card {
  display: grid;
  gap: 18px;
  align-content: start;
  padding: 22px;
}

.card-heading {
  display: grid;
  gap: 8px;
}

.card-heading h2 {
  font-size: 1.5rem;
  letter-spacing: -0.04em;
}

.readonly {
  background: var(--c-surface-dim);
}

@media (max-width: 820px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}
</style>
