import { http, type ApiEnvelope } from '../utils/request'
import type { PaginatedData } from './admin-user'

export type AsyncTaskStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'cancelled'

export interface AsyncTaskItem {
  task_no: string
  task_type: string
  status: AsyncTaskStatus
  object_type: string | null
  object_no: string | null
  scheduled_at: string
  attempts: number
  max_attempts: number
  last_error_code: string | null
  last_error_message: string | null
  locked_by: string | null
  locked_until: string | null
  created_at: string
  completed_at: string | null
}

export async function getAsyncTasks(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<AsyncTaskItem>>>('/async-tasks', { params })
  return response.data.data
}

export async function retryAsyncTask(taskNo: string, remark?: string) {
  const response = await http.post<ApiEnvelope<AsyncTaskItem>>(`/async-tasks/${taskNo}/retry`, { remark })
  return response.data.data
}
