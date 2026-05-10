<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

export interface AppSelectOption {
  label: string
  value: string
  description?: string | null
}

const props = withDefaults(defineProps<{
  options: AppSelectOption[]
  placeholder?: string
}>(), {
  placeholder: '请选择',
})

const model = defineModel<string>({ required: true })

const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

const selectedOption = computed(() =>
  props.options.find((option) => option.value === model.value) || null,
)

function selectOption(option: AppSelectOption) {
  model.value = option.value
  open.value = false
}

function onDocumentClick(event: MouseEvent) {
  if (!rootRef.value?.contains(event.target as Node)) {
    open.value = false
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    open.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div ref="rootRef" class="relative">
    <button
      type="button"
    class="group flex w-full cursor-pointer items-center justify-between gap-4 rounded-2xl border border-neutral-300 bg-white px-4 py-3 text-left transition hover:border-neutral-950 hover:bg-neutral-50 focus:border-neutral-950 focus:outline-none focus:ring-4 focus:ring-neutral-200"
      :aria-expanded="open"
      @click="open = !open"
    >
      <span class="min-w-0">
        <span class="block truncate text-sm font-black text-neutral-950">
          {{ selectedOption?.label || props.placeholder }}
        </span>
        <span v-if="selectedOption?.description" class="mt-1 block truncate text-xs text-neutral-500">
          {{ selectedOption.description }}
        </span>
      </span>
      <span :class="['shrink-0 text-neutral-500 transition-transform', open ? 'rotate-180' : '']">⌄</span>
    </button>

    <Transition name="select-pop">
      <div
        v-if="open"
        class="absolute left-0 right-0 top-[calc(100%+0.5rem)] z-50 overflow-hidden rounded-2xl border border-neutral-200 bg-white shadow-lg shadow-neutral-200/70"
      >
        <button
          v-for="option in props.options"
          :key="option.value"
          type="button"
          :class="[
            'app-select__option block w-full cursor-pointer px-4 py-3 text-left transition',
            model === option.value ? 'app-select__option--selected text-neutral-950' : 'text-neutral-950',
          ]"
          @click="selectOption(option)"
        >
          <span class="block text-sm font-black">{{ option.label }}</span>
          <span v-if="option.description" class="mt-1 block text-xs leading-5 text-neutral-500">
            {{ option.description }}
          </span>
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.select-pop-enter-active,
.select-pop-leave-active {
  transition: opacity 0.14s ease, transform 0.14s ease;
}

.select-pop-enter-from,
.select-pop-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.app-select__option:hover {
  background: #f5f5f5;
}

.app-select__option--selected {
  background: #111111;
  color: #ffffff;
}

.app-select__option--selected span {
  color: inherit;
}

.app-select__option--selected:hover {
  background: #262626;
}
</style>
