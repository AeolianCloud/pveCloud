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
  <section class="profile-page content-page">
    <div class="profile-shell">
      <div class="profile-hero">
        <p class="eyebrow">ACCOUNT PROFILE</p>
        <h1>管理你的登录资料</h1>
        <p>当前阶段只开放邮箱、显示名称和密码修改。用户名保持只读。</p>
        <div class="profile-chip-row">
          <span class="profile-chip">用户名只读</span>
          <span class="profile-chip profile-chip--accent">密码修改后其它会话失效</span>
        </div>
      </div>

      <div class="profile-grid">
        <form class="profile-card" @submit.prevent="handleProfileSubmit">
          <div class="card-heading">
            <span>基础资料</span>
            <h2>编辑账户信息</h2>
          </div>
          <div class="field-list">
            <label class="field field--readonly">
              <span>用户名</span>
              <input :value="user?.username" type="text" disabled />
            </label>
            <label class="field">
              <span>邮箱</span>
              <input v-model="profileForm.email" type="email" autocomplete="email" />
            </label>
            <label class="field">
              <span>显示名称</span>
              <input v-model="profileForm.displayName" type="text" autocomplete="name" />
            </label>
          </div>
          <p v-if="profileMessage" class="notice success">{{ profileMessage }}</p>
          <p v-if="profileError" class="notice error">{{ profileError }}</p>
          <button class="btn btn-primary profile-action" type="submit" :disabled="!canSaveProfile">
            {{ profileLoading ? '保存中...' : '保存资料' }}
          </button>
        </form>

        <form class="profile-card profile-card--deep" @submit.prevent="handlePasswordSubmit">
          <div class="card-heading">
            <span>密码安全</span>
            <h2>更新登录密码</h2>
          </div>
          <div class="field-list">
            <label class="field">
              <span>当前密码</span>
              <input v-model="passwordForm.currentPassword" type="password" autocomplete="current-password" />
            </label>
            <label class="field">
              <span>新密码</span>
              <input v-model="passwordForm.password" type="password" autocomplete="new-password" />
            </label>
            <label class="field">
              <span>确认新密码</span>
              <input v-model="passwordForm.confirmPassword" type="password" autocomplete="new-password" />
            </label>
          </div>
          <p v-if="passwordForm.password && passwordForm.confirmPassword && passwordForm.password !== passwordForm.confirmPassword" class="notice error">两次输入的密码不一致</p>
          <p v-if="passwordMessage" class="notice success">{{ passwordMessage }}</p>
          <p v-if="passwordError" class="notice error">{{ passwordError }}</p>
          <button class="btn btn-primary profile-action" type="submit" :disabled="!canChangePassword">
            {{ passwordLoading ? '更新中...' : '更新密码' }}
          </button>
        </form>
      </div>
    </div>
  </section>
</template>

<style scoped>
.profile-page {
  min-height: calc(100vh - 96px);
  padding: 8px 0 48px;
}

.profile-shell {
  display: grid;
  gap: 28px;
  width: min(1180px, calc(100% - 40px));
  margin: 0 auto;
}

.profile-hero {
  display: grid;
  gap: 14px;
  padding: clamp(26px, 4vw, 42px);
  border: 1px solid var(--c-border);
  border-radius: 30px;
  background:
    radial-gradient(circle at 100% 0%, rgba(59, 130, 246, 0.16), transparent 36%),
    var(--c-card);
  box-shadow: var(--shadow);
}

.profile-hero h1 {
  font-size: clamp(2.2rem, 4vw, 3.8rem);
  line-height: 1;
  letter-spacing: -0.06em;
}

.profile-hero p {
  max-width: 760px;
  color: var(--c-text-2);
  line-height: 1.8;
}

.eyebrow {
  color: var(--c-primary);
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.16em;
}

.profile-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.profile-chip {
  padding: 8px 12px;
  border-radius: 999px;
  color: var(--c-text-2);
  background: var(--c-surface-dim);
  border: 1px solid var(--c-border);
  font-size: 0.85rem;
  font-weight: 700;
}

.profile-chip--accent {
  color: var(--c-primary);
  background: var(--c-primary-soft);
}

.profile-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.profile-card {
  display: grid;
  gap: 18px;
  padding: 26px;
  border: 1px solid var(--c-border);
  border-radius: 26px;
  background: var(--c-card);
  box-shadow: var(--shadow-sm);
}

.profile-card--deep {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.02), transparent),
    var(--c-card);
}

.card-heading {
  display: grid;
  gap: 6px;
}

.card-heading span {
  color: var(--c-text-3);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.card-heading h2 {
  font-size: 1.5rem;
  letter-spacing: -0.04em;
}

.field-list {
  display: grid;
  gap: 14px;
}

.field {
  display: grid;
  gap: 8px;
  font-weight: 700;
}

.field span {
  color: var(--c-text-2);
}

.field input {
  min-height: 48px;
  padding: 0 14px;
  border: 1px solid var(--c-border);
  border-radius: 14px;
  color: var(--c-text);
  background: var(--c-surface-dim);
}

.field input:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.field--readonly input {
  background: rgba(255, 255, 255, 0.02);
}

.notice {
  margin: 0;
  padding: 12px 14px;
  border-radius: 14px;
  line-height: 1.6;
}

.notice.success {
  color: var(--c-success);
  background: var(--c-success-soft);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.notice.error {
  color: var(--c-error);
  background: var(--c-error-soft);
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.profile-action {
  justify-self: start;
}

@media (max-width: 900px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .profile-page {
    padding-top: 0;
  }

  .profile-shell {
    width: min(100% - 28px, 1180px);
  }

  .profile-hero,
  .profile-card {
    padding: 22px;
  }
}
</style>
