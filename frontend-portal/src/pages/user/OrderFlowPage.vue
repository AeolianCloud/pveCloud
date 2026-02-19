<template>
  <section class="panel grid">
    <h2>下单流程</h2>
    <p>步骤：选配置 → 确认订单 → 等待开通</p>

    <div class="panel">
      <h3>步骤 1：选择配置</h3>
      <BaseInput
        :model-value="String(form.product_id)"
        type="number"
        label="Product ID"
        :error="errors.product_id"
        @update:model-value="(v) => (form.product_id = Number(v))"
      />
      <label>
        Billing Cycle
        <select class="input" v-model="form.billing_cycle">
          <option value="hourly">hourly</option>
          <option value="monthly">monthly</option>
          <option value="annual">annual</option>
        </select>
      </label>
    </div>

    <div class="panel">
      <h3>步骤 2：确认并提交</h3>
      <BaseButton @click="submit">提交订单</BaseButton>
      <p v-if="taskId">任务ID：{{ taskId }}</p>
    </div>

    <div class="panel" v-if="taskStatus">
      <h3>步骤 3：开通进度</h3>
      <progress :value="taskStatus.progress" max="100" style="width: 100%;"></progress>
      <p>{{ taskStatus.status }} - {{ taskStatus.message }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import http from '../../api/http';
import BaseInput from '../../components/ui/BaseInput.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import { minNumber, runValidators } from '../../utils/validators';

const form = reactive({ product_id: 1, billing_cycle: 'monthly', os: 'ubuntu-22.04', cpu: 2, memory_gb: 4, disk_gb: 60 });
const errors = ref({ product_id: '' });
const taskId = ref<number | null>(null);
const taskStatus = ref<any>(null);

async function submit(): Promise<void> {
  errors.value.product_id = runValidators(String(form.product_id), [minNumber('Product ID', 1)]);
  if (errors.value.product_id) {
    return;
  }

  const res = await http.post('/user/orders', form);
  taskId.value = res.data.data.task_id;
  poll();
}

async function poll(): Promise<void> {
  if (!taskId.value) return;
  const timer = window.setInterval(async () => {
    const res = await http.get(`/user/tasks/${taskId.value}/status`);
    taskStatus.value = res.data.data;
    if (taskStatus.value.status === 'success' || taskStatus.value.status === 'failed') {
      clearInterval(timer);
    }
  }, 3000);
}
</script>
