<!-- 核心优势：阿里云风格，深灰背景，左标题+右六宫格 -->
<template>
  <section id="features" class="bg-[#0e1829] py-20">
    <div class="max-w-[1200px] mx-auto px-6">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-12 items-start">

        <!-- 左侧：标题区 -->
        <div class="lg:col-span-1">
          <div class="inline-block px-3 py-1 rounded-full border border-blue-500/30 bg-blue-500/10 text-xs text-blue-300 mb-5">
            核心优势
          </div>
          <h2 class="text-3xl font-bold text-white mb-4 leading-snug">
            为什么选择<br />pveCloud
          </h2>
          <p class="text-gray-400 text-sm leading-relaxed">
            深耕 IDC 行业十年，全自建基础设施，无中间商。从采购、运维到交付，全程掌控。
          </p>
        </div>

        <!-- 右侧：6 个特性，两列网格 -->
        <div class="lg:col-span-2 grid grid-cols-1 sm:grid-cols-2 gap-5">
          <div
            v-for="feature in features"
            :key="feature.title"
            class="flex gap-4 p-5 rounded-xl border border-white/8 bg-white/3 hover:border-white/15 hover:bg-white/5 transition-all duration-200"
          >
            <!-- 图标 -->
            <div class="w-10 h-10 rounded-lg bg-blue-500/15 border border-blue-500/20 flex items-center justify-center shrink-0">
              <component :is="feature.icon" class="w-5 h-5 text-blue-400" />
            </div>
            <div>
              <h3 class="text-sm font-semibold text-white mb-1.5">{{ feature.title }}</h3>
              <p class="text-xs text-gray-500 leading-relaxed">{{ feature.desc }}</p>
            </div>
          </div>
        </div>

      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { defineComponent, h } from 'vue'

const makeIcon = (d: string) =>
  defineComponent({
    render: () =>
      h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
        h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d }),
      ]),
  })

const features = [
  {
    icon: makeIcon('M13 10V3L4 14h7v7l9-11h-7z'),
    title: '4 小时极速交付',
    desc: '在线下单后，工作日 4 小时内完成装机并推送账号，无需等待人工审核。',
  },
  {
    icon: makeIcon('M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z'),
    title: 'SLA 99.9% 保障',
    desc: '网络和电力双路由冗余，年度不可用时间不超过 8.76 小时，否则按比例赔偿。',
  },
  {
    icon: makeIcon('M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z'),
    title: 'BGP 多线接入',
    desc: '直连中国电信、联通、移动骨干网，全国平均延迟 <30ms，国内访问无障碍。',
  },
  {
    icon: makeIcon('M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01'),
    title: '裸金属独享',
    desc: '物理资源完全独占，无邻居干扰，CPU 不降频，适合高 I/O 和计算密集型业务。',
  },
  {
    icon: makeIcon('M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z'),
    title: 'OOB 带外管理',
    desc: '每台服务器配备独立 IPMI / iDRAC，系统崩溃后可远程 KVM 接管，无需现场。',
  },
  {
    icon: makeIcon('M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z'),
    title: '7×24h 运维',
    desc: '驻场运维团队全天值守，工单平均响应 <15 分钟，紧急故障优先处理通道。',
  },
]
</script>
