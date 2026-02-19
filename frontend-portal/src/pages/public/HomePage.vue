<template>
  <header class="panel" style="margin-bottom: 16px;">
    <div class="container" style="display: flex; justify-content: space-between; align-items: center; gap: 16px;">
      <strong>IDC 云门户</strong>
      <nav style="display: flex; gap: 12px;">
        <RouterLink to="/">首页</RouterLink>
        <RouterLink to="/products">产品</RouterLink>
        <RouterLink to="/login">登录</RouterLink>
      </nav>
    </div>
  </header>

  <main class="container grid">
    <section class="panel">
      <h1>高可用云服务器</h1>
      <p>按需购买，分钟级开通，面向业务稳定性设计。</p>
      <RouterLink class="btn" to="/products">查看套餐</RouterLink>
    </section>

    <section class="grid" style="grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));">
      <article class="panel" v-for="card in cards" :key="card.title">
        <h3>{{ card.title }}</h3>
        <p>{{ card.desc }}</p>
      </article>
    </section>
  </main>
</template>

<script setup lang="ts">
const cards = [
  { title: '弹性扩展', desc: 'CPU/内存/磁盘可按业务阶段灵活升级。' },
  { title: '可视化控制台', desc: '实例、账单、工单一站式管理。' },
  { title: '多周期定价', desc: '支持按小时与包年包月等多种周期。' },
];
</script>
