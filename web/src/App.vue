<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppConfirmDialog from './components/AppConfirmDialog.vue'
import AppToast from './components/AppToast.vue'
import WebLayout from './layouts/WebLayout.vue'
</script>

<template>
  <WebLayout>
    <RouterView />
  </WebLayout>
  <AppToast />
  <AppConfirmDialog />
</template>
