import { useConfirm } from 'primevue/useconfirm'

interface ConfirmActionOptions {
  header: string
  message: string
  acceptLabel?: string
}

/**
 * 统一后台危险操作确认弹窗，避免分散使用浏览器原生确认框。
 */
export function useConfirmAction() {
  const confirm = useConfirm()

  return (options: ConfirmActionOptions) =>
    new Promise<boolean>((resolve) => {
      let settled = false

      const finish = (result: boolean) => {
        if (settled) {
          return
        }
        settled = true
        resolve(result)
      }

      confirm.require({
        header: options.header,
        message: options.message,
        icon: 'pi pi-exclamation-triangle',
        rejectProps: {
          label: '取消',
          severity: 'secondary',
          outlined: true,
        },
        acceptProps: {
          label: options.acceptLabel || '确认',
          severity: 'danger',
        },
        accept: () => finish(true),
        reject: () => finish(false),
        onHide: () => finish(false),
      })
    })
}
