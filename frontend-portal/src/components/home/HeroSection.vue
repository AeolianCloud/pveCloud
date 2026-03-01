<!-- Hero 区块：阿里云风格，深蓝渐变背景，左文案+右数据卡片 -->
<template>
  <section class="relative overflow-hidden bg-[#0e1829] pt-16 pb-20">
    <!-- 背景渐变装饰（克制，不过分） -->
    <div class="absolute inset-0 pointer-events-none">
      <!-- 左上角蓝色光晕 -->
      <div class="absolute -top-20 -left-20 w-96 h-96 bg-blue-600/10 rounded-full blur-3xl"></div>
      <!-- 右下角橙色微光（阿里云橙） -->
      <div class="absolute bottom-0 right-0 w-80 h-80 bg-[#ff6a00]/5 rounded-full blur-3xl"></div>
      <!-- 网格底纹 -->
      <div class="absolute inset-0 opacity-[0.03]"
        style="background-image: linear-gradient(#fff 1px, transparent 1px), linear-gradient(90deg, #fff 1px, transparent 1px); background-size: 40px 40px;">
      </div>
    </div>

    <div class="relative max-w-[1200px] mx-auto px-6">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">

        <!-- 左侧：标题 + 卖点 + CTA -->
        <div>
          <!-- 标签 -->
          <div class="inline-flex items-center gap-2 px-3 py-1 mb-6 rounded-full border border-blue-500/30 bg-blue-500/10">
            <span class="w-1.5 h-1.5 rounded-full bg-[#ff6a00]"></span>
            <span class="text-xs text-blue-300">企业级 IDC 裸金属服务</span>
          </div>

          <h1 class="text-4xl lg:text-5xl font-bold text-white leading-tight mb-5">
            高性能独立服务器<br />
            <span class="text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-cyan-300">
              BGP 多线 · 4h 交付
            </span>
          </h1>

          <p class="text-gray-400 text-base leading-relaxed mb-8">
            自建华北、华东、华南三大 IDC，全系 Intel/AMD 处理器，企业级 NVMe SSD，
            SLA 99.9% 可用性保障。
          </p>

          <!-- 卖点列表 -->
          <ul class="space-y-3 mb-10">
            <li
              v-for="point in sellingPoints"
              :key="point"
              class="flex items-center gap-3 text-sm text-gray-300"
            >
              <svg class="w-4 h-4 text-[#ff6a00] shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
              {{ point }}
            </li>
          </ul>

          <!-- CTA 按钮组 -->
          <div class="flex flex-wrap gap-4">
            <a
              href="#products"
              class="px-6 py-3 rounded text-sm font-semibold text-white bg-[#ff6a00] hover:bg-[#e85f00] transition-colors duration-150 shadow-lg shadow-[#ff6a00]/20"
            >
              立即选购
            </a>
            <a
              href="#features"
              class="px-6 py-3 rounded text-sm font-semibold text-white border border-white/20 hover:border-white/40 hover:bg-white/5 transition-colors duration-150"
            >
              了解详情
            </a>
          </div>
        </div>

        <!-- 右侧：指标卡片网格 -->
        <div class="grid grid-cols-2 gap-4">
          <div
            v-for="stat in stats"
            :key="stat.label"
            class="rounded-lg border border-white/10 bg-white/5 p-5 hover:bg-white/8 hover:border-white/20 transition-all duration-150"
          >
            <!-- 顶部图标 -->
            <div class="w-8 h-8 rounded bg-blue-500/20 flex items-center justify-center mb-4">
              <component :is="stat.icon" class="w-4 h-4 text-blue-400" />
            </div>
            <div class="text-2xl font-bold text-white mb-1">{{ stat.value }}</div>
            <div class="text-sm font-medium text-gray-300 mb-1">{{ stat.label }}</div>
            <div class="text-xs text-gray-500 leading-relaxed">{{ stat.desc }}</div>
          </div>
        </div>

      </div>

      <!-- 底部品牌带（客户 logo 等，占位） -->
      <div class="mt-16 pt-8 border-t border-white/10">
        <p class="text-xs text-gray-600 text-center mb-5">已服务超过 2,000+ 企业客户</p>
        <div class="flex flex-wrap justify-center items-center gap-8 opacity-30">
          <span v-for="brand in brands" :key="brand" class="text-gray-400 text-sm font-medium tracking-wider">
            {{ brand }}
          </span>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { defineComponent, h } from 'vue'

// 内联 SVG 图标
const makeIcon = (d: string) =>
  defineComponent({
    render: () =>
      h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
        h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d }),
      ]),
  })

const sellingPoints = [
  '独享物理资源，无虚拟化损耗，性能不缩水',
  'BGP 多线接入，联通 / 移动 / 电信全覆盖',
  '在线下单，4 小时内完成装机并交付',
  'OOB 带外管理，远程 KVM，故障自助处理',
]

const stats = [
  {
    icon: makeIcon('M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z'),
    value: '99.9%',
    label: 'SLA 可用性',
    desc: '年度网络及电力保障承诺',
  },
  {
    icon: makeIcon('M13 10V3L4 14h7v7l9-11h-7z'),
    value: '≤4h',
    label: '极速交付',
    desc: '工作日下单，4 小时内开机',
  },
  {
    icon: makeIcon('M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z'),
    value: '10Gbps',
    label: '上联带宽',
    desc: '机房出口带宽，多路由冗余',
  },
  {
    icon: makeIcon('M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z'),
    value: '7×24h',
    label: '运维响应',
    desc: '平均工单响应时间 15 分钟',
  },
]

// 品牌占位（待替换为真实客户）
const brands = ['某科技集团', '某互联网公司', '某金融机构', '某电商平台', '某游戏公司']
</script>
