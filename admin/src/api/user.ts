import { request } from '../lib/http'

export interface UserRow {
  id: number
  user_no: string
  email: string
  phone: string
  status: string
  created_at: string
}

export interface AdminRow {
  id: number
  admin_no: string
  username: string
  status: string
  created_at: string
}

export function listUsers(limit = 20): Promise<UserRow[]> {
  return request<UserRow[]>(`/users?limit=${limit}`)
}

export function listAdmins(limit = 20): Promise<AdminRow[]> {
  return request<AdminRow[]>(`/admins?limit=${limit}`)
}
