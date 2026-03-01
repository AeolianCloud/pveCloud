<!-- 联系区块：询价表单 -->
<template>
  <section id="contact" class="bg-white py-14 border-b border-gray-200">
    <div class="max-w-7xl mx-auto px-6">
      <div class="max-w-xl mx-auto">
        <!-- 询价表单 -->
        <div class="border border-gray-200 rounded p-6 bg-gray-50">
          <h3 class="text-base font-semibold text-gray-900 mb-5">快速询价</h3>

          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-gray-600 mb-1.5">姓名</label>
                <input
                  v-model="form.name"
                  type="text"
                  placeholder="联系人姓名"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded text-sm text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 bg-white transition-colors duration-150"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-600 mb-1.5">手机号</label>
                <input
                  v-model="form.phone"
                  type="tel"
                  placeholder="186 0000 0000"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded text-sm text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 bg-white transition-colors duration-150"
                />
              </div>
            </div>

            <div>
              <label class="block text-xs font-medium text-gray-600 mb-1.5">公司名称（选填）</label>
              <input
                v-model="form.company"
                type="text"
                placeholder="所在公司"
                class="w-full px-3 py-2 border border-gray-300 rounded text-sm text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 bg-white transition-colors duration-150"
              />
            </div>

            <div>
              <label class="block text-xs font-medium text-gray-600 mb-1.5">需求描述</label>
              <textarea
                v-model="form.message"
                placeholder="请描述您的业务规模、所需配置、带宽要求等"
                rows="3"
                class="w-full px-3 py-2 border border-gray-300 rounded text-sm text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 bg-white transition-colors duration-150 resize-none"
              ></textarea>
            </div>

            <div v-if="submitted" class="px-3 py-2.5 bg-green-50 border border-green-200 rounded text-sm text-green-700">
              已收到您的信息，我们将在 1 小时内联系您。
            </div>

            <button
              type="submit"
              :disabled="submitting"
              class="w-full py-2.5 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white text-sm font-medium rounded transition-colors duration-150"
            >
              {{ submitting ? '提交中...' : '提交询价' }}
            </button>

            <p class="text-xs text-gray-400 text-center">
              提交即表示同意我们的隐私政策，信息将严格保密
            </p>
          </form>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

// 表单数据
const form = ref({ name: '', phone: '', company: '', message: '' })
const submitting = ref(false)
const submitted = ref(false)

// 提交询价（TODO: 对接后端接口）
async function handleSubmit() {
  submitting.value = true
  try {
    await new Promise((resolve) => setTimeout(resolve, 800))
    submitted.value = true
    form.value = { name: '', phone: '', company: '', message: '' }
  } finally {
    submitting.value = false
  }
}
</script>
