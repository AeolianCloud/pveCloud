<!-- 产品套餐：阿里云风格，深色背景，卡片式展示 -->
<template>
  <section id="products" class="bg-[#080f1a] py-20">
    <div class="max-w-[1200px] mx-auto px-6">
      <!-- 区块标题 -->
      <div class="text-center mb-12">
        <h2 class="text-3xl font-bold text-white mb-3">独立服务器套餐</h2>
        <p class="text-gray-400 text-sm">所有套餐均为独享物理资源，按自然月计费，支持 Linux / Windows</p>
      </div>

      <!-- 分类 Tab -->
      <div class="flex justify-center mb-10">
        <div class="flex gap-1 p-1 rounded-lg bg-white/5 border border-white/10">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            @click="activeTab = tab.key"
            :class="[
              'px-5 py-2 rounded text-sm font-medium transition-all duration-150',
              activeTab === tab.key
                ? 'bg-[#ff6a00] text-white shadow-lg shadow-[#ff6a00]/20'
                : 'text-gray-400 hover:text-white'
            ]"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>

      <!-- 产品卡片网格 -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div
          v-for="product in filteredProducts"
          :key="product.id"
          :class="[
            'relative rounded-xl border p-6 transition-all duration-200 hover:-translate-y-1 hover:shadow-xl group',
            product.featured
              ? 'border-[#ff6a00]/40 bg-gradient-to-b from-[#ff6a00]/5 to-[#0e1829] shadow-lg shadow-[#ff6a00]/10'
              : 'border-white/10 bg-[#0e1829] hover:border-white/20'
          ]"
        >
          <!-- 推荐角标 -->
          <div v-if="product.featured" class="absolute -top-3 left-1/2 -translate-x-1/2">
            <span class="px-3 py-1 rounded-full bg-[#ff6a00] text-white text-xs font-semibold">热销推荐</span>
          </div>

          <!-- 型号 + 描述 -->
          <div class="mb-5">
            <div class="flex items-center gap-2 mb-1">
              <span class="text-lg font-bold text-white">{{ product.name }}</span>
              <span class="px-2 py-0.5 rounded text-xs font-medium bg-blue-500/20 text-blue-300">
                {{ product.typeLabel }}
              </span>
            </div>
            <p class="text-xs text-gray-500">{{ product.desc }}</p>
          </div>

          <!-- 规格参数 -->
          <div class="space-y-2.5 mb-6">
            <div v-for="spec in product.specs" :key="spec.label" class="flex justify-between text-sm">
              <span class="text-gray-500">{{ spec.label }}</span>
              <span class="text-gray-200 font-medium">{{ spec.value }}</span>
            </div>
          </div>

          <!-- 分割线 -->
          <div class="border-t border-white/10 mb-5"></div>

          <!-- 价格 + 购买 -->
          <div class="flex items-end justify-between">
            <div>
              <div class="text-xs text-gray-500 mb-0.5">月付起价</div>
              <div class="flex items-baseline gap-0.5">
                <span class="text-xs text-gray-300">¥</span>
                <span class="text-3xl font-bold text-white">{{ product.price }}</span>
                <span class="text-sm text-gray-500">/月</span>
              </div>
            </div>
            <a
              href="#"
              :class="[
                'px-5 py-2.5 rounded text-sm font-semibold transition-all duration-150',
                product.featured
                  ? 'bg-[#ff6a00] hover:bg-[#e85f00] text-white'
                  : 'border border-white/20 text-white hover:border-[#ff6a00] hover:text-[#ff6a00]'
              ]"
            >
              立即购买
            </a>
          </div>
        </div>
      </div>

      <!-- 底部说明 -->
      <div class="mt-8 text-center text-xs text-gray-600">
        价格均含税，按自然月结算 · 支持支付宝 / 微信 / 企业对公转账
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const tabs = [
  { key: 'all', label: '全部' },
  { key: 'entry', label: '入门型' },
  { key: 'standard', label: '标准型' },
  { key: 'performance', label: '高性能型' },
]

const activeTab = ref('all')

interface Spec { label: string; value: string }
interface Product {
  id: number; name: string; type: string; typeLabel: string
  desc: string; specs: Spec[]; price: number; featured: boolean
}

const products: Product[] = [
  {
    id: 1, name: 'EX-E1', type: 'entry', typeLabel: '入门型', featured: false,
    desc: '适合开发测试、小型网站',
    specs: [
      { label: '处理器', value: 'Intel Xeon E-2314 / 4C8T' },
      { label: '内存', value: '16GB ECC DDR4' },
      { label: '存储', value: '500GB SSD' },
      { label: '带宽', value: '100Mbps 独享' },
      { label: '流量', value: '5TB/月' },
    ],
    price: 299,
  },
  {
    id: 2, name: 'EX-E2', type: 'entry', typeLabel: '入门型', featured: false,
    desc: '适合游戏、低延迟场景',
    specs: [
      { label: '处理器', value: 'Intel Core i7-12700K / 8C16T' },
      { label: '内存', value: '32GB DDR5' },
      { label: '存储', value: '500GB NVMe' },
      { label: '带宽', value: '100Mbps 独享' },
      { label: '流量', value: '5TB/月' },
    ],
    price: 399,
  },
  {
    id: 3, name: 'AX-S1', type: 'standard', typeLabel: '标准型', featured: true,
    desc: '适合中小企业业务系统',
    specs: [
      { label: '处理器', value: 'Intel Xeon E-2388G / 8C16T' },
      { label: '内存', value: '32GB ECC DDR4' },
      { label: '存储', value: '1TB NVMe SSD' },
      { label: '带宽', value: '200Mbps 独享' },
      { label: '流量', value: '10TB/月' },
    ],
    price: 599,
  },
  {
    id: 4, name: 'AX-S2', type: 'standard', typeLabel: '标准型', featured: false,
    desc: '适合数据库、应用中间件',
    specs: [
      { label: '处理器', value: 'AMD EPYC 7302 / 16C32T' },
      { label: '内存', value: '64GB ECC DDR4' },
      { label: '存储', value: '2×1TB NVMe RAID1' },
      { label: '带宽', value: '500Mbps 独享' },
      { label: '流量', value: '20TB/月' },
    ],
    price: 899,
  },
  {
    id: 5, name: 'AX-P1', type: 'performance', typeLabel: '高性能型', featured: false,
    desc: '适合大数据、AI 计算',
    specs: [
      { label: '处理器', value: 'AMD EPYC 7453 / 28C56T' },
      { label: '内存', value: '128GB ECC DDR4' },
      { label: '存储', value: '4×2TB NVMe RAID10' },
      { label: '带宽', value: '1Gbps 独享' },
      { label: '流量', value: '不限' },
    ],
    price: 1899,
  },
  {
    id: 6, name: 'AX-P2', type: 'performance', typeLabel: '高性能型', featured: false,
    desc: '适合高并发、分布式集群',
    specs: [
      { label: '处理器', value: 'AMD EPYC 9654 / 96C192T' },
      { label: '内存', value: '256GB ECC DDR5' },
      { label: '存储', value: '6×4TB NVMe RAID10' },
      { label: '带宽', value: '10Gbps 独享' },
      { label: '流量', value: '不限' },
    ],
    price: 2999,
  },
]

const filteredProducts = computed(() =>
  activeTab.value === 'all' ? products : products.filter((p) => p.type === activeTab.value)
)
</script>
