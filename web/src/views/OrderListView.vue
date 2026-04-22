<template>
  <section class="panel">
    <p class="tag">订单列表</p>
    <h2>最近订单</h2>
    <table>
      <thead>
        <tr>
          <th>订单号</th>
          <th>产品</th>
          <th>状态</th>
          <th>应付金额</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>O202604220001</td>
          <td>通用型云主机</td>
          <td>待支付</td>
          <td>100.00 元</td>
        </tr>
      </tbody>
    </table>
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

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 14px 10px;
  text-align: left;
  border-bottom: 1px solid #e1eaef;
}
</style>
