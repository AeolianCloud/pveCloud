<!-- 导航栏：阿里云风格，深色背景，带二级导航结构 -->
<template>
  <header
    :class="[
      'fixed top-8 left-0 right-0 z-50 transition-all duration-200',
      scrolled
        ? 'bg-[#0e1829]/95 backdrop-blur-sm border-b border-white/5'
        : 'bg-[#0e1829]'
    ]"
  >
    <div class="max-w-[1200px] mx-auto px-6 h-14 flex items-center justify-between gap-8">
      <!-- Logo -->
      <a href="/" class="flex items-center gap-2.5 shrink-0">
        <svg class="w-6 h-6 text-[#ff6a00]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
        </svg>
        <span class="text-white font-bold text-base tracking-wide">pveCloud</span>
      </a>

      <!-- 导航链接（桌面端） -->
      <nav class="hidden lg:flex items-center gap-1 flex-1">
        <a
          v-for="link in navLinks"
          :key="link.label"
          :href="link.href"
          class="px-3 py-1.5 text-sm text-gray-300 hover:text-white hover:bg-white/5 rounded transition-colors duration-150 whitespace-nowrap"
        >
          {{ link.label }}
        </a>
      </nav>

      <!-- 右侧操作 -->
      <div class="flex items-center gap-3 shrink-0">
        <a href="/login" class="hidden md:inline text-sm text-gray-300 hover:text-white transition-colors duration-150">
          登录
        </a>
        <a
          href="/register"
          class="hidden md:inline px-4 py-1.5 rounded text-sm font-medium text-white bg-[#ff6a00] hover:bg-[#e85f00] transition-colors duration-150"
        >
          免费注册
        </a>
        <a
          href="#products"
          class="px-4 py-1.5 rounded text-sm font-medium text-[#ff6a00] border border-[#ff6a00] hover:bg-[#ff6a00]/10 transition-colors duration-150"
        >
          查看产品
        </a>
        <!-- 移动端菜单 -->
        <button class="lg:hidden p-1.5 text-gray-400 hover:text-white" @click="mobileMenuOpen = !mobileMenuOpen">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path v-if="!mobileMenuOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 移动端菜单 -->
    <div v-show="mobileMenuOpen" class="lg:hidden bg-[#0e1829] border-t border-white/10">
      <nav class="max-w-[1200px] mx-auto px-6 py-4 flex flex-col gap-1">
        <a
          v-for="link in navLinks"
          :key="link.label"
          :href="link.href"
          class="px-3 py-2 text-sm text-gray-300 hover:text-white hover:bg-white/5 rounded"
          @click="mobileMenuOpen = false"
        >
          {{ link.label }}
        </a>
        <div class="pt-3 mt-2 border-t border-white/10 flex gap-3">
          <a href="/login" class="text-sm text-gray-300">登录</a>
          <a href="/register" class="text-sm text-[#ff6a00]">免费注册</a>
        </div>
      </nav>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const navLinks = [
  { label: '独立服务器', href: '#products' },
  { label: '带宽 & 网络', href: '#network' },
  { label: '数据中心', href: '#datacenter' },
  { label: '解决方案', href: '#solutions' },
  { label: '定价', href: '#pricing' },
  { label: '帮助文档', href: '#docs' },
]

const scrolled = ref(false)
const mobileMenuOpen = ref(false)

function onScroll() {
  scrolled.value = window.scrollY > 10
}

onMounted(() => window.addEventListener('scroll', onScroll, { passive: true }))
onUnmounted(() => window.removeEventListener('scroll', onScroll))
</script>
