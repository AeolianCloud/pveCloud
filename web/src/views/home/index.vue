<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'

import TrafficPanel from './components/TrafficPanel.vue'

let revealObserver: IntersectionObserver | null = null

onMounted(() => {
  const revealSections = Array.from(document.querySelectorAll<HTMLElement>('.home-scroll-reveal'))

  if (!revealSections.length) {
    return
  }

  if (!('IntersectionObserver' in window)) {
    revealSections.forEach((section) => section.classList.add('is-visible'))
    return
  }

  revealObserver = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) {
          return
        }

        entry.target.classList.add('is-visible')
        revealObserver?.unobserve(entry.target)
      })
    },
    {
      rootMargin: '0px 0px -12% 0px',
      threshold: 0.08,
    },
  )

  revealSections.forEach((section) => revealObserver?.observe(section))
})

onBeforeUnmount(() => {
  revealObserver?.disconnect()
})
</script>

<template>
  <div class="page-reveal home-animated bg-white">
    <section class="relative overflow-hidden border-b border-neutral-200">
      <div class="absolute inset-x-0 top-0 h-px bg-neutral-950"></div>
      <div class="hero-orb hero-orb-left"></div>
      <div class="hero-orb hero-orb-right"></div>
      <div class="mx-auto grid max-w-7xl gap-12 px-4 py-20 sm:px-6 lg:grid-cols-[1.05fr_0.95fr] lg:px-8 lg:py-28">
        <div class="hero-stack relative z-10">
          <div class="chip-hover mb-8 inline-flex items-center gap-2 rounded-full border border-neutral-950 px-4 py-2 text-xs font-black uppercase tracking-[0.2em] text-neutral-950">
            <span class="pulse-dot"></span>
            Game Cloud / High Frequency Nodes
          </div>
          <h1 class="hero-title max-w-4xl text-5xl font-black leading-[0.95] tracking-tight text-neutral-950 sm:text-6xl lg:text-7xl">
            给游戏开服准备的轻量云平台
          </h1>
          <p class="hero-copy mt-8 max-w-2xl text-lg leading-8 text-neutral-600">
            面向 Minecraft、Steam 游戏服、联机社区和独立项目展示高主频计算、大带宽节点与基础防护能力。当前订单用于提交购买意向，由后台人工处理。
          </p>
          <div class="hero-actions mt-10 flex flex-col gap-3 sm:flex-row">
            <RouterLink to="/products" class="btn-dark inline-flex items-center justify-center rounded-full border px-7 py-3 text-sm font-black">
              查看产品配置
            </RouterLink>
            <RouterLink to="/register" class="inline-flex items-center justify-center rounded-full border border-neutral-300 bg-white px-7 py-3 text-sm font-black text-neutral-950 hover:border-neutral-950 hover:bg-neutral-50 hover:-translate-y-px">
              注册账号
            </RouterLink>
          </div>
          <div class="hero-mini-grid mt-10 grid gap-3 sm:grid-cols-3">
            <div class="rounded-2xl border border-neutral-200 bg-white/85 p-4 backdrop-blur">
              <div class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">响应</div>
              <div class="mt-2 text-2xl font-black text-neutral-950">毫秒级</div>
            </div>
            <div class="rounded-2xl border border-neutral-200 bg-white/85 p-4 backdrop-blur">
              <div class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">节点</div>
              <div class="mt-2 text-2xl font-black text-neutral-950">多区域</div>
            </div>
            <div class="rounded-2xl border border-neutral-200 bg-white/85 p-4 backdrop-blur">
              <div class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">状态</div>
              <div class="mt-2 text-2xl font-black text-neutral-950">持续在线</div>
            </div>
          </div>
        </div>

        <div class="surface-pop hero-dashboard relative rounded-[2rem] border border-neutral-950 bg-white p-4 shadow-[12px_12px_0_#111]">
          <div class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5">
            <div class="flex items-center justify-between border-b border-neutral-200 pb-4">
              <div>
                <div class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Live Node Preview</div>
                <div class="mt-1 text-xl font-black text-neutral-950">CN-HZ-GAME-01</div>
              </div>
              <span class="status-pill rounded-full border border-neutral-950 bg-white px-3 py-1 text-xs font-black">在线</span>
            </div>
            <div class="mt-5 grid grid-cols-2 gap-3">
              <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                <div class="text-xs text-neutral-500">CPU</div>
                <div class="mt-2 text-2xl font-black">5.2GHz</div>
                <div class="mt-3 h-1.5 overflow-hidden rounded-full bg-neutral-100"><div class="meter-bar w-[72%]"></div></div>
              </div>
              <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                <div class="text-xs text-neutral-500">带宽</div>
                <div class="mt-2 text-2xl font-black">100M</div>
                <div class="mt-3 h-1.5 overflow-hidden rounded-full bg-neutral-100"><div class="meter-bar meter-bar-delayed w-[86%]"></div></div>
              </div>
              <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                <div class="text-xs text-neutral-500">防护</div>
                <div class="mt-2 text-2xl font-black">基础</div>
                <div class="mt-3 h-1.5 overflow-hidden rounded-full bg-neutral-100"><div class="meter-bar meter-bar-late w-[64%]"></div></div>
              </div>
              <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                <div class="text-xs text-neutral-500">部署</div>
                <div class="mt-2 text-2xl font-black">分钟级</div>
                <div class="mt-3 h-1.5 overflow-hidden rounded-full bg-neutral-100"><div class="meter-bar meter-bar-delayed w-[92%]"></div></div>
              </div>
            </div>
            <div class="mt-5 rounded-2xl border border-neutral-950 bg-white p-5">
              <div class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Recommended For</div>
              <div class="mt-3 grid gap-2 text-sm text-neutral-950">
                <div class="flex justify-between gap-4"><span class="text-neutral-600">Minecraft 生存服</span><span class="font-black">4C / 8G</span></div>
                <div class="flex justify-between gap-4"><span class="text-neutral-600">幻兽帕鲁联机</span><span class="font-black">8C / 16G</span></div>
                <div class="flex justify-between gap-4"><span class="text-neutral-600">通用轻量业务</span><span class="font-black">2C / 4G</span></div>
              </div>
            </div>
          </div>
          <TrafficPanel />
        </div>
      </div>
    </section>

    <section class="home-scroll-reveal mx-auto max-w-7xl px-4 py-18 sm:px-6 lg:px-8">
      <div class="grid gap-10 lg:grid-cols-[0.9fr_1.1fr] lg:items-end">
        <div>
          <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Scenarios</p>
          <h2 class="mt-3 text-3xl font-black tracking-tight text-neutral-950 sm:text-4xl">把开服前的关键选择摊开看</h2>
          <p class="mt-4 text-sm leading-7 text-neutral-600">
            首页先帮你判断场景、配置和网络侧重点。真正提交订单前，仍以产品中心和订单确认页返回的价格、地域、系统模板为准。
          </p>
        </div>
        <div class="grid gap-4 sm:grid-cols-3">
          <div class="soft-lift rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5">
            <div class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">低延迟</div>
            <h3 class="mt-4 text-xl font-black text-neutral-950">联机房间</h3>
            <p class="mt-3 text-sm leading-6 text-neutral-600">适合好友联机、语音社区和小规模活动服，优先关注线路和响应。</p>
          </div>
          <div class="soft-lift rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5">
            <div class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">高主频</div>
            <h3 class="mt-4 text-xl font-black text-neutral-950">模组生存服</h3>
            <p class="mt-3 text-sm leading-6 text-neutral-600">适合 Minecraft 模组、插件服和 Tick 敏感业务，优先关注 CPU 与内存。</p>
          </div>
          <div class="soft-lift rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5">
            <div class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">大带宽</div>
            <h3 class="mt-4 text-xl font-black text-neutral-950">活动节点</h3>
            <p class="mt-3 text-sm leading-6 text-neutral-600">适合短期活动、公会赛事和下载分发，优先关注带宽与地域覆盖。</p>
          </div>
        </div>
      </div>
    </section>

    <section class="home-scroll-reveal mx-auto max-w-7xl px-4 py-18 sm:px-6 lg:px-8">
      <div class="flex flex-col justify-between gap-6 border-b border-neutral-200 pb-8 md:flex-row md:items-end">
        <div>
          <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Products</p>
          <h2 class="mt-3 text-3xl font-black tracking-tight text-neutral-950 sm:text-4xl">为开服场景拆好的产品骨架</h2>
        </div>
        <RouterLink to="/products" class="text-sm font-black text-neutral-950 underline decoration-2 underline-offset-4">进入产品中心</RouterLink>
      </div>

      <div class="stagger-reveal mt-10 grid gap-5 lg:grid-cols-3">
        <div class="interactive-card rounded-[1.75rem] border border-neutral-950 bg-white p-7 shadow-[8px_8px_0_#111]">
          <div class="text-sm font-black text-neutral-500">01</div>
          <h3 class="mt-6 text-2xl font-black text-neutral-950">游戏云服务器</h3>
          <p class="mt-3 text-sm leading-6 text-neutral-600">适合 Minecraft、Steam 游戏服、语音社区和轻量业务。</p>
          <ul class="mt-6 space-y-3 text-sm font-semibold text-neutral-800">
            <li>高主频 CPU</li>
            <li>SSD 系统盘</li>
            <li>独立公网 IP</li>
          </ul>
        </div>
        <div class="interactive-card rounded-[1.75rem] border border-neutral-200 bg-neutral-50 p-7">
          <div class="text-sm font-black text-neutral-500">02</div>
          <h3 class="mt-6 text-2xl font-black text-neutral-950">高防与大带宽</h3>
          <p class="mt-3 text-sm leading-6 text-neutral-600">用于联机高峰、活动流量和基础抗攻击展示场景。</p>
          <ul class="mt-6 space-y-3 text-sm font-semibold text-neutral-800">
            <li>多线节点</li>
            <li>带宽可选</li>
            <li>基础防护说明</li>
          </ul>
        </div>
        <div class="interactive-card rounded-[1.75rem] border border-neutral-200 bg-neutral-50 p-7">
          <div class="text-sm font-black text-neutral-500">03</div>
          <h3 class="mt-6 text-2xl font-black text-neutral-950">物理服务器</h3>
          <p class="mt-3 text-sm leading-6 text-neutral-600">适合多人社区、私有化环境和更高性能的独享资源。</p>
          <ul class="mt-6 space-y-3 text-sm font-semibold text-neutral-800">
            <li>独享硬件</li>
            <li>远程管理</li>
            <li>定制咨询</li>
          </ul>
        </div>
      </div>
    </section>

    <section class="home-scroll-reveal border-y border-neutral-200 bg-neutral-950 text-white">
      <div class="mx-auto grid max-w-7xl gap-8 px-4 py-18 sm:px-6 lg:grid-cols-[1fr_1.2fr] lg:px-8">
        <div>
          <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-400">How It Works</p>
          <h2 class="mt-3 text-3xl font-black tracking-tight sm:text-4xl">从看配置到提交意向，流程更短</h2>
          <p class="mt-4 text-sm leading-7 text-neutral-400">
            当前阶段不展示支付、自动开通或实例交付进度。订单提交后进入后台处理，后续状态以订单详情为准。
          </p>
        </div>
        <div class="grid gap-3 sm:grid-cols-2">
          <div class="rounded-[1.5rem] border border-white/15 bg-white/8 p-5">
            <div class="text-3xl font-black text-white/35">01</div>
            <h3 class="mt-4 text-xl font-black">浏览产品</h3>
            <p class="mt-3 text-sm leading-6 text-white/60">查看套餐、周期价格、销售地域、系统模板和网络类型。</p>
          </div>
          <div class="rounded-[1.5rem] border border-white/15 bg-white/8 p-5">
            <div class="text-3xl font-black text-white/35">02</div>
            <h3 class="mt-4 text-xl font-black">确认配置</h3>
            <p class="mt-3 text-sm leading-6 text-white/60">在产品页选择周期、地域、系统和网络后，再进入订单创建。</p>
          </div>
          <div class="rounded-[1.5rem] border border-white/15 bg-white/8 p-5">
            <div class="text-3xl font-black text-white/35">03</div>
            <h3 class="mt-4 text-xl font-black">账号与实名</h3>
            <p class="mt-3 text-sm leading-6 text-white/60">未登录会先跳转登录；如后台要求实名，需要先完成认证。</p>
          </div>
          <div class="rounded-[1.5rem] border border-white/15 bg-white/8 p-5">
            <div class="text-3xl font-black text-white/35">04</div>
            <h3 class="mt-4 text-xl font-black">提交订单</h3>
            <p class="mt-3 text-sm leading-6 text-white/60">订单代表购买意向，可在用户中心查看详情或取消待处理订单。</p>
          </div>
        </div>
      </div>
    </section>

    <section class="home-scroll-reveal mx-auto max-w-7xl px-4 py-18 sm:px-6 lg:px-8">
      <div class="grid gap-5 lg:grid-cols-3">
        <div class="surface-pop rounded-[2rem] border border-neutral-950 bg-white p-7 shadow-[8px_8px_0_#111] lg:col-span-1">
          <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Compare</p>
          <h2 class="mt-3 text-3xl font-black tracking-tight text-neutral-950">按关注点选产品</h2>
          <p class="mt-4 text-sm leading-7 text-neutral-600">首页只做快速理解，完整可售套餐和金额以产品中心接口返回为准。</p>
          <RouterLink to="/products" class="mt-8 inline-flex rounded-full border border-neutral-950 px-6 py-3 text-sm font-black text-neutral-950 hover:bg-neutral-950 hover:text-white">
            查看全部产品
          </RouterLink>
        </div>
        <div class="grid gap-4 sm:grid-cols-2 lg:col-span-2">
          <div class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5">
            <div class="flex items-center justify-between gap-4">
              <h3 class="font-black text-neutral-950">CPU 密集</h3>
              <span class="rounded-full border border-neutral-300 bg-white px-3 py-1 text-xs font-black text-neutral-500">Tick</span>
            </div>
            <p class="mt-4 text-sm leading-6 text-neutral-600">适合高 Tick、插件较多、对单核性能敏感的游戏服。</p>
          </div>
          <div class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5">
            <div class="flex items-center justify-between gap-4">
              <h3 class="font-black text-neutral-950">内存容量</h3>
              <span class="rounded-full border border-neutral-300 bg-white px-3 py-1 text-xs font-black text-neutral-500">RAM</span>
            </div>
            <p class="mt-4 text-sm leading-6 text-neutral-600">适合模组整合包、多人在线和需要更大缓存空间的场景。</p>
          </div>
          <div class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5">
            <div class="flex items-center justify-between gap-4">
              <h3 class="font-black text-neutral-950">网络线路</h3>
              <span class="rounded-full border border-neutral-300 bg-white px-3 py-1 text-xs font-black text-neutral-500">Route</span>
            </div>
            <p class="mt-4 text-sm leading-6 text-neutral-600">适合跨地域玩家接入，重点关注地域、网络类型和带宽。</p>
          </div>
          <div class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5">
            <div class="flex items-center justify-between gap-4">
              <h3 class="font-black text-neutral-950">系统模板</h3>
              <span class="rounded-full border border-neutral-300 bg-white px-3 py-1 text-xs font-black text-neutral-500">OS</span>
            </div>
            <p class="mt-4 text-sm leading-6 text-neutral-600">按熟悉的系统环境选择模板，减少后续部署和维护成本。</p>
          </div>
        </div>
      </div>
    </section>

    <section class="home-scroll-reveal border-y border-neutral-200 bg-neutral-50">
      <div class="mx-auto grid max-w-7xl grid-cols-2 gap-px bg-neutral-200 px-4 sm:px-6 md:grid-cols-4 lg:px-8">
        <div class="soft-lift bg-neutral-50 py-10 text-center">
          <div class="text-4xl font-black text-neutral-950">5.2G</div>
          <div class="mt-2 text-sm text-neutral-500">高主频展示</div>
        </div>
        <div class="soft-lift bg-neutral-50 py-10 text-center">
          <div class="text-4xl font-black text-neutral-950">100M+</div>
          <div class="mt-2 text-sm text-neutral-500">带宽套餐</div>
        </div>
        <div class="soft-lift bg-neutral-50 py-10 text-center">
          <div class="text-4xl font-black text-neutral-950">4+</div>
          <div class="mt-2 text-sm text-neutral-500">游戏场景</div>
        </div>
        <div class="soft-lift bg-neutral-50 py-10 text-center">
          <div class="text-4xl font-black text-neutral-950">24/7</div>
          <div class="mt-2 text-sm text-neutral-500">服务支持</div>
        </div>
      </div>
    </section>

    <section class="home-scroll-reveal mx-auto max-w-7xl px-4 py-20 sm:px-6 lg:px-8">
      <div class="surface-pop rounded-[2rem] border border-neutral-950 bg-white p-8 shadow-[10px_10px_0_#111] md:p-12">
        <div class="grid gap-8 md:grid-cols-[1fr_auto] md:items-center">
          <div>
            <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Ready</p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-neutral-950 sm:text-4xl">先选配置，再提交订单意向</h2>
            <p class="mt-4 max-w-2xl text-sm leading-6 text-neutral-600">当前阶段开放产品展示、账号入口和订单 MVP，不展示支付、实例或自动交付承诺。</p>
          </div>
          <RouterLink to="/register" class="btn-dark inline-flex justify-center rounded-full border px-7 py-3 text-sm font-black">
            注册
          </RouterLink>
        </div>
      </div>
    </section>
  </div>
</template>
