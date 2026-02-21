<script setup lang="ts">
// components/ErrorPage.vue
// 通用错误页组件，供 403、404 等错误页复用。
//
// 使用示例：
//   <ErrorPage code="403" title="无权限访问" description="您没有访问该页面的权限，请联系管理员。" />
import { useRouter } from 'vue-router'

defineProps<{
  code: string
  title: string
  description: string
}>()

const router = useRouter()
</script>

<template>
  <div class="error-page">
    <div class="error-content">
      <div class="error-code">{{ code }}</div>
      <div class="error-title">{{ title }}</div>
      <div class="error-desc">{{ description }}</div>
      <n-button type="primary" @click="router.push('/dashboard')">
        返回首页
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.error-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f7f8fa;
}

.error-content {
  text-align: center;
}

.error-code {
  font-size: 96px;
  font-weight: 700;
  color: #d1d5db;
  line-height: 1;
  margin-bottom: 16px;
}

.error-title {
  font-size: 22px;
  font-weight: 600;
  color: #18181c;
  margin-bottom: 8px;
}

.error-desc {
  font-size: 14px;
  color: #8c8c8c;
  margin-bottom: 28px;
}
</style>
