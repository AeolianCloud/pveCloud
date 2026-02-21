// utils/format.ts
// 通用格式化工具函数。

/**
 * formatTime 将 ISO 时间字符串格式化为中文本地时间。
 * 输出示例：2026/02/21 14:30:00
 */
export function formatTime(t: string): string {
  return new Date(t).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}
