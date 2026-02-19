<template>
  <section class="grid">
    <article class="panel">
      <h2>钱包</h2>
      <p>可用余额：¥{{ walletStore.balance }}</p>
      <p>冻结金额：¥{{ walletStore.frozenBalance }}</p>
      <BaseButton @click="dialog = true">充值</BaseButton>
    </article>

    <article class="panel" v-if="dialog">
      <h3>充值对话框</h3>
      <form style="display: flex; gap: 8px;" @submit.prevent="recharge">
        <BaseInput
          :model-value="String(amount)"
          type="number"
          min="10"
          placeholder="充值金额"
          :error="amountError"
          @update:model-value="onAmountChange"
        />
        <BaseButton type="submit">确认充值</BaseButton>
        <BaseButton type="button" @click="dialog = false">取消</BaseButton>
      </form>
    </article>

    <article class="panel">
      <h3>流水列表</h3>
      <ul>
        <li v-for="log in logs" :key="log.id">{{ log.type }} {{ log.amount }} {{ log.created_at }}</li>
      </ul>
    </article>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import http from '../../api/http';
import { useWalletStore } from '../../stores/wallet';
import BaseButton from '../../components/ui/BaseButton.vue';
import BaseInput from '../../components/ui/BaseInput.vue';
import { minNumber, runValidators } from '../../utils/validators';

const walletStore = useWalletStore();
const dialog = ref(false);
const amount = ref(100);
const amountError = ref('');
const logs = ref<any[]>([]);

async function load(): Promise<void> {
  await walletStore.fetchWallet();
  const res = await http.get('/user/wallet/logs');
  logs.value = res.data.data ?? [];
}

function onAmountChange(value: string): void {
  amount.value = Number(value);
}

async function recharge(): Promise<void> {
  amountError.value = runValidators(String(amount.value), [minNumber('充值金额', 10)]);
  if (amountError.value) {
    return;
  }
  await http.post('/user/wallet/recharge', { amount: amount.value });
  dialog.value = false;
  await load();
}

onMounted(load);
</script>
