import type { Directive } from 'vue'
import { useAuthStore } from '@/store/auth'

/**
 * v-permission 权限指令
 * 无权限时将元素从 DOM 中移除（而非仅隐藏，防止通过 CSS 绕过）
 *
 * 用法：
 *   v-permission="'admin:create'"          单个权限
 *   v-permission="['admin:create', 'admin:update']"  拥有其中任意一个即显示
 */
export const permission: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    const authStore = useAuthStore()
    const required = Array.isArray(binding.value) ? binding.value : [binding.value]
    const hasAny = required.some(name => authStore.hasPermission(name))
    if (!hasAny) {
      el.parentNode?.removeChild(el)
    }
  },
}
