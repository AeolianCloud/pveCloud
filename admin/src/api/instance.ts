import { request } from '../lib/http'

export interface Instance {
  id: number
  instance_no: string
  user_id: number
  order_id: number
  node_id: number
  status: string
  instance_ref: string
  created_at: string
}

export function listInstances() {
  return request<Instance[]>('/instances')
}
