<script setup lang="ts">
import { useDialog, useMessage, useNotification } from 'naive-ui'
import { onMounted } from 'vue'

import { registerNaiveApis } from '../utils/feedback'

const message = useMessage()
const dialog = useDialog()
const notification = useNotification()

onMounted(() => {
  registerNaiveApis({ message, dialog, notification })
})
</script>

<template>
  <slot />
</template>
