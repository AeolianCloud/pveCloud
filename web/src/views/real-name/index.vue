<script setup lang="ts">
import { ref } from 'vue'

const realName = ref('')
const idCard = ref('')
const provider = ref('manual')
const loading = ref(false)
const error = ref('')
const success = ref('')
const status = ref('unverified')

const handleSubmit = async () => {
  if (!realName.value || !idCard.value) {
    error.value = '请填写所有字段'
    return
  }

  loading.value = true
  error.value = ''
  success.value = ''

  try {
    success.value = '实名认证申请已提交'
    status.value = 'pending'
  } catch (err) {
    error.value = '提交失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="bg-white">
    <div class="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
      <p class="text-sm font-black uppercase tracking-[0.18em] text-neutral-500">Real Name</p>
      <h1 class="mt-3 text-4xl font-black tracking-tight text-neutral-950">实名认证</h1>

      <div v-if="error" class="mt-6 rounded-xl border border-neutral-950 bg-neutral-50 p-4 text-sm font-bold text-neutral-950">{{ error }}</div>
      <div v-if="success" class="mt-6 rounded-xl border border-neutral-300 bg-white p-4 text-sm font-bold text-neutral-950">{{ success }}</div>

      <section class="mt-8 rounded-[1.5rem] border border-neutral-950 bg-white p-6 shadow-[8px_8px_0_#111]">
        <h2 class="text-xl font-black text-neutral-950">实名状态</h2>
        <div class="mt-5 rounded-2xl border border-neutral-200 bg-neutral-50 p-5">
          <div class="flex items-center gap-4">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl border border-neutral-950 bg-white text-sm font-black">ID</div>
            <div>
              <div class="text-sm font-black text-neutral-950">
                <span v-if="status === 'unverified'">未实名</span>
                <span v-else-if="status === 'pending'">核验中</span>
                <span v-else-if="status === 'approved'">已通过</span>
                <span v-else-if="status === 'rejected'">已拒绝</span>
              </div>
              <div class="mt-1 text-sm text-neutral-500">当前状态以后端返回为准，这里仅为静态展示骨架。</div>
            </div>
          </div>
        </div>

        <form v-if="status === 'unverified' || status === 'rejected'" class="mt-8 space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label for="real-name" class="mb-2 block text-sm font-black text-neutral-800">真实姓名</label>
            <input id="real-name" v-model="realName" type="text" class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入真实姓名" />
          </div>
          <div>
            <label for="id-card" class="mb-2 block text-sm font-black text-neutral-800">身份证号码</label>
            <input id="id-card" v-model="idCard" type="text" class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm outline-none focus:border-neutral-950" placeholder="请输入身份证号码" />
          </div>
          <div>
            <label class="mb-2 block text-sm font-black text-neutral-800">实名方式</label>
            <label class="flex items-center gap-3 rounded-xl border border-neutral-200 px-4 py-3 text-sm font-bold text-neutral-700">
              <input v-model="provider" name="provider" type="radio" value="manual" class="h-4 w-4 border-neutral-300 text-neutral-950" />
              人工审核
            </label>
          </div>
          <button type="submit" :disabled="loading" class="btn-dark w-full rounded-full border py-3 text-sm font-black disabled:opacity-50">{{ loading ? '提交中...' : '提交实名认证' }}</button>
        </form>
      </section>
    </div>
  </div>
</template>
