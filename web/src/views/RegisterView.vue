<template>
  <section class="panel stack">
    <div>
      <p class="tag">新用户注册</p>
      <h2>创建你的 pveCloud 账户</h2>
    </div>
    <form class="grid">
      <label>
        <span>手机号</span>
        <input type="tel" placeholder="13800000000" />
      </label>
      <label>
        <span>邮箱</span>
        <input type="email" placeholder="name@example.com" />
      </label>
      <label>
        <span>密码</span>
        <input type="password" placeholder="设置登录密码" />
      </label>
      <button type="button">提交注册</button>
    </form>
  </section>
</template>

<style scoped>
.panel {
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(29, 42, 51, 0.08);
  padding: 28px;
}

.stack,
.grid {
  display: grid;
  gap: 18px;
}

.tag {
  margin: 0 0 8px;
  font-size: 13px;
  letter-spacing: 0.18em;
  color: #9b5d32;
}

h2 {
  margin: 0;
}

label {
  display: grid;
  gap: 8px;
}

input,
button {
  border-radius: 16px;
  padding: 14px 16px;
  font: inherit;
}

input {
  border: 1px solid #c8d7e1;
  background: #f9fbfc;
}

button {
  border: 0;
  background: #1d2a33;
  color: #fff;
}
</style>
