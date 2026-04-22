<template>
  <section class="dashboard">
    <header class="panel intro">
      <p class="tag">运营总览</p>
      <h2>管理仪表盘</h2>
      <p>聚合订单、实例和异步任务的最小运营视图，供 MVP 阶段值守使用。</p>
    </header>

    <div class="stats">
      <article class="panel stat">
        <span>今日订单</span>
        <strong>18</strong>
      </article>
      <article class="panel stat">
        <span>运行实例</span>
        <strong>126</strong>
      </article>
      <article class="panel stat">
        <span>待处理任务</span>
        <strong>4</strong>
      </article>
    </div>
  </section>
</template>

<style scoped>
.dashboard,
.stats {
  display: grid;
  gap: 18px;
}

.stats {
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.panel {
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.tag {
  margin: 0 0 8px;
  color: #557257;
}

h2 {
  margin: 0 0 10px;
}

.stat span {
  display: block;
  color: #5f6b65;
}

.stat strong {
  display: block;
  margin-top: 12px;
  font-size: 38px;
}
</style>
