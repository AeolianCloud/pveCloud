<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="min-h-screen bg-white text-neutral-950">
    <header class="sticky top-0 z-40 border-b border-neutral-200 bg-white/95 backdrop-blur">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex h-16 items-center justify-between">
          <RouterLink to="/" class="flex items-center gap-3">
            <span class="flex h-9 w-9 items-center justify-center rounded-xl border border-neutral-950 bg-white">
              <svg class="h-5 w-5 text-neutral-950" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
            </svg>
            </span>
            <span class="text-lg font-black tracking-tight">PVECloud</span>
          </RouterLink>
          <nav class="hidden items-center gap-8 md:flex">
            <RouterLink to="/products" class="text-sm font-semibold text-neutral-600 hover:text-neutral-950">
              产品中心
            </RouterLink>
            <RouterLink to="/pricing" class="text-sm font-semibold text-neutral-600 hover:text-neutral-950">
              价格方案
            </RouterLink>
            <RouterLink to="/login" class="text-sm font-semibold text-neutral-600 hover:text-neutral-950">
              登录
            </RouterLink>
            <RouterLink
              to="/register"
              class="btn-dark rounded-full border px-5 py-2 text-sm font-bold"
            >
              免费注册
            </RouterLink>
          </nav>
        </div>
      </div>
    </header>

    <main class="min-h-[calc(100vh-4rem)]">
      <RouterView />
    </main>

    <footer class="mt-20 border-t border-neutral-200 bg-white">
      <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div class="grid gap-8 md:grid-cols-4">
          <div>
            <div class="mb-4 flex items-center gap-2 text-neutral-950">
              <span class="flex h-7 w-7 items-center justify-center rounded-lg border border-neutral-950 text-xs font-black">P</span>
              <span class="font-black">PVECloud</span>
            </div>
            <p class="max-w-xs text-sm leading-6 text-neutral-500">面向游戏开服、轻量应用和独立开发者的云服务器展示站。</p>
          </div>
          <div>
            <h4 class="mb-4 text-sm font-black text-neutral-950">产品</h4>
            <ul class="space-y-2 text-sm text-neutral-500">
              <li><RouterLink to="/products" class="hover:text-neutral-950">游戏云服务器</RouterLink></li>
              <li><RouterLink to="/products" class="hover:text-neutral-950">高主频实例</RouterLink></li>
              <li><RouterLink to="/products" class="hover:text-neutral-950">大带宽节点</RouterLink></li>
            </ul>
          </div>
          <div>
            <h4 class="mb-4 text-sm font-black text-neutral-950">支持</h4>
            <ul class="space-y-2 text-sm text-neutral-500">
              <li><a href="#" class="hover:text-neutral-950">帮助中心</a></li>
              <li><a href="#" class="hover:text-neutral-950">工单系统</a></li>
              <li><a href="#" class="hover:text-neutral-950">联系客服</a></li>
            </ul>
          </div>
          <div>
            <h4 class="mb-4 text-sm font-black text-neutral-950">联系</h4>
            <ul class="space-y-2 text-sm text-neutral-500">
              <li>客服电话：400-888-8888</li>
              <li>客服邮箱：support@pvecloud.com</li>
            </ul>
          </div>
        </div>
        <div class="mt-8 border-t border-neutral-200 pt-8 text-center text-sm text-neutral-500">
          © 2024 PVECloud. All rights reserved.
        </div>
      </div>
    </footer>
  </div>
</template>
