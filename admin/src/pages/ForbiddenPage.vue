<template>
  <section class="forbidden-page">
    <div class="forbidden-panel">
      <span>403</span>
      <h1>无权访问</h1>
      <p>当前账号没有访问该页面的权限，请联系系统管理员调整角色或返回控制台。</p>
      <RouterLink class="primary-button" to="/dashboard">返回控制台</RouterLink>
    </div>
  </section>
</template>

<style scoped>
.forbidden-page {
  min-height: calc(100vh - 108px);
  display: grid;
  place-items: center;
}

.forbidden-panel {
  width: min(100%, 380px);
  display: grid;
  justify-items: center;
  gap: 10px;
  padding: 30px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow);
  text-align: center;
}

.forbidden-panel span {
  color: var(--primary);
  font-size: 42px;
  font-weight: 850;
  line-height: 1;
}

.forbidden-panel h1 {
  margin: 0;
  font-size: 22px;
}

.forbidden-panel p {
  margin: 0 0 8px;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.7;
}

.forbidden-panel a {
  text-decoration: none;
}
</style>
