<template>
  <section class="forbidden-page">
    <Card class="forbidden-panel">
      <template #content>
        <span>403</span>
        <h1>无权访问</h1>
        <Message severity="warn" :closable="false">
          当前账号没有访问这个页面的权限，请联系系统管理员调整角色权限。
        </Message>
        <RouterLink to="/dashboard">
          <Button label="返回控制台" icon="pi pi-arrow-left" />
        </RouterLink>
      </template>
    </Card>
  </section>
</template>

<style scoped>
.forbidden-page {
  min-height: calc(100vh - 108px);
  display: grid;
  place-items: center;
}

.forbidden-panel {
  width: min(100%, 420px);
  text-align: center;
}

.forbidden-panel :deep(.p-card-content) {
  display: grid;
  justify-items: center;
  gap: 14px;
}

.forbidden-panel span {
  color: var(--primary);
  font-size: 42px;
  font-weight: 850;
  line-height: 1;
}

.forbidden-panel h1 {
  margin: 0;
  color: var(--text);
  font-size: 22px;
}

.forbidden-panel a {
  text-decoration: none;
}
</style>
