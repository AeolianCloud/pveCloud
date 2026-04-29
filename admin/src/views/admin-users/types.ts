export type EditorMode = 'create' | 'edit'
export type AdminStatus = 'active' | 'disabled'
export type RoleStatus = 'active' | 'disabled'
export type SessionStatus = 'active' | 'revoked' | 'expired'

export interface UserQueryFormState {
  keyword: string
  status: '' | AdminStatus
  role_id: number | undefined
}

export interface RoleQueryFormState {
  keyword: string
  status: '' | RoleStatus
}

export interface AdminSessionQueryFormState {
  keyword: string
  status: '' | SessionStatus
}

export interface UserEditorState {
  username: string
  email: string
  display_name: string
  password: string
  status: AdminStatus
  role_ids: number[]
}

export interface UserEditorSnapshot {
  email: string
  display_name: string
  status: AdminStatus
  role_ids: number[]
}

export interface PasswordFormState {
  password: string
}

export interface RoleEditorState {
  code: string
  name: string
  description: string
  status: RoleStatus
  permission_codes: string[]
}

export interface RoleEditorSnapshot {
  name: string
  description: string
  status: RoleStatus
  permission_codes: string[]
}

export interface PermissionTreeNode {
  id: string
  label: string
  type: 'root' | 'segment' | 'permission'
  code?: string
  count?: number
  description?: string | null
  children?: PermissionTreeNode[]
  disabled?: boolean
  meta_label?: string
  path_hint?: string
  keywords?: string[]
  sort_order?: number
}

export interface PaginationState {
  page: number
  per_page: number
  total: number
  last_page: number
}
