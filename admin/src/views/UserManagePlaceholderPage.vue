<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listUsers, listAdmins, type UserRow, type AdminRow } from '../api/user'

const users = ref<UserRow[]>([])
const admins = ref<AdminRow[]>([])
const loading = ref(true)
const error = ref('')
const tab = ref<'users' | 'admins'>('users')

onMounted(async () => {
  try {
    const [u, a] = await Promise.all([listUsers(), listAdmins()])
    users.value = u
    admins.value = a
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="panel">
    <p class="tag">USERS</p>
    <h2>用户管理</h2>

    <div class="tabs">
      <button :class="{ active: tab === 'users' }" @click="tab = 'users'">注册用户</button>
      <button :class="{ active: tab === 'admins' }" @click="tab = 'admins'">管理员</button>
    </div>

    <p v-if="loading">加载中...</p>
    <p v-else-if="error" class="error">{{ error }}</p>

    <template v-else-if="tab === 'users'">
      <p v-if="users.length === 0">暂无用户</p>
      <table v-else>
        <thead><tr><th>ID</th><th>编号</th><th>手机</th><th>邮箱</th><th>状态</th><th>注册时间</th></tr></thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.id }}</td><td>{{ u.user_no }}</td><td>{{ u.phone }}</td><td>{{ u.email || '-' }}</td><td>{{ u.status }}</td><td>{{ u.created_at }}</td>
          </tr>
        </tbody>
      </table>
    </template>

    <template v-else>
      <p v-if="admins.length === 0">暂无管理员</p>
      <table v-else>
        <thead><tr><th>ID</th><th>编号</th><th>用户名</th><th>状态</th><th>创建时间</th></tr></thead>
        <tbody>
          <tr v-for="a in admins" :key="a.id">
            <td>{{ a.id }}</td><td>{{ a.admin_no }}</td><td>{{ a.username }}</td><td>{{ a.status }}</td><td>{{ a.created_at }}</td>
          </tr>
        </tbody>
      </table>
    </template>
  </section>
</template>

<style scoped>
.panel {
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.tag { margin: 0 0 8px; color: #557257; }
h2 { margin: 0 0 16px; }

.tabs { margin-bottom: 16px; }
.tabs button {
  padding: 6px 16px;
  margin-right: 8px;
  border: 1px solid #557257;
  border-radius: 6px;
  background: transparent;
  color: #557257;
  cursor: pointer;
}
.tabs button.active {
  background: #557257;
  color: #fff;
}

table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px 12px; text-align: left; border-bottom: 1px solid rgba(0,0,0,0.06); }
th { background: rgba(85, 114, 87, 0.08); }

.error { color: #c00; }
</style>
