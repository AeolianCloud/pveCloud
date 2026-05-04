<template>
  <section class="not-found-page">
    <div class="not-found-card">
      <p class="label">404</p>
      <h1>页面不存在</h1>
      <p>你访问的路径没有匹配到当前路由。</p>
      <RouterLink to="/" class="btn btn-primary">返回首页</RouterLink>
    </div>
  </section>
</template>

<style scoped>
.not-found-page {
  min-height: calc(100vh - 96px);
  display: grid;
  place-items: center;
  padding: 32px 20px 56px;
}

.not-found-card {
  width: min(100%, 560px);
  display: grid;
  gap: 18px;
  justify-items: start;
  padding: clamp(30px, 5vw, 52px);
  border: 1px solid var(--c-border);
  border-radius: 30px;
  background:
    radial-gradient(circle at 100% 0%, rgba(59, 130, 246, 0.16), transparent 40%),
    var(--c-card);
  box-shadow: var(--shadow-lg);
}

.label {
  color: var(--c-primary);
  font-size: 0.82rem;
  font-weight: 800;
  letter-spacing: 0.14em;
}

.not-found-card h1 {
  font-size: clamp(2.25rem, 4vw, 3.6rem);
  line-height: 1;
  letter-spacing: -0.06em;
}

.not-found-card p {
  color: var(--c-text-2);
  line-height: 1.8;
}

@media (max-width: 620px) {
  .not-found-page {
    min-height: calc(100vh - 72px);
  }
}
</style>
