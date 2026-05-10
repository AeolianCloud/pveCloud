<script setup lang="ts">
import { ref } from 'vue'

const username = ref('user')
const email = ref('user@example.com')
const displayName = ref('')
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

const handleUpdateProfile = async () => {
  loading.value = true
  error.value = ''
  success.value = ''

  try {
    success.value = '资料更新成功'
  } catch (err) {
    error.value = '更新失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

const handleChangePassword = async () => {
  if (!currentPassword.value || !newPassword.value || !confirmPassword.value) {
    error.value = '请填写所有密码字段'
    return
  }

  if (newPassword.value !== confirmPassword.value) {
    error.value = '两次输入的新密码不一致'
    return
  }

  loading.value = true
  error.value = ''
  success.value = ''

  try {
    success.value = '密码修改成功'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (err) {
    error.value = '密码修改失败，请检查当前密码是否正确'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="bg-white">
    <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Profile</p>
      <h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">账号资料</h1>

      <div v-if="error" class="mt-6 rounded-xl border border-neutral-950 bg-neutral-50 p-4 text-sm font-bold text-neutral-950">{{ error }}</div>
      <div v-if="success" class="mt-6 rounded-xl border border-neutral-300 bg-white p-4 text-sm font-bold text-neutral-950">{{ success }}</div>

      <div class="mt-8 grid gap-6 lg:grid-cols-2">
        <section class="rounded-[1.5rem] border border-neutral-950 bg-white p-6 shadow-[8px_8px_0_#111]">
          <h2 class="text-xl font-black text-neutral-950">基本资料</h2>
          <form class="mt-6 space-y-5" @submit.prevent="handleUpdateProfile">
            <div>
              <label class="mb-2 block text-sm font-black text-neutral-800">用户名</label>
              <input :value="username" disabled class="w-full rounded-xl border border-neutral-200 bg-neutral-50 px-4 py-3 text-sm text-neutral-500" />
            </div>
            <div>
              <label for="email" class="mb-2 block text-sm font-black text-neutral-800">邮箱</label>
              <input id="email" v-model="email" type="email" class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" />
            </div>
            <div>
              <label for="display-name" class="mb-2 block text-sm font-black text-neutral-800">显示名称</label>
              <input id="display-name" v-model="displayName" type="text" class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入显示名称" />
            </div>
            <button type="submit" :disabled="loading" class="btn-dark w-full rounded-full border py-3 text-sm font-black disabled:opacity-50">{{ loading ? '保存中...' : '保存资料' }}</button>
          </form>
        </section>

        <section class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6">
          <h2 class="text-xl font-black text-neutral-950">修改密码</h2>
          <form class="mt-6 space-y-5" @submit.prevent="handleChangePassword">
            <div>
              <label for="current-password" class="mb-2 block text-sm font-black text-neutral-800">当前密码</label>
              <input id="current-password" v-model="currentPassword" type="password" class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入当前密码" />
            </div>
            <div>
              <label for="new-password" class="mb-2 block text-sm font-black text-neutral-800">新密码</label>
              <input id="new-password" v-model="newPassword" type="password" class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入新密码" />
            </div>
            <div>
              <label for="confirm-new-password" class="mb-2 block text-sm font-black text-neutral-800">确认新密码</label>
              <input id="confirm-new-password" v-model="confirmPassword" type="password" class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请再次输入新密码" />
            </div>
            <button type="submit" :disabled="loading" class="w-full rounded-full border border-neutral-950 bg-white py-3 text-sm font-black text-neutral-950 hover:bg-neutral-950 hover:text-white disabled:opacity-50">{{ loading ? '修改中...' : '修改密码' }}</button>
          </form>
        </section>
      </div>
    </div>
  </div>
</template>
