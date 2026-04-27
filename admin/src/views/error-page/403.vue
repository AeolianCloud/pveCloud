<script setup lang="ts">
import { ADMIN_ROUTE_PATH } from '../../router/constants'
</script>

<template>
  <section class="forbidden-page">
    <div class="forbidden-page__halo forbidden-page__halo--left"></div>
    <div class="forbidden-page__halo forbidden-page__halo--right"></div>

    <div class="forbidden-page__panel">
      <span class="forbidden-page__eyebrow">Permission Gate</span>
      <el-result
        icon="warning"
        title="403 无权访问"
        sub-title="当前账号没有访问这个页面的权限，请联系系统管理员调整角色权限。"
      >
        <template #extra>
          <router-link :to="ADMIN_ROUTE_PATH.dashboard">
            <el-button type="primary" size="large">返回控制台</el-button>
          </router-link>
        </template>
      </el-result>
    </div>
  </section>
</template>

<style scoped>
.forbidden-page {
  min-height: calc(100vh - 140px);
  position: relative;
  display: grid;
  place-items: center;
  overflow: hidden;
}

.forbidden-page__halo {
  position: absolute;
  width: 280px;
  aspect-ratio: 1;
  border-radius: 999px;
  filter: blur(12px);
  opacity: 0.48;
}

.forbidden-page__halo--left {
  top: 10%;
  left: 6%;
  background: rgba(14, 165, 233, 0.18);
}

.forbidden-page__halo--right {
  right: 8%;
  bottom: 10%;
  background: rgba(245, 158, 11, 0.16);
}

.forbidden-page__panel {
  position: relative;
  z-index: 1;
  width: min(640px, 100%);
  padding: 28px;
  border-radius: var(--pc-card-radius);
  background: var(--pc-shell-bg-strong);
  box-shadow: var(--pc-card-shadow);
  backdrop-filter: blur(18px);
}

.forbidden-page__eyebrow {
  display: inline-flex;
  margin-bottom: 10px;
  padding: 6px 12px;
  border-radius: 999px;
  color: #9a3412;
  background: rgba(251, 191, 36, 0.18);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
</style>
