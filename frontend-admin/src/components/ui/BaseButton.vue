<template>
  <button class="btn" :type="type" :disabled="disabled" @click="onClick">
    <slot />
  </button>
</template>

<script setup lang="ts">
interface Props {
  type?: 'button' | 'submit' | 'reset';
  disabled?: boolean;
}

withDefaults(defineProps<Props>(), {
  type: 'button',
  disabled: false,
});

const emit = defineEmits<{ (e: 'click', event: MouseEvent): void }>();

function onClick(event: MouseEvent): void {
  emit('click', event);
}
</script>
