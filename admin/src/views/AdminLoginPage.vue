<template>
  <section class="panel login">
    <div>
      <p class="tag">ADMIN LOGIN</p>
      <h2>进入管理后台</h2>
    </div>
    <form class="form" @submit.prevent="submit">
      <label>
        <span>用户名</span>
        <input v-model="username" type="text" placeholder="admin" />
      </label>
      <label>
        <span>密码</span>
        <input v-model="password" type="password" placeholder="请输入密码" />
      </label>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      <button type="submit" :disabled="submitting">{{ submitting ? '登录中...' : '登录后台' }}</button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const username = ref('')
const password = ref('')
const errorMessage = ref('')
const submitting = ref(false)
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

async function submit() {
  errorMessage.value = ''
  submitting.value = true

  try {
    await authStore.login(username.value, password.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/products'
    await router.push(redirect)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.panel {
  padding: 28px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.login {
  display: grid;
  grid-template-columns: 1fr 0.9fr;
  gap: 24px;
}

.tag {
  margin: 0 0 8px;
  color: #557257;
}

.form {
  display: grid;
  gap: 16px;
}

label {
  display: grid;
  gap: 8px;
}

input,
button {
  border-radius: 16px;
  padding: 14px 16px;
  font: inherit;
}

input {
  border: 1px solid #c8d7cf;
}

button {
  border: 0;
  background: #1d3a2f;
  color: #fff;
}

.error {
  margin: 0;
  color: #b42318;
}

@media (max-width: 800px) {
  .login {
    grid-template-columns: 1fr;
  }
}
</style>
