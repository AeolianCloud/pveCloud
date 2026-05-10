<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')

const handleRegister = async () => {
  if (!username.value || !email.value || !password.value || !confirmPassword.value) {
    error.value = '请填写所有字段'
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = '密码不一致'
    return
  }
  loading.value = true
  error.value = ''
  try {
    router.push('/user')
  } catch (err) {
    error.value = '注册失败'
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

          <button
            type="submit"
            :disabled="loading"
            class="btn-dark flex w-full items-center justify-center rounded-full border py-3 text-sm font-black disabled:cursor-not-allowed disabled:opacity-50"
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
