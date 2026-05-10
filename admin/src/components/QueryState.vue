<script setup lang="ts">
import { NButton, NResult, NSkeleton } from 'naive-ui'

defineProps<{
  loading?: boolean
  errorMessage?: string
}>()

defineEmits<{
  retry: []
}>()
</script>

<template>
  <div v-if="loading" class="query-state">
    <NSkeleton :repeat="4" text :sharp="false" />
  </div>
  <NResult
    v-else-if="errorMessage"
    class="query-state"
    status="error"
    title="加载失败"
    :description="errorMessage"
  >
    <template #footer>
      <NButton type="primary" @click="$emit('retry')">重新加载</NButton>
    </template>
  </NResult>
  <slot v-else />
</template>

<style scoped>
.query-state {
  min-height: 240px;
  display: grid;
  align-items: center;
}
</style>
