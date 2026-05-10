<template>
  <div class="bg-white">
    <section class="relative overflow-hidden border-b border-neutral-200">
      <div class="absolute inset-x-0 top-0 h-px bg-neutral-950"></div>
      <div class="mx-auto grid max-w-7xl gap-12 px-4 py-20 sm:px-6 lg:grid-cols-[1.05fr_0.95fr] lg:px-8 lg:py-28">
        <div>
          <div class="mb-8 inline-flex items-center gap-2 rounded-full border border-neutral-950 px-4 py-2 text-xs font-black uppercase tracking-[0.2em] text-neutral-950">
            Game Cloud / High Frequency Nodes
          </div>
          <h1 class="max-w-4xl text-5xl font-black leading-[0.95] tracking-tight text-neutral-950 sm:text-6xl lg:text-7xl">
            给游戏开服准备的轻量云平台
          </h1>
          <p class="mt-8 max-w-2xl text-lg leading-8 text-neutral-600">
            面向 Minecraft、Steam 游戏服、联机社区和独立项目展示高主频计算、大带宽节点与基础防护能力。当前仅展示产品入口，购买能力即将开放。
          </p>
          <div class="mt-10 flex flex-col gap-3 sm:flex-row">
            <RouterLink to="/products" class="btn-dark inline-flex items-center justify-center rounded-full border px-7 py-3 text-sm font-black">
              查看产品配置
            </RouterLink>
            <RouterLink to="/pricing" class="inline-flex items-center justify-center rounded-full border border-neutral-300 bg-white px-7 py-3 text-sm font-black text-neutral-950 hover:border-neutral-950 hover:bg-neutral-50">
              对比价格方案
            </RouterLink>
          </div>
        </div>

        <div class="rounded-[2rem] border border-neutral-950 bg-white p-4 shadow-[12px_12px_0_#111]">
          <div class="rounded-[1.5rem] border border-neutral-200 bg-neutral-50 p-5">
            <div class="flex items-center justify-between border-b border-neutral-200 pb-4">
              <div>
                <div class="text-xs font-black uppercase tracking-[0.18em] text-neutral-500">Live Node Preview</div>
                <div class="mt-1 text-xl font-black text-neutral-950">CN-HZ-GAME-01</div>
              </div>
              <span class="rounded-full border border-neutral-950 bg-white px-3 py-1 text-xs font-black">在线</span>
            </div>
            <div class="mt-5 grid grid-cols-2 gap-3">
              <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                <div class="text-xs text-neutral-500">CPU</div>
                <div class="mt-2 text-2xl font-black">5.2GHz</div>
              </div>
              <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                <div class="text-xs text-neutral-500">带宽</div>
                <div class="mt-2 text-2xl font-black">100M</div>
              </div>
              <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                <div class="text-xs text-neutral-500">防护</div>
                <div class="mt-2 text-2xl font-black">基础</div>
              </div>
              <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                <div class="text-xs text-neutral-500">部署</div>
                <div class="mt-2 text-2xl font-black">分钟级</div>
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
        </div>
      </div>
    </section>

    <section class="mx-auto max-w-7xl px-4 py-18 sm:px-6 lg:px-8">
      <div class="flex flex-col justify-between gap-6 border-b border-neutral-200 pb-8 md:flex-row md:items-end">
        <div>
          <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Products</p>
          <h2 class="mt-3 text-3xl font-black tracking-tight text-neutral-950 sm:text-4xl">为开服场景拆好的产品骨架</h2>
        </div>
        <RouterLink to="/products" class="text-sm font-black text-neutral-950 underline decoration-2 underline-offset-4">进入产品中心</RouterLink>
      </div>

      <div class="mt-10 grid gap-5 lg:grid-cols-3">
        <div class="rounded-[1.75rem] border border-neutral-950 bg-white p-7 shadow-[8px_8px_0_#111]">
          <div class="text-sm font-black text-neutral-500">01</div>
          <h3 class="mt-6 text-2xl font-black text-neutral-950">游戏云服务器</h3>
          <p class="mt-3 text-sm leading-6 text-neutral-600">适合 Minecraft、Steam 游戏服、语音社区和轻量业务。</p>
          <ul class="mt-6 space-y-3 text-sm font-semibold text-neutral-800">
            <li>高主频 CPU</li>
            <li>SSD 系统盘</li>
            <li>独立公网 IP</li>
          </ul>
        </div>
        <div class="rounded-[1.75rem] border border-neutral-200 bg-neutral-50 p-7">
          <div class="text-sm font-black text-neutral-500">02</div>
          <h3 class="mt-6 text-2xl font-black text-neutral-950">高防与大带宽</h3>
          <p class="mt-3 text-sm leading-6 text-neutral-600">用于联机高峰、活动流量和基础抗攻击展示场景。</p>
          <ul class="mt-6 space-y-3 text-sm font-semibold text-neutral-800">
            <li>多线节点</li>
            <li>带宽可选</li>
            <li>基础防护说明</li>
          </ul>
        </div>
        <div class="rounded-[1.75rem] border border-neutral-200 bg-neutral-50 p-7">
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

    <section class="border-y border-neutral-200 bg-neutral-50">
      <div class="mx-auto grid max-w-7xl grid-cols-2 gap-px bg-neutral-200 px-4 sm:px-6 md:grid-cols-4 lg:px-8">
        <div class="bg-neutral-50 py-10 text-center">
          <div class="text-4xl font-black text-neutral-950">5.2G</div>
          <div class="mt-2 text-sm text-neutral-500">高主频展示</div>
        </div>
        <div class="bg-neutral-50 py-10 text-center">
          <div class="text-4xl font-black text-neutral-950">100M+</div>
          <div class="mt-2 text-sm text-neutral-500">带宽套餐</div>
        </div>
        <div class="bg-neutral-50 py-10 text-center">
          <div class="text-4xl font-black text-neutral-950">4+</div>
          <div class="mt-2 text-sm text-neutral-500">游戏场景</div>
        </div>
        <div class="bg-neutral-50 py-10 text-center">
          <div class="text-4xl font-black text-neutral-950">24/7</div>
          <div class="mt-2 text-sm text-neutral-500">服务支持</div>
        </div>
      </div>
    </section>

    <section class="mx-auto max-w-7xl px-4 py-20 sm:px-6 lg:px-8">
      <div class="rounded-[2rem] border border-neutral-950 bg-white p-8 shadow-[10px_10px_0_#111] md:p-12">
        <div class="grid gap-8 md:grid-cols-[1fr_auto] md:items-center">
          <div>
            <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Ready</p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-neutral-950 sm:text-4xl">先选配置，购买入口即将开放</h2>
            <p class="mt-4 max-w-2xl text-sm leading-6 text-neutral-600">当前阶段只做产品展示和用户账号入口，不展示订单、支付或实例交付承诺。</p>
          </div>
          <RouterLink to="/register" class="btn-dark inline-flex justify-center rounded-full border px-7 py-3 text-sm font-black">
            免费注册
          </RouterLink>
        </div>
      </div>
    </section>
  </div>
</template>
