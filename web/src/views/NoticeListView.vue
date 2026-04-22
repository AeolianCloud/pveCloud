<template>
  <section class="panel">
    <p class="tag">站内通知</p>
    <h2>近期公告</h2>
    <ul>
      <li>资源池维护窗口：2026-04-25 02:00 - 04:00</li>
      <li>华北区域新增高频机型 SKU</li>
      <li>支付回调链路已升级为异步任务处理</li>
    </ul>
  </section>
</template>

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

ul {
  margin: 0;
  padding-left: 20px;
  display: grid;
  gap: 10px;
}
</style>
