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
  email: string | null
  status: number
  last_login_at: string | null
  roles: AdminRole[]
  created_at: string
}

// 登录响应
export interface LoginResult {
  token: string
  refresh_token: string
  user: AdminUser
}

// 刷新 Token 响应
export interface RefreshResult {
  token: string
  refresh_token: string
}

// 菜单树节点（后端动态下发）。
//
// 说明：
// - children 用于树形展示（侧边栏/菜单管理页）。
// - path 为空（null）时表示目录节点（只用于分组/展开，不直接跳转）。
export interface AdminMenuNode {
  id: number
  parent_id: number
  title: string
  path: string | null
  permission: string | null
  super_admin_only: number
  icon: string | null
  sort: number
  visible: number
  children?: AdminMenuNode[]
}
