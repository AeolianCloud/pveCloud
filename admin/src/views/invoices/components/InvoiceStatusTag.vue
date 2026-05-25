<script setup lang="ts">
import { NTag } from 'naive-ui'

import type { InvoiceStatus } from '../../../api/invoice'

defineProps<{
  status: InvoiceStatus | string
}>()

const statusText: Record<string, string> = {
  pending: '待处理',
  processing: '处理中',
  issued: '已开票',
  rejected: '已驳回',
  cancelled: '已取消',
}

const statusType: Record<string, 'default' | 'info' | 'success' | 'warning' | 'error'> = {
  pending: 'warning',
  processing: 'info',
  issued: 'success',
  rejected: 'error',
  cancelled: 'default',
}
</script>

<template>
  <NTag size="small" :type="statusType[status] || 'default'">
    {{ statusText[status] || status }}
  </NTag>
</template>
