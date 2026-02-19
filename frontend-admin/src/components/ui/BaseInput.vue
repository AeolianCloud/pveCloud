<template>
  <label style="display: grid; gap: 6px;">
    <span v-if="label">{{ label }}</span>
    <input
      class="input"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :min="min"
      @input="onInput"
    />
    <small v-if="error" style="color: #a00;">{{ error }}</small>
  </label>
</template>

<script setup lang="ts">
interface Props {
  modelValue: string | number;
  type?: string;
  label?: string;
  placeholder?: string;
  min?: string | number;
  error?: string;
}

withDefaults(defineProps<Props>(), {
  type: 'text',
  label: '',
  placeholder: '',
  min: '',
  error: '',
});

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>();

function onInput(event: Event): void {
  const target = event.target as HTMLInputElement;
  emit('update:modelValue', target.value);
}
</script>
