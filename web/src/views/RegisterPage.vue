<template>
  <section class="panel stack">
    <div>
      <p class="tag">REGISTER</p>
      <h2>创建你的 pveCloud 账户</h2>
    </div>
    <form class="grid" @submit.prevent="submit">
      <label>
        <span>手机号</span>
        <input v-model="phone" type="tel" placeholder="13800000000" />
      </label>
      <label>
        <span>邮箱</span>
        <input v-model="email" type="email" placeholder="name@example.com" />
      </label>
      <label>
        <span>密码</span>
        <input v-model="password" type="password" placeholder="设置登录密码" />
      </label>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      <button type="submit" :disabled="submitting">{{ submitting ? '注册中...' : '提交注册' }}</button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const phone = ref('')
const email = ref('')
const password = ref('')
const errorMessage = ref('')
const submitting = ref(false)
const authStore = useAuthStore()
const router = useRouter()

async function submit() {
  errorMessage.value = ''
  submitting.value = true

  try {
    await authStore.register(phone.value, email.value, password.value)
    await router.push('/products')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '注册失败'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.panel {
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(29, 42, 51, 0.08);
  padding: 28px;
}

.stack,
.grid {
  display: grid;
  gap: 18px;
}

.tag {
  margin: 0 0 8px;
  font-size: 13px;
  letter-spacing: 0.18em;
  color: #9b5d32;
}

h2 {
  margin: 0;
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
  border: 1px solid #c8d7e1;
  background: #f9fbfc;
}

button {
  border: 0;
  background: #1d2a33;
  color: #fff;
}

button:disabled {
  opacity: 0.72;
}

.error {
  margin: 0;
  color: #b42318;
}
</style>
