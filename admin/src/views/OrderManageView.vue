<template>
  <section class="panel">
    <p class="tag">订单管理</p>
    <h2>订单状态</h2>
    <table>
      <thead>
        <tr>
          <th>订单号</th>
          <th>用户</th>
          <th>状态</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>O202604220001</td>
          <td>1001</td>
          <td>pending_payment</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
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

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 12px 10px;
  text-align: left;
  border-bottom: 1px solid #d9e3dc;
}
</style>
