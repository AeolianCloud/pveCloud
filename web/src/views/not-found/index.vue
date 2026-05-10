<template>
  <div class="flex min-h-screen flex-col justify-center bg-white px-4 py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-lg">
      <div class="text-center">
        <h1 class="text-9xl font-black text-neutral-950">404</h1>
        <h2 class="mt-4 text-3xl font-black text-neutral-950 sm:text-4xl">
          页面未找到
        </h2>
        <p class="mt-4 text-lg text-neutral-500">
          抱歉，您访问的页面不存在或已被移除。
        </p>
        <div class="mt-8">
          <RouterLink
            to="/"
            class="btn-dark inline-flex items-center rounded-full border px-7 py-3 text-sm font-black"
          >
            返回首页
          </RouterLink>
        </div>
        <div class="mt-4">
          <RouterLink
            to="/products"
            class="font-black text-neutral-950 underline decoration-2 underline-offset-4"
          >
            查看产品
          </RouterLink>
          <span class="mx-2 text-neutral-300">|</span>
          <RouterLink
            to="/pricing"
            class="font-black text-neutral-950 underline decoration-2 underline-offset-4"
          >
            查看价格
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>
