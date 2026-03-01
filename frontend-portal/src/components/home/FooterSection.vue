<!-- 页脚：阿里云风格，深色，多栏链接 -->
<template>
  <footer class="bg-[#080f1a] border-t border-white/10">
    <div class="max-w-[1200px] mx-auto px-6 py-12">
      <div class="grid grid-cols-2 md:grid-cols-5 gap-8 mb-10">
        <!-- 品牌 -->
        <div class="col-span-2 md:col-span-1">
          <div class="flex items-center gap-2 mb-4">
            <svg class="w-5 h-5 text-[#ff6a00]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2" />
            </svg>
            <span class="text-white font-bold text-sm">pveCloud</span>
          </div>
          <p class="text-xs text-gray-600 leading-relaxed mb-4">
            自建 IDC 数据中心，提供裸金属服务器租用和企业托管服务。
          </p>
          <!-- 客服热线 -->
          <div class="text-xs text-gray-500">
            <div class="mb-1">销售热线</div>
            <div class="text-base font-bold text-white">400-800-0000</div>
            <div class="text-xs text-gray-600 mt-0.5">工作日 9:00–18:00</div>
          </div>
        </div>

        <!-- 链接组 -->
        <div v-for="group in linkGroups" :key="group.title">
          <h4 class="text-xs font-semibold text-gray-300 mb-4 uppercase tracking-wider">
            {{ group.title }}
          </h4>
          <ul class="space-y-2.5">
            <li v-for="link in group.links" :key="link.label">
              <a :href="link.href" class="text-xs text-gray-500 hover:text-gray-300 transition-colors duration-150">
                {{ link.label }}
              </a>
            </li>
          </ul>
        </div>
      </div>

      <!-- 底部版权 -->
      <div class="pt-6 border-t border-white/10 flex flex-col sm:flex-row justify-between items-center gap-3">
        <p class="text-xs text-gray-600">© 2025 pveCloud Technology Co., Ltd. 保留所有权利</p>
        <div class="flex gap-5">
          <a href="#" class="text-xs text-gray-600 hover:text-gray-400 transition-colors duration-150">隐私政策</a>
          <a href="#" class="text-xs text-gray-600 hover:text-gray-400 transition-colors duration-150">服务条款</a>
          <a href="#" class="text-xs text-gray-600 hover:text-gray-400 transition-colors duration-150">ICP 备 XXXXXXXX 号</a>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
const linkGroups = [
  {
    title: '产品',
    links: [
      { label: '独立服务器', href: '#products' },
      { label: '带宽服务', href: '#' },
      { label: '企业托管', href: '#' },
      { label: '定制方案', href: '#' },
    ],
  },
  {
    title: '解决方案',
    links: [
      { label: '游戏行业', href: '#' },
      { label: '金融行业', href: '#' },
      { label: '电商行业', href: '#' },
      { label: 'AI 计算', href: '#' },
    ],
  },
  {
    title: '支持',
    links: [
      { label: '技术文档', href: '#' },
      { label: '工单系统', href: '#' },
      { label: '网络状态', href: '#' },
      { label: '常见问题', href: '#' },
    ],
  },
  {
    title: '公司',
    links: [
      { label: '关于我们', href: '#' },
      { label: '数据中心', href: '#' },
      { label: '合作伙伴', href: '#' },
      { label: '加入我们', href: '#' },
    ],
  },
]
</script>
