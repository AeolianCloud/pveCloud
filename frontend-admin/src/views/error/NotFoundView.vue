<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <div class="not-found-page">
    <div class="not-found-content">
      <div class="not-found-code">404</div>
      <div class="not-found-title">页面不存在</div>
      <div class="not-found-desc">您访问的页面不存在，请检查地址是否正确。</div>
      <n-button type="primary" @click="router.push('/dashboard')">
        返回首页
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.not-found-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f7f8fa;
}

.not-found-content {
  text-align: center;
}

.not-found-code {
  font-size: 96px;
  font-weight: 700;
  color: #d1d5db;
  line-height: 1;
  margin-bottom: 16px;
}

.not-found-title {
  font-size: 22px;
  font-weight: 600;
  color: #18181c;
  margin-bottom: 8px;
}

.not-found-desc {
  font-size: 14px;
  color: #8c8c8c;
  margin-bottom: 28px;
}
</style>
