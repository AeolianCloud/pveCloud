<script setup lang="ts">
import { ADMIN_ROUTE_PATH } from '../../router/constants'
</script>

<template>
  <div class="forbidden-page">
    <el-result
      icon="warning"
      title="无权访问"
      sub-title="当前账号没有访问这个页面的权限，请联系管理员。"
    >
      <template #extra>
        <router-link :to="ADMIN_ROUTE_PATH.dashboard">
          <el-button type="primary" size="large">返回控制台</el-button>
        </router-link>
      </template>
    </el-result>
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
