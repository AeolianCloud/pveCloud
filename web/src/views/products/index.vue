<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createOrder } from '../../api/order'
import { getServerCatalog, type ServerCatalogPlan } from '../../api/product-catalog'
import { getApiErrorMessage } from '../../api/request'
import { useConfirm } from '../../composables/useConfirm'
import { useToast } from '../../composables/useToast'
import { useAuthStore } from '../../stores/auth'
import AppSelect, { type AppSelectOption } from '../../components/AppSelect.vue'

const router = useRouter()
const authStore = useAuthStore()
const confirmDialog = useConfirm()
const toast = useToast()
const catalogLoading = ref(false)
const catalogError = ref('')
const catalogPlans = ref<Array<ServerCatalogPlan & { productName: string }>>([])
const selectedRegion = ref('all')
const orderingPlanNo = ref('')
const orderDialogVisible = ref(false)
const orderPlan = ref<(ServerCatalogPlan & { productName: string }) | null>(null)
const selectedBillingCycle = ref('')
const selectedOrderRegion = ref('')
const selectedTemplate = ref('')
const selectedNetworkType = ref('')
const userNote = ref('')

const regionFilters = computed(() => {
  const regions = new Map<string, string>()
  for (const plan of catalogPlans.value) {
    for (const region of plan.regions) {
      regions.set(region.region_no, region.name)
    }
  }
  return Array.from(regions, ([regionNo, name]) => ({ regionNo, name }))
})

const filteredPlans = computed(() => {
  if (selectedRegion.value === 'all') {
    return catalogPlans.value
  }
  return catalogPlans.value.filter((plan) =>
    plan.regions.some((region) => region.region_no === selectedRegion.value),
  )
})

const displayedPlans = computed(() => filteredPlans.value)

const formatPrice = (plan: ServerCatalogPlan) => {
  const monthly = plan.prices.find((item) => item.billing_cycle === 'monthly') || plan.prices[0]
  if (!monthly) {
    return '价格待定'
  }
  return `¥${Math.round(monthly.price_cents / 100)}`
}

const formatMemory = (memoryMb: number) => {
  if (memoryMb >= 1024) {
    return `${Math.round(memoryMb / 1024)}GB`
  }
  return `${memoryMb}MB`
}

const selectedPrice = computed(() =>
  orderPlan.value?.prices.find((item) => item.billing_cycle === selectedBillingCycle.value) || null,
)

const cycleLabels: Record<string, string> = {
  monthly: '月付',
  quarterly: '季付',
  semi_yearly: '半年付',
  yearly: '年付',
}

const formatCycle = (cycle: string) => cycleLabels[cycle] || cycle

const formatMoney = (cents: number) => `¥${Math.round(cents / 100)}`

const billingCycleOptions = computed<AppSelectOption[]>(() =>
  (orderPlan.value?.prices || []).map((price) => ({
    label: formatCycle(price.billing_cycle),
    value: price.billing_cycle,
    description: `${formatMoney(price.price_cents)} · ${price.currency}`,
  })),
)

const regionOptions = computed<AppSelectOption[]>(() =>
  (orderPlan.value?.regions || []).map((region) => ({
    label: region.name,
    value: region.region_no,
    description: [region.country, region.city, region.summary].filter(Boolean).join(' · '),
  })),
)

const templateOptions = computed<AppSelectOption[]>(() =>
  (orderPlan.value?.os_templates || []).map((template) => ({
    label: template.name,
    value: template.template_no,
    description: [template.distribution, template.version, template.architecture, template.summary].filter(Boolean).join(' · '),
  })),
)

const networkTypeOptions = computed<AppSelectOption[]>(() =>
  (orderPlan.value?.network_types || []).map((networkType) => ({
    label: networkType.name,
    value: networkType.network_type_no,
    description: networkType.summary || `Code: ${networkType.code}`,
  })),
)

async function openOrderDialog(plan: ServerCatalogPlan & { productName: string }) {
  if (plan.status === 'sold_out') return
  if (!authStore.isAuthenticated) {
    await router.push({ path: '/login', query: { redirect: '/products' } })
    return
  }
  if (!plan.prices.length || !plan.regions.length || !plan.os_templates.length || !plan.network_types.length) {
    toast.error('当前套餐缺少价格、地域、系统模板或网络类型，暂不可下单')
    return
  }
  orderPlan.value = plan
  selectedBillingCycle.value = plan.prices[0].billing_cycle
  selectedOrderRegion.value = selectedRegion.value === 'all'
    ? plan.regions[0].region_no
    : (plan.regions.find((item) => item.region_no === selectedRegion.value)?.region_no || plan.regions[0].region_no)
  selectedTemplate.value = plan.os_templates[0].template_no
  selectedNetworkType.value = plan.network_types[0].network_type_no
  userNote.value = ''
  orderDialogVisible.value = true
}

async function createPlanOrder() {
  const plan = orderPlan.value
  if (!plan) return
  if (!selectedBillingCycle.value || !selectedOrderRegion.value || !selectedTemplate.value || !selectedNetworkType.value) {
    toast.error('请选择完整购买配置')
    return
  }
  orderingPlanNo.value = plan.plan_no
  try {
    const order = await createOrder({
      plan_no: plan.plan_no,
      billing_cycle: selectedBillingCycle.value,
      region_no: selectedOrderRegion.value,
      template_no: selectedTemplate.value,
      network_type_no: selectedNetworkType.value,
      quantity: 1,
      user_note: userNote.value.trim() || null,
      client_token: `web-${plan.plan_no}-${Date.now()}`,
    })
    orderDialogVisible.value = false
    await router.push(`/user/orders/${order.order_no}`)
  } catch (err) {
    const message = getApiErrorMessage(err, '订单创建失败')
    if (message.includes('实名')) {
      const confirmed = await confirmDialog.confirm({
        title: '需要实名认证',
        message: `${message}，是否前往实名认证？`,
        confirmText: '去认证',
        cancelText: '稍后再说',
      })
      if (confirmed) await router.push('/user/real-name')
    } else {
      toast.error(message)
    }
  } finally {
    orderingPlanNo.value = ''
  }
}

onMounted(async () => {
  catalogLoading.value = true
  try {
    const data = await getServerCatalog()
    catalogPlans.value = data.products.flatMap((product) =>
      product.plans.map((plan) => ({ ...plan, productName: product.name })),
    )
  } catch (err) {
    catalogError.value = getApiErrorMessage(err, '产品目录加载失败，当前展示静态骨架')
  } finally {
    catalogLoading.value = false
  }
})
</script>

<template>
  <div class="page-reveal bg-white">
    <section class="border-b border-neutral-200 bg-neutral-50">
      <div class="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Product Center</p>
        <div class="mt-4 grid gap-6 lg:grid-cols-[1fr_28rem] lg:items-end">
          <h1 class="text-5xl font-black tracking-tight text-neutral-950 sm:text-6xl">游戏云产品中心</h1>
          <p class="text-base leading-7 text-neutral-600">静态 UI 骨架先覆盖游戏云、高主频实例和物理服务器展示，后续再接入公开产品目录。</p>
        </div>
      </div>
    </section>

    <section class="mx-auto max-w-7xl px-4 py-14 sm:px-6 lg:px-8">
      <div class="mb-8 flex flex-wrap gap-3">
        <button
          type="button"
          :class="[
            'rounded-full border px-4 py-2 text-xs font-black',
            selectedRegion === 'all' ? 'border-neutral-950 bg-neutral-950 text-white' : 'chip-hover border-neutral-300 text-neutral-700',
          ]"
          @click="selectedRegion = 'all'"
        >
          全部
        </button>
        <button
          v-for="region in regionFilters"
          :key="region.regionNo"
          type="button"
          :class="[
            'rounded-full border px-4 py-2 text-xs font-black',
            selectedRegion === region.regionNo ? 'border-neutral-950 bg-neutral-950 text-white' : 'chip-hover border-neutral-300 text-neutral-700',
          ]"
          @click="selectedRegion = region.regionNo"
        >
          {{ region.name }}
        </button>
      </div>

      <div v-if="catalogLoading" class="grid gap-5 md:grid-cols-2 xl:grid-cols-4">
        <div v-for="item in 4" :key="item" class="rounded-[1.5rem] border border-neutral-200 bg-white p-6">
          <div class="skeleton-line h-3 w-24"></div>
          <div class="skeleton-line mt-5 h-6 w-32"></div>
          <div class="skeleton-line mt-8 h-10 w-24"></div>
          <div class="mt-7 space-y-3">
            <div class="skeleton-line h-3 w-full"></div>
            <div class="skeleton-line h-3 w-11/12"></div>
            <div class="skeleton-line h-3 w-10/12"></div>
          </div>
          <div class="skeleton-line mt-8 h-9 w-full"></div>
        </div>
      </div>
      <div v-else-if="catalogError" class="state-panel p-8 text-center">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Catalog Error</p>
        <h2 class="mt-3 text-2xl font-black text-neutral-950">产品目录暂时不可用</h2>
        <p class="mx-auto mt-3 max-w-xl text-sm leading-6 text-neutral-500">{{ catalogError }}</p>
      </div>

      <div v-else-if="displayedPlans.length" class="stagger-reveal grid gap-5 md:grid-cols-2 xl:grid-cols-4">
        <article v-for="plan in displayedPlans" :key="plan.plan_no" :class="['interactive-card rounded-[1.5rem] border bg-white p-6', plan.is_featured ? 'border-neutral-950 shadow-[6px_6px_0_#111]' : 'border-neutral-200']">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs font-black uppercase tracking-[0.16em] text-neutral-500">{{ plan.productName }}</p>
              <h2 class="mt-3 text-xl font-black text-neutral-950">{{ plan.name }}</h2>
            </div>
            <span :class="['rounded-full px-3 py-1 text-xs font-black', plan.is_featured ? 'bg-neutral-950 text-white' : 'border border-neutral-300']">
              {{ plan.is_featured ? '推荐' : plan.status === 'sold_out' ? '售罄' : '可选' }}
            </span>
          </div>
          <div class="mt-6"><span class="text-4xl font-black">{{ formatPrice(plan) }}</span><span v-if="plan.prices.length" class="text-neutral-500">/月起</span></div>
          <dl class="mt-6 space-y-3 text-sm">
            <div class="flex justify-between"><dt class="text-neutral-500">CPU</dt><dd class="font-bold">{{ plan.cpu_cores }} 核</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">内存</dt><dd class="font-bold">{{ formatMemory(plan.memory_mb) }}</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">硬盘</dt><dd class="font-bold">{{ plan.system_disk_gb + plan.data_disk_gb }}GB</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">带宽</dt><dd class="font-bold">{{ plan.bandwidth_mbps }}M</dd></div>
          </dl>
          <button type="button" class="action-pill mt-6 w-full border border-neutral-950 px-4 py-2 text-center text-sm font-black hover:bg-neutral-950 hover:text-white disabled:opacity-50" :disabled="plan.status === 'sold_out' || orderingPlanNo === plan.plan_no" @click="openOrderDialog(plan)">{{ plan.status === 'sold_out' ? '暂时售罄' : authStore.isAuthenticated ? (orderingPlanNo === plan.plan_no ? '创建中...' : '配置并创建订单') : '登录后创建订单' }}</button>
        </article>
      </div>

      <div v-else-if="catalogPlans.length" class="state-panel p-8 text-center">
        <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">No Region Plans</p>
        <h2 class="mt-3 text-2xl font-black text-neutral-950">当前地域暂无套餐</h2>
        <p class="mt-3 text-sm text-neutral-500">切换到全部地域，查看其它可选配置。</p>
        <button type="button" class="action-pill mt-5 border border-neutral-950 px-5 py-2 text-sm font-black hover:bg-neutral-950 hover:text-white" @click="selectedRegion = 'all'">查看全部地域</button>
      </div>

      <div v-else-if="!catalogLoading && !catalogError" class="stagger-reveal grid gap-5 lg:grid-cols-4">
        <article class="interactive-card rounded-[1.5rem] border border-neutral-200 bg-white p-6">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs font-black uppercase tracking-[0.16em] text-neutral-500">Starter</p>
              <h2 class="mt-3 text-xl font-black text-neutral-950">轻量开服</h2>
            </div>
            <span class="rounded-full border border-neutral-300 px-3 py-1 text-xs font-black">入门</span>
          </div>
          <div class="mt-6"><span class="text-4xl font-black">¥49</span><span class="text-neutral-500">/月起</span></div>
          <dl class="mt-6 space-y-3 text-sm">
            <div class="flex justify-between"><dt class="text-neutral-500">CPU</dt><dd class="font-bold">2 核</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">内存</dt><dd class="font-bold">4GB</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">硬盘</dt><dd class="font-bold">40GB SSD</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">带宽</dt><dd class="font-bold">10M</dd></div>
          </dl>
          <RouterLink to="/login" class="action-pill mt-6 w-full border border-neutral-950 px-4 py-2 text-center text-sm font-black hover:bg-neutral-950 hover:text-white">登录后查看购买入口</RouterLink>
        </article>

        <article class="interactive-card rounded-[1.5rem] border border-neutral-950 bg-white p-6 shadow-[8px_8px_0_#111]">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs font-black uppercase tracking-[0.16em] text-neutral-500">Popular</p>
              <h2 class="mt-3 text-xl font-black text-neutral-950">高主频游戏</h2>
            </div>
            <span class="rounded-full bg-neutral-950 px-3 py-1 text-xs font-black text-white">推荐</span>
          </div>
          <div class="mt-6"><span class="text-4xl font-black">¥199</span><span class="text-neutral-500">/月起</span></div>
          <dl class="mt-6 space-y-3 text-sm">
            <div class="flex justify-between"><dt class="text-neutral-500">CPU</dt><dd class="font-bold">4 核高频</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">内存</dt><dd class="font-bold">8GB</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">硬盘</dt><dd class="font-bold">80GB SSD</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">带宽</dt><dd class="font-bold">30M</dd></div>
          </dl>
          <RouterLink to="/login" class="action-pill mt-6 w-full bg-neutral-950 px-4 py-2 text-center text-sm font-black text-white hover:bg-neutral-800">登录后查看购买入口</RouterLink>
        </article>

        <article class="interactive-card rounded-[1.5rem] border border-neutral-200 bg-white p-6">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs font-black uppercase tracking-[0.16em] text-neutral-500">Defense</p>
              <h2 class="mt-3 text-xl font-black text-neutral-950">高防大带宽</h2>
            </div>
            <span class="rounded-full border border-neutral-300 px-3 py-1 text-xs font-black">防护</span>
          </div>
          <div class="mt-6"><span class="text-4xl font-black">¥399</span><span class="text-neutral-500">/月起</span></div>
          <dl class="mt-6 space-y-3 text-sm">
            <div class="flex justify-between"><dt class="text-neutral-500">CPU</dt><dd class="font-bold">4 核</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">内存</dt><dd class="font-bold">8GB</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">硬盘</dt><dd class="font-bold">100GB SSD</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">带宽</dt><dd class="font-bold">100M</dd></div>
          </dl>
          <RouterLink to="/login" class="action-pill mt-6 w-full border border-neutral-950 px-4 py-2 text-center text-sm font-black hover:bg-neutral-950 hover:text-white">登录后查看购买入口</RouterLink>
        </article>

        <article class="interactive-card rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs font-black uppercase tracking-[0.16em] text-neutral-500">Bare Metal</p>
              <h2 class="mt-3 text-xl font-black text-neutral-950">独享物理机</h2>
            </div>
            <span class="rounded-full border border-neutral-300 px-3 py-1 text-xs font-black">咨询</span>
          </div>
          <div class="mt-6"><span class="text-4xl font-black">¥899</span><span class="text-neutral-500">/月起</span></div>
          <dl class="mt-6 space-y-3 text-sm">
            <div class="flex justify-between"><dt class="text-neutral-500">CPU</dt><dd class="font-bold">E5 / i9</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">内存</dt><dd class="font-bold">32GB+</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">硬盘</dt><dd class="font-bold">定制</dd></div>
            <div class="flex justify-between"><dt class="text-neutral-500">带宽</dt><dd class="font-bold">定制</dd></div>
          </dl>
          <RouterLink to="/login" class="action-pill mt-6 w-full border border-neutral-950 px-4 py-2 text-center text-sm font-black hover:bg-neutral-950 hover:text-white">登录后查看购买入口</RouterLink>
        </article>
      </div>
    </section>

    <section class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div class="stagger-reveal grid gap-5 lg:grid-cols-3">
        <div class="soft-lift rounded-[1.5rem] border border-neutral-200 p-6">
          <h3 class="text-lg font-black">杭州 BGP</h3>
          <p class="mt-3 text-sm leading-6 text-neutral-600">适合华东玩家访问、轻量联机场景和常规网站业务。</p>
        </div>
        <div class="soft-lift rounded-[1.5rem] border border-neutral-200 p-6">
          <h3 class="text-lg font-black">镇江高防</h3>
          <p class="mt-3 text-sm leading-6 text-neutral-600">适合需要基础防护展示的游戏社区和活动节点。</p>
        </div>
        <div class="soft-lift rounded-[1.5rem] border border-neutral-200 p-6">
          <h3 class="text-lg font-black">香港节点</h3>
          <p class="mt-3 text-sm leading-6 text-neutral-600">适合海外访问展示，不承诺业务交付，具体以后端目录为准。</p>
        </div>
      </div>
    </section>

    <div v-if="orderDialogVisible && orderPlan" class="fixed inset-0 z-50 flex items-end justify-center bg-neutral-950/45 px-4 py-6 sm:items-center" @click.self="orderDialogVisible = false">
      <div class="max-h-[92vh] w-full max-w-3xl overflow-hidden rounded-[1.5rem] border border-neutral-950 bg-white shadow-[8px_8px_0_#111]">
        <div class="flex items-start justify-between gap-4 border-b border-neutral-200 p-5 pb-4 sm:p-6 sm:pb-4">
          <div>
            <p class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Create Order</p>
            <h2 class="mt-2 text-2xl font-black text-neutral-950">确认购买配置</h2>
            <p class="mt-2 text-sm text-neutral-500">固定套餐规格不可修改，请选择地域、系统模板和网络类型。</p>
          </div>
          <button type="button" class="action-pill border border-neutral-300 px-3 py-1 text-sm font-black hover:border-neutral-950" @click="orderDialogVisible = false">关闭</button>
        </div>

        <div class="max-h-[calc(92vh-12rem)] overflow-y-auto p-5 sm:p-6">
        <div class="grid gap-5 lg:grid-cols-[16rem_1fr]">
          <div class="rounded-2xl border border-neutral-200 bg-neutral-50 p-4">
            <p class="text-xs font-black uppercase tracking-[0.14em] text-neutral-500">{{ orderPlan.productName }}</p>
            <h3 class="mt-2 text-lg font-black text-neutral-950">{{ orderPlan.name }}</h3>
            <dl class="mt-4 grid grid-cols-2 gap-2 text-xs">
              <div class="rounded-xl bg-white p-3"><dt class="text-neutral-500">CPU</dt><dd class="mt-1 font-black">{{ orderPlan.cpu_cores }} 核</dd></div>
              <div class="rounded-xl bg-white p-3"><dt class="text-neutral-500">内存</dt><dd class="mt-1 font-black">{{ formatMemory(orderPlan.memory_mb) }}</dd></div>
              <div class="rounded-xl bg-white p-3"><dt class="text-neutral-500">硬盘</dt><dd class="mt-1 font-black">{{ orderPlan.system_disk_gb + orderPlan.data_disk_gb }}GB</dd></div>
              <div class="rounded-xl bg-white p-3"><dt class="text-neutral-500">带宽</dt><dd class="mt-1 font-black">{{ orderPlan.bandwidth_mbps }}M</dd></div>
              <div class="col-span-2 rounded-xl bg-white p-3"><dt class="text-neutral-500">公网 IP</dt><dd class="mt-1 font-black">{{ orderPlan.public_ip_count }} 个</dd></div>
            </dl>
          </div>

          <div class="space-y-4">
            <div>
              <label class="mb-2 block text-sm font-black text-neutral-800">计费周期</label>
              <AppSelect v-model="selectedBillingCycle" :options="billingCycleOptions" placeholder="请选择计费周期" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-black text-neutral-800">销售地域</label>
              <AppSelect v-model="selectedOrderRegion" :options="regionOptions" placeholder="请选择销售地域" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-black text-neutral-800">系统模板</label>
              <AppSelect v-model="selectedTemplate" :options="templateOptions" placeholder="请选择系统模板" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-black text-neutral-800">网络类型</label>
              <AppSelect v-model="selectedNetworkType" :options="networkTypeOptions" placeholder="请选择网络类型" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-black text-neutral-800">备注</label>
              <textarea v-model="userNote" rows="2" class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="可填写开服用途或其它人工处理说明" />
            </div>
          </div>
        </div>

        </div>

        <div class="sticky bottom-0 flex flex-col gap-3 border-t border-neutral-200 bg-white/95 p-5 backdrop-blur sm:flex-row sm:items-center sm:justify-between sm:p-6">
          <div>
            <p class="text-sm text-neutral-500">订单金额</p>
            <p class="text-3xl font-black text-neutral-950">{{ selectedPrice ? formatMoney(selectedPrice.price_cents) : '价格待定' }}</p>
          </div>
          <button type="button" :disabled="orderingPlanNo === orderPlan.plan_no" class="btn-dark action-pill border px-8 py-3 text-sm font-black disabled:opacity-50" @click="createPlanOrder">
            {{ orderingPlanNo === orderPlan.plan_no ? '创建中...' : '确认创建订单' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
