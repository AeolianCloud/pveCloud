<template>
  <main class="container grid" style="padding: 16px 0;">
    <section class="panel" v-if="detail">
      <h1>{{ detail.product.name }}</h1>
      <p>{{ detail.product.description }}</p>
      <p>规格：{{ detail.product.cpu }}C / {{ detail.product.memory_gb }}G / {{ detail.product.disk_gb }}G</p>

      <label>
        计费周期
        <select class="input" v-model="selectedCycle">
          <option v-for="p in detail.prices" :key="p.billing_cycle" :value="p.billing_cycle">
            {{ p.billing_cycle }} - ¥{{ p.unit_price }}
          </option>
        </select>
      </label>

      <p style="margin-top: 12px;">价格计算器：当前总价 <strong>¥{{ selectedPrice }}</strong></p>
      <RouterLink class="btn" to="/console/order-flow">去下单</RouterLink>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import http from '../../api/http';

const route = useRoute();
const detail = ref<any>(null);
const selectedCycle = ref('monthly');

const selectedPrice = computed(() => {
  if (!detail.value) return 0;
  const hit = (detail.value.prices ?? []).find((p: any) => p.billing_cycle === selectedCycle.value);
  return hit?.unit_price ?? 0;
});

onMounted(async () => {
  const res = await http.get(`/pub/products/${route.params.id}`);
  detail.value = res.data.data;
  selectedCycle.value = detail.value.prices?.[0]?.billing_cycle ?? 'monthly';
});
</script>
