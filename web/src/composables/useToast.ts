import { reactive } from 'vue'

export type ToastTone = 'success' | 'error' | 'info'

export interface ToastItem {
  id: number
  tone: ToastTone
  message: string
}

const toasts = reactive<ToastItem[]>([])
let nextToastId = 1

function push(tone: ToastTone, message: string) {
  const id = nextToastId++
  toasts.push({ id, tone, message })
  window.setTimeout(() => remove(id), 3200)
}

function remove(id: number) {
  const index = toasts.findIndex((item) => item.id === id)
  if (index >= 0) {
    toasts.splice(index, 1)
  }
}

export function useToast() {
  return {
    toasts,
    remove,
    success: (message: string) => push('success', message),
    error: (message: string) => push('error', message),
    info: (message: string) => push('info', message),
  }
}
