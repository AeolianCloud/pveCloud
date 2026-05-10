<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { useConfirm } from '../composables/useConfirm'

const confirm = useConfirm()

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && confirm.state.visible) {
    confirm.close(false)
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <Transition name="dialog-fade">
      <div v-if="confirm.state.visible" class="fixed inset-0 z-[90] flex items-center justify-center bg-neutral-950/45 px-4 backdrop-blur-sm" @click.self="confirm.close(false)">
        <section
          class="w-full max-w-md rounded-[1.5rem] border border-neutral-950 bg-white p-6 shadow-[10px_10px_0_#111]"
          role="dialog"
          aria-modal="true"
          :aria-label="confirm.state.title"
        >
          <p class="text-xs font-black uppercase tracking-[0.16em] text-neutral-500">Confirm</p>
          <h2 class="mt-3 text-2xl font-black text-neutral-950">{{ confirm.state.title }}</h2>
          <p class="mt-4 text-sm leading-6 text-neutral-600">{{ confirm.state.message }}</p>
          <div class="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
            <button type="button" class="action-pill border border-neutral-300 px-5 py-2 text-sm font-black text-neutral-700 hover:border-neutral-950" @click="confirm.close(false)">
              {{ confirm.state.cancelText }}
            </button>
            <button
              type="button"
              :class="[
                'action-pill border px-5 py-2 text-sm font-black text-white',
                confirm.state.tone === 'danger' ? 'border-red-600 bg-red-600 hover:bg-red-700' : 'border-neutral-950 bg-neutral-950 hover:bg-neutral-800',
              ]"
              @click="confirm.close(true)"
            >
              {{ confirm.state.confirmText }}
            </button>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.dialog-fade-enter-active,
.dialog-fade-leave-active {
  transition: opacity 0.18s ease;
}

.dialog-fade-enter-from,
.dialog-fade-leave-to {
  opacity: 0;
}
</style>
