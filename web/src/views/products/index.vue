<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getServerCatalog, type ServerCatalogPlan } from '../../api/product-catalog'
import { getApiErrorMessage } from '../../api/request'

const catalogLoading = ref(false)
const catalogError = ref('')
const catalogPlans = ref<Array<ServerCatalogPlan & { productName: string }>>([])

const featuredPlans = computed(() => catalogPlans.value.slice(0, 4))

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
        <span class="rounded-full border border-neutral-950 bg-neutral-950 px-4 py-2 text-xs font-black text-white">全部</span>
        <span class="chip-hover rounded-full border border-neutral-300 px-4 py-2 text-xs font-black text-neutral-700">Minecraft</span>
        <span class="chip-hover rounded-full border border-neutral-300 px-4 py-2 text-xs font-black text-neutral-700">Steam 游戏服</span>
        <span class="chip-hover rounded-full border border-neutral-300 px-4 py-2 text-xs font-black text-neutral-700">高防节点</span>
      </div>

      <div v-if="catalogLoading" class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">
        产品目录加载中...
      </div>
      <div v-else-if="catalogError" class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6 text-sm font-bold text-neutral-600">
        {{ catalogError }}
      </div>

      <div v-if="featuredPlans.length" class="stagger-reveal grid gap-5 lg:grid-cols-4">
        <article v-for="plan in featuredPlans" :key="plan.plan_no" :class="['interactive-card rounded-[1.5rem] border bg-white p-6', plan.is_featured ? 'border-neutral-950 shadow-[8px_8px_0_#111]' : 'border-neutral-200']">
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
          <RouterLink to="/login" class="mt-6 block rounded-full border border-neutral-950 px-4 py-2 text-center text-sm font-black hover:bg-neutral-950 hover:text-white">登录后查看购买入口</RouterLink>
        </article>
      </div>

      <div v-else class="stagger-reveal grid gap-5 lg:grid-cols-4">
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
          <RouterLink to="/login" class="mt-6 block rounded-full border border-neutral-950 px-4 py-2 text-center text-sm font-black hover:bg-neutral-950 hover:text-white">登录后查看购买入口</RouterLink>
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
          <RouterLink to="/login" class="mt-6 block rounded-full bg-neutral-950 px-4 py-2 text-center text-sm font-black text-white hover:bg-neutral-800">登录后查看购买入口</RouterLink>
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
          <RouterLink to="/login" class="mt-6 block rounded-full border border-neutral-950 px-4 py-2 text-center text-sm font-black hover:bg-neutral-950 hover:text-white">登录后查看购买入口</RouterLink>
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
          <RouterLink to="/login" class="mt-6 block rounded-full border border-neutral-950 px-4 py-2 text-center text-sm font-black hover:bg-neutral-950 hover:text-white">登录后查看购买入口</RouterLink>
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
  </div>
</template>
