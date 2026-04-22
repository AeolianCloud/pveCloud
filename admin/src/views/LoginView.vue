<template>
  <section class="panel login">
    <div>
      <p class="tag">管理员认证</p>
      <h2>进入平台运营控制台</h2>
    </div>
    <form class="form">
      <label>
        <span>用户名</span>
        <input type="text" placeholder="admin" />
      </label>
      <label>
        <span>密码</span>
        <input type="password" placeholder="请输入密码" />
      </label>
      <button type="button">登录后台</button>
    </form>
  </section>
</template>

<style scoped>
.panel {
  padding: 28px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.login {
  display: grid;
  grid-template-columns: 1fr 0.9fr;
  gap: 24px;
}

.tag {
  margin: 0 0 8px;
  color: #557257;
}

.form {
  display: grid;
  gap: 16px;
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
  border: 1px solid #c8d7cf;
}

button {
  border: 0;
  background: #1d3a2f;
  color: #fff;
}

@media (max-width: 800px) {
  .login {
    grid-template-columns: 1fr;
  }
}
</style>
