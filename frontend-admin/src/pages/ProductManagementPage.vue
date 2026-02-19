<template>
  <section class="panel" style="display: grid; gap: 12px;">
    <h2>商品管理</h2>

    <form style="display: grid; gap: 8px;" @submit.prevent="create">
      <BaseInput v-model="form.name" placeholder="商品名称" label="商品名称" :error="errors.name" />
      <BaseInput
        :model-value="String(form.cpu)"
        type="number"
        placeholder="CPU"
        label="CPU"
        :error="errors.cpu"
        @update:model-value="(v) => (form.cpu = Number(v))"
      />
      <BaseInput
        :model-value="String(form.memory_gb)"
        type="number"
        placeholder="内存"
        label="内存"
        :error="errors.memory"
        @update:model-value="(v) => (form.memory_gb = Number(v))"
      />
      <BaseInput
        :model-value="String(form.disk_gb)"
        type="number"
        placeholder="磁盘"
        label="磁盘"
        :error="errors.disk"
        @update:model-value="(v) => (form.disk_gb = Number(v))"
      />
      <BaseButton type="submit">创建商品</BaseButton>
    </form>

    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th align="left">ID</th>
          <th align="left">名称</th>
          <th align="left">状态</th>
          <th align="left">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in products" :key="p.id">
          <td>{{ p.id }}</td>
          <td>{{ p.name }}</td>
          <td>{{ p.status }}</td>
          <td>
            <BaseButton @click="toggleStatus(p)">上下架</BaseButton>
            <BaseButton @click="remove(p.id)">删除</BaseButton>
          </td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import http from '../api/http';
import BaseInput from '../components/ui/BaseInput.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import { minNumber, required, runValidators } from '../utils/validators';

const form = reactive({ name: '', cpu: 2, memory_gb: 4, disk_gb: 60 });
const errors = ref({ name: '', cpu: '', memory: '', disk: '' });
const products = ref<any[]>([]);

async function load(): Promise<void> {
  const res = await http.get('/admin/products');
  products.value = res.data.data ?? [];
}

function validate(): boolean {
  errors.value.name = runValidators(form.name, [required('商品名称')]);
  errors.value.cpu = runValidators(String(form.cpu), [minNumber('CPU', 1)]);
  errors.value.memory = runValidators(String(form.memory_gb), [minNumber('内存', 1)]);
  errors.value.disk = runValidators(String(form.disk_gb), [minNumber('磁盘', 20)]);
  return !errors.value.name && !errors.value.cpu && !errors.value.memory && !errors.value.disk;
}

async function create(): Promise<void> {
  if (!validate()) {
    return;
  }

  await http.post('/admin/products', {
    product: {
      ...form,
      description: '后台创建套餐',
      region_id: 1,
      bandwidth_mbps: 100,
      disk_type: 'SSD',
      status: 'draft',
    },
    prices: [
      { billing_cycle: 'monthly', unit_price: 99 },
      { billing_cycle: 'annual', unit_price: 999 },
    ],
  });
  await load();
}

async function toggleStatus(p: any): Promise<void> {
  const nextStatus = p.status === 'published' ? 'offline' : 'published';
  await http.put(`/admin/products/${p.id}`, {
    product: { ...p, status: nextStatus },
    prices: [],
  });
  await load();
}

async function remove(id: number): Promise<void> {
  await http.delete(`/admin/products/${id}`);
  await load();
}

onMounted(load);
</script>
