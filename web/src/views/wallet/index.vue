<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { createWalletRecharge, getWallet, getWalletLedger, getWalletRecharge, type WalletLedgerItem, type WalletRechargeStatus, type WalletSummary } from '../../api/wallet'
import { getApiErrorMessage } from '../../api/request'
import QrCodeImage from '../../components/QrCodeImage.vue'
import { useToast } from '../../composables/useToast'

const toast = useToast()
const loading = ref(false)
const ledgerLoading = ref(false)
const rechargeLoading = ref(false)
const errorMessage = ref('')
const wallet = ref<WalletSummary | null>(null)
const ledger = ref<WalletLedgerItem[]>([])
const latestRecharge = ref<WalletRechargeStatus | null>(null)
const rechargeAmount = ref(1000)
let rechargeTimer: number | undefined

const ledgerQuery = reactive({ page: 1, per_page: 10 })
const statusText: Record<string, string> = { active: '正常', disabled: '已停用' }
const directionText: Record<string, string> = { credit: '入账', debit: '支出' }
const entryTypeText: Record<string, string> = { recharge: '充值', payment: '余额支付', refund: '退款退回' }
const rechargeStatusText: Record<string, string> = { pending: '待支付', paid: '已入账', closed: '已关闭', failed: '失败' }

const formatMoney = (cents: number) => `¥${(cents / 100).toFixed(2)}`

async function loadWallet() {
  loading.value = true
  errorMessage.value = ''
  try {
    wallet.value = await getWallet()
  } catch (err) {
    errorMessage.value = getApiErrorMessage(err, '钱包加载失败')
  } finally {
    loading.value = false
  }
}

async function loadLedger() {
  ledgerLoading.value = true
  try {
    const data = await getWalletLedger(ledgerQuery)
    ledger.value = data.list
  } catch (err) {
    toast.error(getApiErrorMessage(err, '流水加载失败'))
  } finally {
    ledgerLoading.value = false
  }
}

async function startRecharge(provider: 'alipay' | 'wechat', method: 'alipay_page' | 'wechat_native') {
  rechargeLoading.value = true
  try {
    latestRecharge.value = await createWalletRecharge({
      provider,
      method,
      amount_cents: rechargeAmount.value,
      client_token: `rch-${provider}-${method}-${Date.now()}`,
    })
    if (latestRecharge.value.redirect_url) {
      window.location.href = latestRecharge.value.redirect_url
      return
    }
    startRechargePolling()
  } catch (err) {
    toast.error(getApiErrorMessage(err, '创建充值失败'))
  } finally {
    rechargeLoading.value = false
  }
}

function startRechargePolling() {
  stopRechargePolling()
  rechargeTimer = window.setInterval(async () => {
    if (!latestRecharge.value || latestRecharge.value.status !== 'pending') {
      stopRechargePolling()
      return
    }
    latestRecharge.value = await getWalletRecharge(latestRecharge.value.recharge_no)
    if (latestRecharge.value.status !== 'pending') {
      stopRechargePolling()
      await loadWallet()
      await loadLedger()
    }
  }, 3000)
}

function stopRechargePolling() {
  if (rechargeTimer) {
    window.clearInterval(rechargeTimer)
    rechargeTimer = undefined
  }
}

onMounted(async () => {
  await loadWallet()
  await loadLedger()
})

onBeforeUnmount(stopRechargePolling)
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-6xl px-4 py-12 sm:px-6 lg:px-8">
      <div class="mb-8 flex flex-col justify-between gap-4 border-b border-neutral-200 pb-6 md:flex-row md:items-end">
        <div>
          <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Wallet</p>
          <h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">我的钱包</h1>
        </div>
      </div>

      <div v-if="loading && !wallet" class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">钱包加载中...</div>
      <div v-else-if="errorMessage" class="rounded-[1.5rem] border border-red-200 bg-red-50 p-6 text-sm font-bold text-red-700">{{ errorMessage }}</div>
      <div v-else-if="wallet" class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_24rem]">
        <section class="rounded-[1.5rem] border border-neutral-950 bg-white p-6 shadow-[8px_8px_0_#111]">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <p class="text-xs font-black uppercase tracking-[0.16em] text-neutral-500">{{ wallet.wallet_no }}</p>
              <div class="mt-3 text-4xl font-black">{{ formatMoney(wallet.available_balance_cents) }}</div>
            </div>
            <span class="rounded-full border border-neutral-950 px-3 py-1 text-xs font-black">{{ statusText[wallet.status] || wallet.status }}</span>
          </div>
          <dl class="mt-6 grid gap-3 md:grid-cols-3">
            <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">累计充值</dt><dd class="mt-1 text-sm font-black">{{ formatMoney(wallet.total_recharged_cents) }}</dd></div>
            <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">累计消费</dt><dd class="mt-1 text-sm font-black">{{ formatMoney(wallet.total_spent_cents) }}</dd></div>
            <div class="rounded-xl bg-neutral-50 p-3"><dt class="text-xs font-black text-neutral-500">退款退回</dt><dd class="mt-1 text-sm font-black">{{ formatMoney(wallet.total_refunded_cents) }}</dd></div>
          </dl>
        </section>

        <section class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6">
          <h2 class="text-lg font-black text-neutral-950">充值</h2>
          <label class="mt-5 block text-xs font-black text-neutral-500">金额</label>
          <input v-model.number="rechargeAmount" min="100" step="100" type="number" class="mt-2 w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm font-black outline-none focus:border-neutral-950" />
          <div class="mt-4 flex flex-wrap gap-3">
            <button type="button" class="action-pill border border-neutral-950 px-4 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white disabled:opacity-50" :disabled="rechargeLoading" @click="startRecharge('alipay', 'alipay_page')">支付宝</button>
            <button type="button" class="action-pill border border-neutral-950 px-4 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white disabled:opacity-50" :disabled="rechargeLoading" @click="startRecharge('wechat', 'wechat_native')">微信扫码</button>
          </div>
          <div v-if="latestRecharge" class="mt-5 rounded-xl border border-neutral-200 bg-white p-4 text-sm">
            <div class="font-black">{{ latestRecharge.recharge_no }} · {{ rechargeStatusText[latestRecharge.status] || latestRecharge.status }}</div>
            <QrCodeImage v-if="latestRecharge.qr_code_url" class="mt-3" :value="latestRecharge.qr_code_url" alt="钱包充值二维码" />
            <div v-if="latestRecharge.qr_code_url" class="mt-3 break-all rounded-lg bg-neutral-50 p-3 text-xs font-bold text-neutral-600">{{ latestRecharge.qr_code_url }}</div>
          </div>
        </section>

        <section class="lg:col-span-2">
          <div class="mb-3 flex items-center justify-between">
            <h2 class="text-lg font-black text-neutral-950">钱包流水</h2>
            <button type="button" class="text-sm font-black underline" :disabled="ledgerLoading" @click="loadLedger">刷新</button>
          </div>
          <div class="overflow-hidden rounded-[1.25rem] border border-neutral-200">
            <div v-if="ledgerLoading" class="bg-neutral-50 p-5 text-sm font-bold text-neutral-600">流水加载中...</div>
            <div v-else-if="ledger.length === 0" class="bg-neutral-50 p-5 text-sm font-bold text-neutral-600">暂无流水</div>
            <div v-for="item in ledger" :key="item.entry_no" class="grid gap-2 border-b border-neutral-100 p-4 text-sm last:border-b-0 md:grid-cols-[8rem_minmax(0,1fr)_9rem_11rem] md:items-center">
              <div class="font-black">{{ directionText[item.direction] }}</div>
              <div class="min-w-0"><div class="font-black">{{ entryTypeText[item.entry_type] || item.entry_type }} · {{ item.related_no }}</div><div class="truncate text-xs text-neutral-500">{{ item.created_at }}</div></div>
              <div class="font-black" :class="item.direction === 'credit' ? 'text-emerald-700' : 'text-neutral-950'">{{ item.direction === 'credit' ? '+' : '-' }}{{ formatMoney(item.amount_cents) }}</div>
              <div class="text-neutral-500">余额 {{ formatMoney(item.balance_after_cents) }}</div>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>
