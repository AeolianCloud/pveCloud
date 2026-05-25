<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '../../stores/auth'

const authStore = useAuthStore()
const displayName = computed(() => authStore.user?.display_name || authStore.user?.username || '用户')
const email = computed(() => authStore.user?.email || '-')
const statusText = computed(() => authStore.user?.status === 'active' ? '正常' : authStore.user?.status || '-')
</script>

<template>
  <div class="page-reveal bg-white">
    <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div class="mb-10 flex flex-col justify-between gap-4 border-b border-neutral-200 pb-8 md:flex-row md:items-end">
        <div>
          <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">User Center</p>
          <h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">用户中心</h1>
        </div>
        <span class="rounded-full border border-neutral-950 px-4 py-2 text-xs font-black">展示骨架</span>
      </div>

      <div class="stagger-reveal grid gap-5 lg:grid-cols-3">
        <section class="interactive-card rounded-[1.5rem] border border-neutral-950 bg-white p-6 shadow-[8px_8px_0_#111]">
          <h2 class="text-lg font-black text-neutral-950">用户信息</h2>
          <div class="mt-6 flex items-center gap-4">
            <div class="flex h-14 w-14 items-center justify-center rounded-2xl border border-neutral-950 bg-neutral-950 text-xl font-black text-white">U</div>
            <div>
              <div class="text-sm font-black text-neutral-950">{{ displayName }}</div>
              <div class="text-sm text-neutral-500">{{ email }}</div>
            </div>
          </div>
          <div class="mt-6 grid gap-3 text-sm">
            <div class="flex justify-between border-t border-neutral-200 pt-3"><span class="text-neutral-500">账户状态</span><span class="font-black">{{ statusText }}</span></div>
            <div class="flex justify-between border-t border-neutral-200 pt-3"><span class="text-neutral-500">实名状态</span><span class="font-black">未实名</span></div>
          </div>
        </section>

        <section class="interactive-card rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-6">
          <h2 class="text-lg font-black text-neutral-950">快速入口</h2>
          <div class="mt-5 space-y-3">
            <RouterLink to="/user/profile" class="soft-lift block rounded-xl border border-neutral-200 bg-white px-4 py-3 text-sm font-black hover:border-neutral-950">账号资料</RouterLink>
            <RouterLink to="/user/real-name" class="soft-lift block rounded-xl border border-neutral-200 bg-white px-4 py-3 text-sm font-black hover:border-neutral-950">实名认证</RouterLink>
            <RouterLink to="/user/wallet" class="soft-lift block rounded-xl border border-neutral-200 bg-white px-4 py-3 text-sm font-black hover:border-neutral-950">我的钱包</RouterLink>
            <RouterLink to="/user/orders" class="soft-lift block rounded-xl border border-neutral-200 bg-white px-4 py-3 text-sm font-black hover:border-neutral-950">订单管理</RouterLink>
            <RouterLink to="/user/invoices" class="soft-lift block rounded-xl border border-neutral-200 bg-white px-4 py-3 text-sm font-black hover:border-neutral-950">发票管理</RouterLink>
            <RouterLink to="/user/instances" class="soft-lift block rounded-xl border border-neutral-200 bg-white px-4 py-3 text-sm font-black hover:border-neutral-950">实例管理</RouterLink>
            <RouterLink to="/user/tickets" class="soft-lift block rounded-xl border border-neutral-200 bg-white px-4 py-3 text-sm font-black hover:border-neutral-950">工单支持</RouterLink>
          </div>
        </section>

        <section class="interactive-card rounded-[1.5rem] border border-neutral-200 bg-white p-6">
          <h2 class="text-lg font-black text-neutral-950">账户概览</h2>
          <dl class="mt-6 space-y-4 text-sm">
            <div><dt class="text-neutral-500">最近登录</dt><dd class="mt-1 font-black text-neutral-950">2024-01-01 12:00:00</dd></div>
            <div><dt class="text-neutral-500">注册时间</dt><dd class="mt-1 font-black text-neutral-950">2024-01-01</dd></div>
            <div><dt class="text-neutral-500">订单入口</dt><dd class="mt-1"><RouterLink to="/user/orders" class="font-black text-neutral-950 underline">查看我的订单</RouterLink></dd></div>
            <div><dt class="text-neutral-500">发票入口</dt><dd class="mt-1"><RouterLink to="/user/invoices" class="font-black text-neutral-950 underline">查看我的发票</RouterLink></dd></div>
            <div><dt class="text-neutral-500">实例入口</dt><dd class="mt-1"><RouterLink to="/user/instances" class="font-black text-neutral-950 underline">查看我的实例</RouterLink></dd></div>
            <div><dt class="text-neutral-500">支持入口</dt><dd class="mt-1"><RouterLink to="/user/tickets" class="font-black text-neutral-950 underline">查看我的工单</RouterLink></dd></div>
          </dl>
        </section>
      </div>
    </div>
  </div>
</template>
