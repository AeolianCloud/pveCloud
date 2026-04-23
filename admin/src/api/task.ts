import { request } from '../lib/http'

export interface Task {
  id: number
  task_no: string
  task_type: string
  business_type: string
  business_id: number
  status: string
  retry_count: number
  max_retry_count: number
  next_run_at: string
}

export function listTasks() {
  return request<Task[]>('/tasks')
}
