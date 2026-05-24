<script setup lang="ts">
import QRCode from 'qrcode'
import { ref, watch } from 'vue'

const props = defineProps<{
  value: string
  alt?: string
}>()

const imageUrl = ref('')
const errorMessage = ref('')

watch(
  () => props.value,
  async (value) => {
    imageUrl.value = ''
    errorMessage.value = ''
    const content = value.trim()
    if (!content) return
    try {
      imageUrl.value = await QRCode.toDataURL(content, {
        errorCorrectionLevel: 'M',
        margin: 2,
        width: 224,
      })
    } catch {
      errorMessage.value = '二维码生成失败'
    }
  },
  { immediate: true },
)
</script>

<template>
  <div class="qr-code-image">
    <img v-if="imageUrl" :src="imageUrl" :alt="alt || '支付二维码'" class="qr-code-image__img" />
    <div v-else class="qr-code-image__fallback">{{ errorMessage || '二维码生成中...' }}</div>
  </div>
</template>

<style scoped>
.qr-code-image {
  width: 14rem;
  max-width: 100%;
  border: 1px solid #e5e5e5;
  border-radius: 1rem;
  background: #fff;
  padding: 0.75rem;
}

.qr-code-image__img {
  display: block;
  width: 100%;
  height: auto;
}

.qr-code-image__fallback {
  display: flex;
  aspect-ratio: 1 / 1;
  align-items: center;
  justify-content: center;
  color: #525252;
  font-size: 0.875rem;
  font-weight: 700;
}
</style>
