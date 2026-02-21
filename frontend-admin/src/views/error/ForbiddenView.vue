<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <div class="forbidden-page">
    <div class="forbidden-content">
      <div class="forbidden-code">403</div>
      <div class="forbidden-title">无权限访问</div>
      <div class="forbidden-desc">您没有访问该页面的权限，请联系管理员。</div>
      <n-button type="primary" @click="router.push('/dashboard')">
        返回首页
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.forbidden-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f7f8fa;
}

.forbidden-content {
  text-align: center;
}

.forbidden-code {
  font-size: 96px;
  font-weight: 700;
  color: #d1d5db;
  line-height: 1;
  margin-bottom: 16px;
}

.forbidden-title {
  font-size: 22px;
  font-weight: 600;
  color: #18181c;
  margin-bottom: 8px;
}

.forbidden-desc {
  font-size: 14px;
  color: #8c8c8c;
  margin-bottom: 28px;
}
</style>
