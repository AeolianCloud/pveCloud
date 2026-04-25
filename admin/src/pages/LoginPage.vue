<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { KeyRound, LogIn, ServerCog, Shield, UserRound } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const form = reactive({
  username: '',
  password: '',
})

const loading = ref(false)
const errorMessage = ref('')

const canSubmit = computed(() => form.username.trim().length > 0 && form.password.length >= 8)

async function submit() {
  if (!canSubmit.value || loading.value) {
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    await auth.login({
      username: form.username.trim(),
      password: form.password,
    })
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard'
    await router.replace(redirect)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="login-page">
    <section class="login-brand" aria-label="pveCloud 管理后台">
      <div class="brand-mark">
        <ServerCog :size="30" aria-hidden="true" />
      </div>
      <div>
        <span class="eyebrow">pveCloud Admin</span>
        <h1>管理后台</h1>
        <p>订单、支付、实例和工单的运营入口。</p>
      </div>
      <div class="brand-status">
        <Shield :size="18" aria-hidden="true" />
        <span>RBAC 权限控制</span>
      </div>
    </section>

    <section class="login-panel" aria-label="管理员登录">
      <div class="panel-heading">
        <KeyRound :size="22" aria-hidden="true" />
        <div>
          <h2>管理员登录</h2>
          <p>使用后台账号进入控制台。</p>
        </div>
      </div>

      <form class="login-form" @submit.prevent="submit">
        <label>
          <span>账号</span>
          <div class="input-shell">
            <UserRound :size="18" aria-hidden="true" />
            <input
              v-model="form.username"
              autocomplete="username"
              name="username"
              placeholder="用户名或邮箱"
              type="text"
            />
          </div>
        </label>

        <label>
          <span>密码</span>
          <div class="input-shell">
            <KeyRound :size="18" aria-hidden="true" />
            <input
              v-model="form.password"
              autocomplete="current-password"
              name="password"
              placeholder="至少 8 位"
              type="password"
            />
          </div>
        </label>

        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

        <button class="primary-button" type="submit" :disabled="!canSubmit || loading">
          <LogIn :size="18" aria-hidden="true" />
          <span>{{ loading ? '登录中' : '登录' }}</span>
        </button>
      </form>
    </section>
  </main>
</template>
