import type { InvoiceStatus, InvoiceTitleType } from '../../api/invoice'

export const invoiceStatusText: Record<InvoiceStatus | string, string> = {
  pending: '待处理',
  processing: '处理中',
  issued: '已开票',
  rejected: '已驳回',
  cancelled: '已取消',
}

export const invoiceStatusClass: Record<InvoiceStatus | string, string> = {
  pending: 'border-amber-200 bg-amber-50 text-amber-700',
  processing: 'border-sky-200 bg-sky-50 text-sky-700',
  issued: 'border-emerald-200 bg-emerald-50 text-emerald-700',
  rejected: 'border-red-200 bg-red-50 text-red-700',
  cancelled: 'border-neutral-200 bg-neutral-100 text-neutral-700',
}

export const titleTypeText: Record<InvoiceTitleType | string, string> = {
  personal: '个人',
  company: '企业',
}

export const orderTypeText: Record<string, string> = {
  purchase: '新购',
  renewal: '续费',
}

export function formatMoney(cents: number, currency = 'CNY') {
  const prefix = currency === 'CNY' ? '¥' : `${currency} `
  return `${prefix}${(cents / 100).toFixed(2)}`
}

export function formatDateTime(value: string | null | undefined) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

export function saveInvoiceBlob(blob: Blob, invoiceNo: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${invoiceNo}.pdf`
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
