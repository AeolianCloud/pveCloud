<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NPagination,
  NSelect,
  NSpace,
  NTabPane,
  NTabs,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onMounted, reactive, ref } from 'vue'

import { getWalletDetail, getWalletLedger, getWalletRecharges, getWallets, type WalletDetail, type WalletItem, type WalletLedgerItem, type WalletRechargeItem } from '../../api/wallet'
import { formatDateTime } from '../../utils/datetime'
import { message } from '../../utils/feedback'
import type { WalletLedgerQueryState, WalletQueryState, WalletRechargeQueryState } from './types'

const activeTab = ref<'wallets' | 'ledger' | 'recharges'>('wallets')
const loadingWallets = ref(false)
const loadingLedger = ref(false)
const loadingRecharges = ref(false)
const detailLoading = ref(false)
const wallets = ref<WalletItem[]>([])
const ledger = ref<WalletLedgerItem[]>([])
const recharges = ref<WalletRechargeItem[]>([])
const walletTotal = ref(0)
const ledgerTotal = ref(0)
const rechargeTotal = ref(0)
const detailVisible = ref(false)
const selectedWallet = ref<WalletDetail | null>(null)

const walletQuery = reactive<WalletQueryState>({ page: 1, per_page: 15, wallet_no: '', user_keyword: '', status: '' })
const ledgerQuery = reactive<WalletLedgerQueryState>({ page: 1, per_page: 15, wallet_no: '', user_keyword: '', direction: '', entry_type: '', related_no: '' })
const rechargeQuery = reactive<WalletRechargeQueryState>({ page: 1, per_page: 15, wallet_no: '', user_keyword: '', provider: '', method: '', status: '', recharge_no: '' })

const statusOptions = [{ label: '正常', value: 'active' }, { label: '已停用', value: 'disabled' }]
const directionOptions = [{ label: '入账', value: 'credit' }, { label: '支出', value: 'debit' }]
const entryTypeOptions = [{ label: '充值', value: 'recharge' }, { label: '余额支付', value: 'payment' }, { label: '退款退回', value: 'refund' }]
const providerOptions = [{ label: '支付宝', value: 'alipay' }, { label: '微信支付', value: 'wechat' }]
const methodOptions = [{ label: '支付宝电脑网页', value: 'alipay_page' }, { label: '支付宝手机网页', value: 'alipay_wap' }, { label: '微信 Native 扫码', value: 'wechat_native' }, { label: '微信 H5', value: 'wechat_h5' }]
const rechargeStatusOptions = [{ label: '待支付', value: 'pending' }, { label: '已入账', value: 'paid' }, { label: '已关闭', value: 'closed' }, { label: '失败', value: 'failed' }]

const statusText: Record<string, string> = { active: '正常', disabled: '已停用' }
const directionText: Record<string, string> = { credit: '入账', debit: '支出' }
const entryTypeText: Record<string, string> = { recharge: '充值', payment: '余额支付', refund: '退款退回' }
const providerText: Record<string, string> = { alipay: '支付宝', wechat: '微信支付' }
const methodText: Record<string, string> = { alipay_page: '支付宝电脑网页', alipay_wap: '支付宝手机网页', wechat_native: '微信 Native 扫码', wechat_h5: '微信 H5' }
const rechargeStatusText: Record<string, string> = { pending: '待支付', paid: '已入账', closed: '已关闭', failed: '失败' }

function formatMoney(cents: number, currency = 'CNY') {
  return `${currency} ${(cents / 100).toFixed(2)}`
}

function userLabel(row: { user: WalletItem['user'] }) {
  return row.user.display_name || row.user.username || row.user.email || `用户 ${row.user.id}`
}

async function loadWallets() {
  loadingWallets.value = true
  try {
    const data = await getWallets({ ...walletQuery })
    wallets.value = data.list
    walletTotal.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '钱包列表加载失败')
  } finally {
    loadingWallets.value = false
  }
}

async function loadLedger() {
  loadingLedger.value = true
  try {
    const data = await getWalletLedger({ ...ledgerQuery })
    ledger.value = data.list
    ledgerTotal.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '钱包流水加载失败')
  } finally {
    loadingLedger.value = false
  }
}

async function loadRecharges() {
  loadingRecharges.value = true
  try {
    const data = await getWalletRecharges({ ...rechargeQuery })
    recharges.value = data.list
    rechargeTotal.value = data.total
  } catch (err) {
    message.error(err instanceof Error ? err.message : '充值记录加载失败')
  } finally {
    loadingRecharges.value = false
  }
}

async function openDetail(walletNo: string) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    selectedWallet.value = await getWalletDetail(walletNo)
  } catch (err) {
    message.error(err instanceof Error ? err.message : '钱包详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

function resetWalletQuery() {
  Object.assign(walletQuery, { page: 1, per_page: 15, wallet_no: '', user_keyword: '', status: '' })
  void loadWallets()
}

const walletColumns = computed<DataTableColumns<WalletItem>>(() => [
  { key: 'wallet_no', title: '钱包编号', minWidth: 170 },
  { key: 'user', title: '用户', minWidth: 150, render: userLabel },
  { key: 'status', title: '状态', width: 100, render: (row) => h(NTag, { size: 'small', type: row.status === 'active' ? 'success' : 'warning' }, { default: () => statusText[row.status] || row.status }) },
  { key: 'balance', title: '可用余额', width: 140, render: (row) => formatMoney(row.available_balance_cents, row.currency) },
  { key: 'total_recharged', title: '累计充值', width: 140, render: (row) => formatMoney(row.total_recharged_cents, row.currency) },
  { key: 'total_spent', title: '累计消费', width: 140, render: (row) => formatMoney(row.total_spent_cents, row.currency) },
  { key: 'total_refunded', title: '退款退回', width: 140, render: (row) => formatMoney(row.total_refunded_cents, row.currency) },
  { key: 'created_at', title: '创建时间', minWidth: 170, render: (row) => formatDateTime(row.created_at) },
  { key: 'actions', title: '操作', width: 100, fixed: 'right', render: (row) => h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row.wallet_no) }, { default: () => '详情' }) },
])

const ledgerColumns = computed<DataTableColumns<WalletLedgerItem>>(() => [
  { key: 'entry_no', title: '流水编号', minWidth: 170 },
  { key: 'wallet_no', title: '钱包编号', minWidth: 170 },
  { key: 'user', title: '用户', minWidth: 150, render: userLabel },
  { key: 'direction', title: '方向', width: 90, render: (row) => directionText[row.direction] || row.direction },
  { key: 'entry_type', title: '类型', width: 110, render: (row) => entryTypeText[row.entry_type] || row.entry_type },
  { key: 'amount', title: '金额', width: 130, render: (row) => formatMoney(row.amount_cents, row.currency) },
  { key: 'balance', title: '变动后余额', width: 140, render: (row) => formatMoney(row.balance_after_cents, row.currency) },
  { key: 'related_no', title: '关联编号', minWidth: 170 },
  { key: 'created_at', title: '创建时间', minWidth: 170, render: (row) => formatDateTime(row.created_at) },
])

const rechargeColumns = computed<DataTableColumns<WalletRechargeItem>>(() => [
  { key: 'recharge_no', title: '充值编号', minWidth: 170 },
  { key: 'wallet_no', title: '钱包编号', minWidth: 170 },
  { key: 'user', title: '用户', minWidth: 150, render: userLabel },
  { key: 'provider', title: '渠道', width: 100, render: (row) => providerText[row.provider] || row.provider },
  { key: 'method', title: '方式', minWidth: 150, render: (row) => methodText[row.method] || row.method },
  { key: 'status', title: '状态', width: 100, render: (row) => h(NTag, { size: 'small', type: row.status === 'paid' ? 'success' : row.status === 'failed' ? 'error' : 'default' }, { default: () => rechargeStatusText[row.status] || row.status }) },
  { key: 'amount', title: '金额', width: 130, render: (row) => formatMoney(row.amount_cents, row.currency) },
  { key: 'created_at', title: '创建时间', minWidth: 170, render: (row) => formatDateTime(row.created_at) },
])

onMounted(() => {
  void loadWallets()
  void loadLedger()
  void loadRecharges()
})
</script>

<template>
  <div class="wallets-page">
    <NCard :bordered="false">
      <template #header>
        <div class="page-header">
          <h2>钱包管理</h2>
          <p class="muted">只读查看用户钱包、充值记录和余额流水。</p>
        </div>
      </template>

      <NTabs v-model:value="activeTab" type="line" animated>
        <NTabPane name="wallets" tab="钱包账户">
          <NForm inline label-placement="left" class="query-form">
            <NFormItem label="状态"><NSelect v-model:value="walletQuery.status" :options="statusOptions" clearable placeholder="全部" style="width: 120px" /></NFormItem>
            <NFormItem label="钱包"><NInput v-model:value="walletQuery.wallet_no" clearable placeholder="钱包编号" /></NFormItem>
            <NFormItem label="用户"><NInput v-model:value="walletQuery.user_keyword" clearable placeholder="用户名/邮箱" /></NFormItem>
            <NFormItem :show-label="false"><NSpace><NButton type="primary" @click="walletQuery.page = 1; loadWallets()">查询</NButton><NButton @click="resetWalletQuery">重置</NButton></NSpace></NFormItem>
          </NForm>
          <NDataTable :loading="loadingWallets" :columns="walletColumns" :data="wallets" :row-key="(row: WalletItem) => row.wallet_no" :bordered="false" />
          <div class="pagination"><NPagination v-model:page="walletQuery.page" v-model:page-size="walletQuery.per_page" :item-count="walletTotal" show-size-picker :page-sizes="[10, 15, 20, 50]" @update:page="loadWallets" @update:page-size="loadWallets" /></div>
        </NTabPane>

        <NTabPane name="ledger" tab="钱包流水">
          <NForm inline label-placement="left" class="query-form">
            <NFormItem label="方向"><NSelect v-model:value="ledgerQuery.direction" :options="directionOptions" clearable placeholder="全部" style="width: 110px" /></NFormItem>
            <NFormItem label="类型"><NSelect v-model:value="ledgerQuery.entry_type" :options="entryTypeOptions" clearable placeholder="全部" style="width: 130px" /></NFormItem>
            <NFormItem label="钱包"><NInput v-model:value="ledgerQuery.wallet_no" clearable placeholder="钱包编号" /></NFormItem>
            <NFormItem label="用户"><NInput v-model:value="ledgerQuery.user_keyword" clearable placeholder="用户名/邮箱" /></NFormItem>
            <NFormItem label="关联"><NInput v-model:value="ledgerQuery.related_no" clearable placeholder="关联编号" /></NFormItem>
            <NFormItem :show-label="false"><NButton type="primary" @click="ledgerQuery.page = 1; loadLedger()">查询</NButton></NFormItem>
          </NForm>
          <NDataTable :loading="loadingLedger" :columns="ledgerColumns" :data="ledger" :row-key="(row: WalletLedgerItem) => row.entry_no" :bordered="false" />
          <div class="pagination"><NPagination v-model:page="ledgerQuery.page" v-model:page-size="ledgerQuery.per_page" :item-count="ledgerTotal" show-size-picker :page-sizes="[10, 15, 20, 50]" @update:page="loadLedger" @update:page-size="loadLedger" /></div>
        </NTabPane>

        <NTabPane name="recharges" tab="充值记录">
          <NForm inline label-placement="left" class="query-form">
            <NFormItem label="渠道"><NSelect v-model:value="rechargeQuery.provider" :options="providerOptions" clearable placeholder="全部" style="width: 120px" /></NFormItem>
            <NFormItem label="方式"><NSelect v-model:value="rechargeQuery.method" :options="methodOptions" clearable placeholder="全部方式" style="width: 180px" /></NFormItem>
            <NFormItem label="状态"><NSelect v-model:value="rechargeQuery.status" :options="rechargeStatusOptions" clearable placeholder="全部" style="width: 120px" /></NFormItem>
            <NFormItem label="编号"><NSpace><NInput v-model:value="rechargeQuery.recharge_no" clearable placeholder="充值编号" /><NInput v-model:value="rechargeQuery.wallet_no" clearable placeholder="钱包编号" /></NSpace></NFormItem>
            <NFormItem label="用户"><NInput v-model:value="rechargeQuery.user_keyword" clearable placeholder="用户名/邮箱" /></NFormItem>
            <NFormItem :show-label="false"><NButton type="primary" @click="rechargeQuery.page = 1; loadRecharges()">查询</NButton></NFormItem>
          </NForm>
          <NDataTable :loading="loadingRecharges" :columns="rechargeColumns" :data="recharges" :row-key="(row: WalletRechargeItem) => row.recharge_no" :bordered="false" />
          <div class="pagination"><NPagination v-model:page="rechargeQuery.page" v-model:page-size="rechargeQuery.per_page" :item-count="rechargeTotal" show-size-picker :page-sizes="[10, 15, 20, 50]" @update:page="loadRecharges" @update:page-size="loadRecharges" /></div>
        </NTabPane>
      </NTabs>
    </NCard>

    <NDrawer v-model:show="detailVisible" :width="620" placement="right">
      <NDrawerContent title="钱包详情">
        <div v-if="detailLoading">详情加载中...</div>
        <template v-else-if="selectedWallet">
          <NDescriptions :column="1" label-placement="left" bordered size="small">
            <NDescriptionsItem label="钱包编号">{{ selectedWallet.wallet_no }}</NDescriptionsItem>
            <NDescriptionsItem label="用户">{{ userLabel(selectedWallet) }}</NDescriptionsItem>
            <NDescriptionsItem label="状态">{{ statusText[selectedWallet.status] }}</NDescriptionsItem>
            <NDescriptionsItem label="可用余额">{{ formatMoney(selectedWallet.available_balance_cents, selectedWallet.currency) }}</NDescriptionsItem>
            <NDescriptionsItem label="累计充值">{{ formatMoney(selectedWallet.total_recharged_cents, selectedWallet.currency) }}</NDescriptionsItem>
            <NDescriptionsItem label="累计消费">{{ formatMoney(selectedWallet.total_spent_cents, selectedWallet.currency) }}</NDescriptionsItem>
            <NDescriptionsItem label="退款退回">{{ formatMoney(selectedWallet.total_refunded_cents, selectedWallet.currency) }}</NDescriptionsItem>
          </NDescriptions>
        </template>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
.page-header h2 {
  margin: 0;
  font-size: 20px;
}
.muted {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
}
.query-form {
  margin-bottom: 16px;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
