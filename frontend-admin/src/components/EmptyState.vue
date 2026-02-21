<script setup lang="ts">
// 通用空状态组件：表格无数据时显示。
// 用法：<EmptyState /> 或 <EmptyState description="暂无日志" />
withDefaults(defineProps<{
  description?: string
}>(), {
  description: '暂无数据',
})
</script>

<template>
  <div class="empty-state">
    <!-- 简约线条插图 -->
    <svg class="empty-icon" viewBox="0 0 120 80" fill="none" xmlns="http://www.w3.org/2000/svg">
      <!-- 文档轮廓 -->
      <rect x="30" y="10" width="60" height="60" rx="6" fill="#f0f2f5" stroke="#e0e3ea" stroke-width="1.5"/>
      <!-- 折角 -->
      <path d="M72 10 L90 28" stroke="#e0e3ea" stroke-width="1.5"/>
      <path d="M72 10 L72 28 L90 28" fill="#e8eaf0" stroke="#e0e3ea" stroke-width="1.5"/>
      <!-- 内容线条 -->
      <line x1="44" y1="40" x2="76" y2="40" stroke="#d0d4de" stroke-width="2" stroke-linecap="round"/>
      <line x1="44" y1="50" x2="68" y2="50" stroke="#d0d4de" stroke-width="2" stroke-linecap="round"/>
    </svg>
    <p class="empty-desc">{{ description }}</p>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 0 40px;
  gap: 12px;
}

.empty-icon {
  width: 100px;
  height: 68px;
  opacity: 0.85;
}

.empty-desc {
  font-size: 13px;
  color: #b0b5bf;
}
</style>
