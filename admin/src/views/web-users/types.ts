export type WebUsersTabKey = 'users' | 'sessions'
export type EditorMode = 'create' | 'edit'

export interface PaginationState {
  page: number
  per_page: number
  total: number
  last_page: number
}

export interface UserQueryState {
  keyword: string
  status: string
}

export interface SessionQueryState {
  user_id: number | undefined
  status: string
}

export interface UserFormState {
  username: string
  email: string
  display_name: string
  password: string
  status: string
}

export interface PasswordFormState {
  password: string
}
