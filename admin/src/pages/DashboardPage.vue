<script setup lang="ts">
import { LogOut, ShieldCheck } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <main class="dashboard-page">
    <header class="dashboard-topbar">
      <div>
        <span class="eyebrow">pveCloud Admin</span>
        <h1>管理后台</h1>
      </div>
      <button class="icon-button" type="button" title="退出登录" @click="logout">
        <LogOut :size="18" aria-hidden="true" />
        <span>退出</span>
      </button>
    </header>

    <section class="dashboard-main">
      <div class="summary-panel">
        <ShieldCheck :size="28" aria-hidden="true" />
        <div>
          <h2>{{ auth.admin?.display_name || auth.admin?.username }}</h2>
          <p>{{ auth.admin?.username }} 已登录</p>
        </div>
      </div>

      <div class="permission-panel">
        <h2>权限码</h2>
        <div class="permission-list">
          <span v-for="code in auth.permissionCodes" :key="code">{{ code }}</span>
          <span v-if="auth.permissionCodes.length === 0">暂无权限</span>
        </div>
      </div>
    </section>
  </main>
</template>
