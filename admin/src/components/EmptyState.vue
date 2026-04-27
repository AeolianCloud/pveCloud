<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <el-empty :description="description || title">
    <template #description>
      <div class="empty-state">
        <strong>{{ title }}</strong>
        <p v-if="description">{{ description }}</p>
      </div>
    </template>
  </el-empty>
</template>

<style scoped>
.empty-state {
  display: grid;
  gap: 8px;
  text-align: center;
}

.empty-state strong {
  color: var(--pc-title-text);
  font-family: var(--pc-display-font);
  font-size: 18px;
}

.empty-state p {
  max-width: 360px;
  margin: 0 auto;
  color: var(--pc-muted-text);
  line-height: 1.6;
}
</style>
