<template>
  <section class="panel hero">
    <div class="copy">
      <p class="tag">USER LOGIN</p>
      <h2>登录并进入用户侧控制台</h2>
      <p class="lead">这里接的是后端真实登录接口。登录成功后可以继续查看订单、支付状态和实例。</p>
    </div>

    <form class="form" @submit.prevent="submit">
      <label>
        <span>手机号</span>
        <input v-model="phone" name="phone" type="tel" placeholder="13800000000" />
      </label>
      <label>
        <span>密码</span>
        <input v-model="password" name="password" type="password" placeholder="请输入密码" />
      </label>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      <button type="submit" :disabled="submitting">{{ submitting ? '登录中...' : '立即登录' }}</button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const phone = ref('')
const password = ref('')
const errorMessage = ref('')
const submitting = ref(false)
const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

async function submit() {
  errorMessage.value = ''
  submitting.value = true

  try {
    await authStore.login(phone.value, password.value)
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
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(29, 42, 51, 0.08);
  padding: 28px;
  box-shadow: 0 18px 50px rgba(54, 76, 90, 0.1);
}

.hero {
  display: grid;
  grid-template-columns: 1.2fr 0.9fr;
  gap: 24px;
}

.copy {
  display: grid;
  align-content: start;
  gap: 10px;
}

.tag {
  margin: 0;
  color: #9b5d32;
  font-size: 13px;
  letter-spacing: 0.18em;
}

h2 {
  margin: 0;
  font-size: clamp(26px, 4vw, 42px);
}

.lead {
  margin: 0;
  color: #4d606e;
  line-height: 1.7;
}

.form {
  display: grid;
  gap: 16px;
}

label {
  display: grid;
  gap: 8px;
  font-weight: 600;
}

input {
  width: 100%;
  border: 1px solid #c8d7e1;
  border-radius: 18px;
  padding: 14px 16px;
  font: inherit;
  background: #f9fbfc;
}

button {
  border: 0;
  border-radius: 18px;
  padding: 14px 18px;
  font: inherit;
  font-weight: 700;
  background: linear-gradient(135deg, #cb6230, #7e4025);
  color: white;
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.72;
}

.error {
  margin: 0;
  color: #b42318;
  font-size: 14px;
}

@media (max-width: 800px) {
  .hero {
    grid-template-columns: 1fr;
  }
}
</style>
