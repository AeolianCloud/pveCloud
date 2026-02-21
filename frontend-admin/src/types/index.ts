// 统一响应结构，与后端 response.go 对应
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 分页响应结构，与后端 pagination.Result 对应
export interface PageResult<T> {
  page_num: number
  page_size: number
  total: number
  list: T[]
}

// 权限
export interface AdminPermission {
  id: number
  name: string
  label: string
  group: string
}

// 管理员角色
export interface AdminRole {
  id: number
  name: string
  label: string
  description: string
  sort: number
  permissions?: AdminPermission[]
}

// 管理员信息
export interface AdminUser {
  id: number
  username: string
  nickname: string
  avatar: string
  email: string
  status: number
  last_login_at: string | null
  roles: AdminRole[]
  created_at: string
}

// 登录响应
export interface LoginResult {
  token: string
  user: AdminUser
}
