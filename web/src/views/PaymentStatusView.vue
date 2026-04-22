<template>
  <section class="panel">
    <p class="tag">支付状态</p>
    <h2>支付单号 {{ route.params.paymentOrderNo }}</h2>
    <p>当前状态：处理中。支付回调完成后，这里会展示成功或失败结果。</p>
  </section>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()
</script>

<style scoped>
.panel {
  padding: 28px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(29, 42, 51, 0.08);
}

.tag {
  margin: 0 0 8px;
  color: #9b5d32;
}
</style>
