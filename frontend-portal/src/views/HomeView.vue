<!-- 首页：组合所有区块 -->
<template>
  <div class="min-h-screen bg-white">
    <!-- 顶部通知条：深色背景，橙色强调，左右两条通知，移动端横向滚动 -->
    <div class="fixed top-0 left-0 right-0 z-[60] h-8 bg-[#0a1525] border-b border-white/10">
      <div class="h-full flex items-center overflow-x-auto scrollbar-none">
        <!-- 左侧通知 -->
        <a
          href="#"
          class="flex-1 min-w-max md:min-w-0 flex items-center justify-center gap-2 px-6 border-r border-white/10 h-full hover:bg-white/5 transition-colors duration-150"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-[#ff6a00] shrink-0"></span>
          <span class="text-gray-300 text-xs whitespace-nowrap">
            双十一促销：独立服务器首月享 8 折，下单即享优惠 →
          </span>
        </a>
        <!-- 右侧通知 -->
        <a
          href="#"
          class="flex-1 min-w-max md:min-w-0 flex items-center justify-center gap-2 px-6 h-full hover:bg-white/5 transition-colors duration-150"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-blue-400 shrink-0"></span>
          <span class="text-gray-300 text-xs whitespace-nowrap">
            华北节点全新上线：北京 BGP 线路，延迟低至 5ms →
          </span>
        </a>
      </div>
    </div>

    <!-- 导航栏：fixed，紧贴通知条下方 top-8 -->
    <NavBar />

    <!-- 内容区：顶部偏移 = 通知条(32px) + 导航栏(56px) = 88px -->
    <main class="pt-[88px]">
      <HeroSection />
      <ProductsSection />
      <FeaturesSection />
    </main>
    <FooterSection />
  </div>
</template>

<script setup lang="ts">
import NavBar from '@/components/home/NavBar.vue'
import HeroSection from '@/components/home/HeroSection.vue'
import ProductsSection from '@/components/home/ProductsSection.vue'
import FeaturesSection from '@/components/home/FeaturesSection.vue'
import FooterSection from '@/components/home/FooterSection.vue'
</script>
