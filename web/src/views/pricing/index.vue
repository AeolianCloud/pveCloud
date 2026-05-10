<template>
  <div class="bg-white">
    <section class="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
      <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Pricing</p>
      <h1 class="mt-4 text-5xl font-black tracking-tight text-neutral-950 sm:text-6xl">价格方案</h1>
      <p class="mt-5 max-w-2xl text-base leading-7 text-neutral-600">浅色价格骨架，保留月付、季付、半年付、年付展示。最终价格以后端公开目录返回为准。</p>

      <div class="mt-12 overflow-hidden rounded-[1.5rem] border border-neutral-950 bg-white shadow-[10px_10px_0_#111]">
        <div class="grid grid-cols-5 border-b border-neutral-950 bg-neutral-950 text-sm font-black text-white">
          <div class="p-4">套餐</div>
          <div class="p-4 text-center">月付</div>
          <div class="p-4 text-center">季付</div>
          <div class="p-4 text-center">半年付</div>
          <div class="p-4 text-center">年付</div>
        </div>
        <div class="divide-y divide-neutral-200 text-sm">
          <div class="grid grid-cols-5 items-center">
            <div class="p-4"><div class="font-black">轻量开服</div><div class="mt-1 text-neutral-500">2C / 4G / 40G / 10M</div></div>
            <div class="p-4 text-center font-bold">¥49</div>
            <div class="p-4 text-center font-bold">¥132</div>
            <div class="p-4 text-center font-bold">¥250</div>
            <div class="p-4 text-center font-black">¥470</div>
          </div>
          <div class="grid grid-cols-5 items-center bg-neutral-50">
            <div class="p-4"><div class="font-black">高主频游戏</div><div class="mt-1 text-neutral-500">4C / 8G / 80G / 30M</div></div>
            <div class="p-4 text-center font-bold">¥199</div>
            <div class="p-4 text-center font-bold">¥537</div>
            <div class="p-4 text-center font-bold">¥1015</div>
            <div class="p-4 text-center font-black">¥1910</div>
          </div>
          <div class="grid grid-cols-5 items-center">
            <div class="p-4"><div class="font-black">高防大带宽</div><div class="mt-1 text-neutral-500">4C / 8G / 100G / 100M</div></div>
            <div class="p-4 text-center font-bold">¥399</div>
            <div class="p-4 text-center font-bold">¥1077</div>
            <div class="p-4 text-center font-bold">¥2035</div>
            <div class="p-4 text-center font-black">¥3830</div>
          </div>
          <div class="grid grid-cols-5 items-center bg-neutral-50">
            <div class="p-4"><div class="font-black">独享物理机</div><div class="mt-1 text-neutral-500">E5/i9 / 32G+ / 定制</div></div>
            <div class="p-4 text-center font-bold">¥899</div>
            <div class="p-4 text-center font-bold">¥2427</div>
            <div class="p-4 text-center font-bold">¥4585</div>
            <div class="p-4 text-center font-black">¥8630</div>
          </div>
        </div>
      </div>
    </section>

    <section class="mx-auto grid max-w-7xl gap-5 px-4 py-10 sm:px-6 lg:grid-cols-3 lg:px-8">
      <div class="rounded-[1.5rem] border border-neutral-200 p-6">
        <h2 class="text-lg font-black">计费说明</h2>
        <p class="mt-3 text-sm leading-6 text-neutral-600">展示月付、季付、半年付、年付价格口径，实际可售周期以后端公开目录为准。</p>
      </div>
      <div class="rounded-[1.5rem] border border-neutral-200 p-6">
        <h2 class="text-lg font-black">购买状态</h2>
        <p class="mt-3 text-sm leading-6 text-neutral-600">当前阶段不开放订单和支付流程，CTA 使用登录后查看购买入口。</p>
      </div>
      <div class="rounded-[1.5rem] border border-neutral-200 p-6">
        <h2 class="text-lg font-black">适合场景</h2>
        <p class="mt-3 text-sm leading-6 text-neutral-600">游戏联机、社区开服、轻量应用、临时活动节点和独立开发测试。</p>
      </div>
    </section>
  </div>
</template>
