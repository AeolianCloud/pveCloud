<script setup lang="ts">
import { NButton, NResult } from 'naive-ui'
import { useRouter } from 'vue-router'

import { ADMIN_ROUTE_PATH } from '../../router/constants'

const router = useRouter()
function goDashboard() {
  void router.push(ADMIN_ROUTE_PATH.dashboard)
}
</script>

<template>
  <div class="forbidden-page">
    <NResult
      status="403"
      title="无权访问"
      description="当前账号没有访问这个页面的权限，请联系管理员。"
    >
      <template #footer>
        <NButton type="primary" size="large" @click="goDashboard">返回控制台</NButton>
      </template>
    </NResult>
  </div>
</template>

<style scoped>
.forbidden-page {
  min-height: calc(100vh - 140px);
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
