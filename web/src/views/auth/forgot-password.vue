<script setup lang="ts">
import { ref } from 'vue'

const email = ref('')
const loading = ref(false)
const success = ref(false)
const error = ref('')

const handleSubmit = async () => {
  if (!email.value) {
    error.value = '请输入邮箱地址'
    return
  }

  loading.value = true
  error.value = ''

  try {
    success.value = true
  } catch (err) {
    error.value = '发送重置邮件失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen flex-col justify-center bg-white px-4 py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <p class="text-center text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Password Reset</p>
      <h2 class="mt-4 text-center text-3xl font-black text-neutral-950">忘记密码</h2>
      <p class="mt-3 text-center text-sm text-neutral-500">输入您的邮箱地址，我们将发送重置链接</p>
    </div>

    <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
      <div class="rounded-[1.5rem] border border-neutral-950 bg-white p-7 shadow-[8px_8px_0_#111]">
        <div v-if="success" class="rounded-xl border border-neutral-300 bg-neutral-50 p-4">
          <h3 class="text-sm font-black text-neutral-950">重置邮件已发送</h3>
          <p class="mt-2 text-sm leading-6 text-neutral-600">请检查您的邮箱，点击重置链接修改密码。</p>
          <RouterLink to="/login" class="mt-4 inline-flex rounded-full border border-neutral-950 px-4 py-2 text-sm font-black text-neutral-950 hover:bg-neutral-950 hover:text-white">返回登录</RouterLink>
        </div>

        <form v-else class="space-y-5" @submit.prevent="handleSubmit">
          <div v-if="error" class="rounded-xl border border-neutral-950 bg-neutral-50 p-3 text-sm font-bold text-neutral-950">{{ error }}</div>
          <div>
            <label for="email" class="mb-2 block text-sm font-black text-neutral-800">邮箱地址</label>
            <input id="email" v-model="email" name="email" type="email" required class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入邮箱地址" />
          </div>
          <button type="submit" :disabled="loading" class="btn-dark w-full rounded-full border py-3 text-sm font-black disabled:opacity-50">{{ loading ? '发送中...' : '发送重置链接' }}</button>
          <div class="text-center text-sm">
            <RouterLink to="/login" class="font-black text-neutral-950 underline decoration-2 underline-offset-4">返回登录</RouterLink>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
