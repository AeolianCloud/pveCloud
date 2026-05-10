import { reactive } from 'vue'

export type ConfirmTone = 'default' | 'danger'

export interface ConfirmOptions {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  tone?: ConfirmTone
}

interface ConfirmState extends Required<ConfirmOptions> {
  visible: boolean
  resolve: ((value: boolean) => void) | null
}

const state = reactive<ConfirmState>({
  visible: false,
  title: '确认操作',
  message: '',
  confirmText: '确认',
  cancelText: '取消',
  tone: 'default',
  resolve: null,
})

function confirm(options: ConfirmOptions) {
  if (state.resolve) {
    state.resolve(false)
  }
  Object.assign(state, {
    visible: true,
    title: options.title || '确认操作',
    message: options.message,
    confirmText: options.confirmText || '确认',
    cancelText: options.cancelText || '取消',
    tone: options.tone || 'default',
  })
  return new Promise<boolean>((resolve) => {
    state.resolve = resolve
  })
}

function close(value: boolean) {
  state.visible = false
  state.resolve?.(value)
  state.resolve = null
}

export function useConfirm() {
  return { state, confirm, close }
}
