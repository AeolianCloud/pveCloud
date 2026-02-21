// composables/useTableScroll.ts
// 响应式表格横向滚动宽度 composable。
//
// 使用方式：
//   const { scrollX } = useTableScroll(900)
//
// 说明：
// - 侧边栏宽度约 260px，故可用宽度 = window.innerWidth - 260
// - 当可用宽度 < minWidth 时，启用横向滚动（scroll-x = minWidth）
// - 当可用宽度 >= minWidth 时，不设 scroll-x（表格自然铺满）
// - 组件卸载时自动移除 resize 监听，无需手动清理
import { ref, onMounted, onUnmounted } from 'vue'

const SIDER_WIDTH = 260

export function useTableScroll(minWidth: number) {
  const scrollX = ref<number | undefined>(
    window.innerWidth - SIDER_WIDTH < minWidth ? minWidth : undefined,
  )

  function onResize() {
    scrollX.value = window.innerWidth - SIDER_WIDTH < minWidth ? minWidth : undefined
  }

  onMounted(() => window.addEventListener('resize', onResize))
  onUnmounted(() => window.removeEventListener('resize', onResize))

  return { scrollX }
}
