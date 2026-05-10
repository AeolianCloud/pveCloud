import type { DialogApi, MessageApi, NotificationApi } from 'naive-ui'

let messageApi: MessageApi | null = null
let dialogApi: DialogApi | null = null
let notificationApi: NotificationApi | null = null

export function registerNaiveApis(payload: {
  message: MessageApi
  dialog: DialogApi
  notification: NotificationApi
}) {
  messageApi = payload.message
  dialogApi = payload.dialog
  notificationApi = payload.notification
}

function ensure<T>(value: T | null, name: string): T {
  if (!value) throw new Error(`naive ${name} api 未初始化，请确保根组件已挂载 NaiveApiBridge`)
  return value
}

export const message = {
  success: (content: string) => ensure(messageApi, 'message').success(content),
  error: (content: string) => ensure(messageApi, 'message').error(content),
  warning: (content: string) => ensure(messageApi, 'message').warning(content),
  info: (content: string) => ensure(messageApi, 'message').info(content),
  loading: (content: string) => ensure(messageApi, 'message').loading(content),
}

export function getDialog(): DialogApi {
  return ensure(dialogApi, 'dialog')
}

export function getNotification(): NotificationApi {
  return ensure(notificationApi, 'notification')
}

export function confirm(options: {
  title?: string
  content: string
  positiveText?: string
  negativeText?: string
  type?: 'warning' | 'error' | 'info' | 'success'
}): Promise<void> {
  return new Promise((resolve, reject) => {
    const dialog = ensure(dialogApi, 'dialog')
    const create =
      options.type === 'error'
        ? dialog.error
        : options.type === 'info'
          ? dialog.info
          : options.type === 'success'
            ? dialog.success
            : dialog.warning
    create({
      title: options.title || '请确认',
      content: options.content,
      positiveText: options.positiveText || '确认',
      negativeText: options.negativeText || '取消',
      onPositiveClick: () => resolve(),
      onNegativeClick: () => reject(new Error('cancel')),
      onClose: () => reject(new Error('cancel')),
      onMaskClick: () => reject(new Error('cancel')),
    })
  })
}
