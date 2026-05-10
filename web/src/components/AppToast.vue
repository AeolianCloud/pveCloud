<script setup lang="ts">
import { useToast } from '../composables/useToast'

const toast = useToast()

const toneClass = {
  success: 'border-emerald-200 bg-emerald-50 text-emerald-900',
  error: 'border-red-200 bg-red-50 text-red-900',
  info: 'border-neutral-200 bg-white text-neutral-900',
}
</script>

<template>
  <div class="pointer-events-none fixed right-4 top-20 z-[80] flex w-[calc(100vw-2rem)] max-w-sm flex-col gap-3 sm:right-6">
    <TransitionGroup name="toast-slide">
      <div
        v-for="item in toast.toasts"
        :key="item.id"
        :class="['pointer-events-auto rounded-2xl border px-4 py-3 text-sm font-bold shadow-[6px_6px_0_rgba(0,0,0,0.14)]', toneClass[item.tone]]"
      >
        <div class="flex items-start justify-between gap-3">
          <span>{{ item.message }}</span>
          <button type="button" class="text-lg leading-none opacity-60 hover:opacity-100" aria-label="关闭提示" @click="toast.remove(item.id)">×</button>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-slide-enter-active,
.toast-slide-leave-active {
  transition: all 0.2s ease;
}

.toast-slide-enter-from,
.toast-slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
