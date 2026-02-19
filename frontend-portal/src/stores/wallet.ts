import { defineStore } from 'pinia';
import { ref } from 'vue';
import http from '../api/http';

export const useWalletStore = defineStore('wallet', () => {
  const balance = ref<number>(0);
  const frozenBalance = ref<number>(0);

  async function fetchWallet(): Promise<void> {
    const res = await http.get('/user/wallet');
    balance.value = Number(res.data.data.balance ?? 0);
    frozenBalance.value = Number(res.data.data.frozen_balance ?? 0);
  }

  return {
    balance,
    frozenBalance,
    fetchWallet,
  };
});
