<template>
  <main class="container grid" style="padding: 16px 0;">
    <section class="panel">
      <h1>产品列表</h1>
      <p>按地域筛选并对比套餐规格。</p>
    </section>

    <section class="panel">
      <table style="width: 100%; border-collapse: collapse;">
        <thead>
          <tr>
            <th align="left">套餐</th>
            <th align="left">规格</th>
            <th align="left">最低价</th>
            <th align="left">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>{{ item.name }}</td>
            <td>{{ item.cpu }}C / {{ item.memory_gb }}G / {{ item.disk_gb }}G</td>
            <td>¥{{ item.minPrice }}</td>
            <td><RouterLink :to="`/products/${item.id}`">查看详情</RouterLink></td>
          </tr>
        </tbody>
      </table>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../../api/http';

const items = ref<Array<{ id: number; name: string; cpu: number; memory_gb: number; disk_gb: number; minPrice: number }>>([]);

onMounted(async () => {
  const res = await http.get('/pub/products');
  items.value = (res.data.data ?? []).map((row: any) => {
    const prices = row.prices ?? [];
    const minPrice = prices.length ? Math.min(...prices.map((p: any) => Number(p.unit_price))) : 0;
    return { ...row.product, minPrice };
  });
});
</script>
