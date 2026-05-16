import { reportAdminClientError, type ClientErrorLogPayload } from '../api/log'

let reporting = false

export function installClientErrorReporter() {
  window.addEventListener('error', (event) => {
    void reportClientError({
      page_path: currentPath(),
      error_type: event.error?.name || 'Error',
      message: event.message || '脚本错误',
      stack: event.error?.stack,
    })
  })
  window.addEventListener('unhandledrejection', (event) => {
    const reason = event.reason
    void reportClientError({
      page_path: currentPath(),
      error_type: reason?.name || 'UnhandledRejection',
      message: reason?.message || String(reason || 'Promise 未处理异常'),
      stack: reason?.stack,
    })
  })
}

async function reportClientError(payload: ClientErrorLogPayload) {
  if (reporting) return
  reporting = true
  try {
    await reportAdminClientError({
      ...payload,
      message: sanitize(payload.message, 500) || '前端错误',
      stack: sanitize(payload.stack || '', 5000),
      browser: navigator.userAgent.slice(0, 255),
    })
  } catch {
    // 客户端错误上报失败不影响正常页面流程。
  } finally {
    reporting = false
  }
}

function currentPath() {
  return `${window.location.pathname}${window.location.search}`.slice(0, 255)
}

function sanitize(value: string, limit: number) {
  return value
    .replace(/Bearer\s+[A-Za-z0-9._~+/=-]+/gi, 'Bearer [masked]')
    .replace(/token=([^&\s]+)/gi, 'token=[masked]')
    .replace(/password=([^&\s]+)/gi, 'password=[masked]')
    .slice(0, limit)
}
