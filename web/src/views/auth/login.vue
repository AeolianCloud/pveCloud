<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    router.push('/user')
  } catch (err) {
    error.value = '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-white px-4 py-12">
    <div class="w-full max-w-md">
      <div class="mb-8 text-center">
        <RouterLink to="/" class="inline-flex items-center gap-3 text-neutral-950">
          <span class="flex h-10 w-10 items-center justify-center rounded-xl border border-neutral-950 text-sm font-black">P</span>
          <span class="text-xl font-black tracking-tight">PVECloud</span>
        </RouterLink>
      </div>

      <div class="rounded-[1.5rem] border border-neutral-950 bg-white p-7 shadow-[8px_8px_0_#111]">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Account Login</p>
        <h2 class="mt-3 text-3xl font-black text-neutral-950">登录</h2>

        <form class="mt-6 space-y-5" @submit.prevent="handleLogin">
          <div v-if="error" class="rounded-xl border border-neutral-950 bg-neutral-50 p-3 text-sm font-semibold text-neutral-950">
            {{ error }}
          </div>

          <div>
            <label class="mb-2 block text-sm font-black text-neutral-800">用户名</label>
            <input
              v-model="username"
              type="text"
              class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950"
              placeholder="请输入用户名"
            />
          </div>

          <div>
            <label class="mb-2 block text-sm font-black text-neutral-800">密码</label>
            <input
              v-model="password"
              type="password"
              class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950"
              placeholder="请输入密码"
            />
          </div>

          <div class="flex items-center justify-between text-sm">
            <label class="flex items-center gap-2 text-neutral-600">
              <input type="checkbox" class="rounded border-neutral-300 text-neutral-950" />
              记住我
            </label>
            <RouterLink to="/forgot-password" class="font-bold text-neutral-950 underline decoration-2 underline-offset-4">
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
            <RouterLink to="/register" class="font-bold text-neutral-950">免费注册</RouterLink>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
