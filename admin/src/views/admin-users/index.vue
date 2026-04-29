<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormRules } from 'element-plus'

import QueryState from '../../components/QueryState.vue'
import {
  getAdminSessions,
  revokeAdminSession,
  type AdminSessionItem,
} from '../../api/admin-session'
import {
  createAdminRole,
  getAdminPermissions,
  getAdminRoles,
  updateAdminRole,
  type AdminPermissionGroup,
  type AdminRoleCreateRequest,
  type AdminRoleItem,
  type AdminRoleUpdateRequest,
} from '../../api/admin-role'
import {
  createAdminUser,
  getAdminUsers,
  resetAdminUserPassword,
  updateAdminUser,
  type AdminUserCreateRequest,
  type AdminUserItem,
  type AdminUserUpdateRequest,
} from '../../api/admin-user'
import { useAuthStore } from '../../store/modules/auth'
import { usePermissionStore } from '../../store/modules/permission'
import AdminRolesTab from './components/AdminRolesTab.vue'
import AdminSessionsTab from './components/AdminSessionsTab.vue'
import AdminUsersTab from './components/AdminUsersTab.vue'
import PasswordResetDialog from './components/PasswordResetDialog.vue'
import RoleEditorDialog from './components/RoleEditorDialog.vue'
import UserEditorDialog from './components/UserEditorDialog.vue'
import type {
  AdminSessionQueryFormState,
  AdminStatus,
  EditorMode,
  PasswordFormState,
  PaginationState,
  RoleEditorSnapshot,
  RoleEditorState,
  RoleQueryFormState,
  RoleStatus,
  UserEditorSnapshot,
  UserEditorState,
  UserQueryFormState,
} from './types'

type AdminSettingsTabKey = 'users' | 'roles' | 'sessions'

const activeTab = ref<AdminSettingsTabKey>('users')
const authStore = useAuthStore()
const permissionStore = usePermissionStore()

const initialLoading = ref(false)
const errorMessage = ref('')

const roleOptionsLoading = ref(false)

const userRefreshing = ref(false)
const userTableLoading = ref(false)
const userSubmitting = ref(false)
const passwordSubmitting = ref(false)
const userStatusUpdatingId = ref<number | null>(null)

const roleRefreshing = ref(false)
const roleTableLoading = ref(false)
const roleSubmitting = ref(false)
const roleStatusUpdatingId = ref<number | null>(null)

const sessionRefreshing = ref(false)
const sessionTableLoading = ref(false)
const sessionRevokingId = ref<string | null>(null)

const users = ref<AdminUserItem[]>([])
const roleOptions = ref<AdminRoleItem[]>([])
const roles = ref<AdminRoleItem[]>([])
const sessions = ref<AdminSessionItem[]>([])
const permissionGroups = ref<AdminPermissionGroup[]>([])

const userPagination = reactive<PaginationState>({
  page: 1,
  per_page: 15,
  total: 0,
  last_page: 0,
})

const rolePagination = reactive<PaginationState>({
  page: 1,
  per_page: 15,
  total: 0,
  last_page: 0,
})

const sessionPagination = reactive<PaginationState>({
  page: 1,
  per_page: 15,
  total: 0,
  last_page: 0,
})

const userQueryForm = reactive<UserQueryFormState>({
  keyword: '',
  status: '',
  role_id: undefined,
})

const roleQueryForm = reactive<RoleQueryFormState>({
  keyword: '',
  status: '',
})

const sessionQueryForm = reactive<AdminSessionQueryFormState>({
  keyword: '',
  status: '',
})

const userEditorVisible = ref(false)
const userEditorMode = ref<EditorMode>('create')
const editingUser = ref<AdminUserItem | null>(null)
const userEditorForm = reactive<UserEditorState>(createDefaultUserEditorForm())
const userEditorSnapshot = ref<UserEditorSnapshot | null>(null)

const passwordVisible = ref(false)
const passwordTarget = ref<AdminUserItem | null>(null)
const passwordForm = reactive<PasswordFormState>({
  password: '',
})

const roleEditorVisible = ref(false)
const roleEditorMode = ref<EditorMode>('create')
const editingRole = ref<AdminRoleItem | null>(null)
const roleEditorForm = reactive<RoleEditorState>(createDefaultRoleEditorForm())
const roleEditorSnapshot = ref<RoleEditorSnapshot | null>(null)

const hasUsers = computed(() => users.value.length > 0)
const hasRoles = computed(() => roles.value.length > 0)
const hasSessions = computed(() => sessions.value.length > 0)
const activeRoleOptions = computed(() => roleOptions.value.filter((role) => role.status === 'active'))
const currentSessionId = computed(() => authStore.session?.session_id || '')

const canViewUsersTab = computed(() => permissionStore.hasPermission('page.system-settings.admin-users'))
const canViewRolesTab = computed(() => permissionStore.hasPermission('page.system-settings.admin-roles'))
const canViewSessionsTab = computed(() => permissionStore.hasPermission('page.system-settings.admin-sessions'))

const canViewUsersResource = computed(() => permissionStore.hasPermission('admin-user:view'))
const canViewRolesResource = computed(() => permissionStore.hasPermission('admin-role:view'))
const canViewSessionsResource = computed(() => permissionStore.hasPermission('admin-session:view'))
const canReadRoleOptions = computed(() => permissionStore.hasPermission('admin-role:view'))

const canCreateUser = computed(() => permissionStore.hasPermission('admin-user:create'))
const canUpdateUser = computed(() => permissionStore.hasPermission('admin-user:update'))
const canResetUserPassword = computed(() => permissionStore.hasPermission('admin-user:password-reset'))
const canCreateRole = computed(() => permissionStore.hasPermission('admin-role:create'))
const canUpdateRole = computed(() => permissionStore.hasPermission('admin-role:update'))
const canRevokeSession = computed(() => permissionStore.hasPermission('admin-session:revoke'))

const isUserCreateMode = computed(() => userEditorMode.value === 'create')
const userEditorTitle = computed(() => (isUserCreateMode.value ? '新建管理员' : '编辑管理员'))

const isRoleCreateMode = computed(() => roleEditorMode.value === 'create')
const roleEditorTitle = computed(() => (isRoleCreateMode.value ? '新建管理组' : '编辑管理组'))
const isBuiltInRole = computed(() => editingRole.value?.code === 'super_admin')
const passwordTargetLabel = computed(() => passwordTarget.value?.display_name || passwordTarget.value?.username || '')

const userEditorRules: FormRules<UserEditorState> = {
  username: [
    { required: true, message: '请输入登录账号', trigger: 'blur' },
    { min: 3, max: 64, message: '账号长度需为 3 到 64 个字符', trigger: 'blur' },
  ],
  email: [{ type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }],
  display_name: [
    { required: true, message: '请输入显示名称', trigger: 'blur' },
    { min: 1, max: 64, message: '显示名称长度需为 1 到 64 个字符', trigger: 'blur' },
  ],
  password: [
    {
      validator: (_rule, value, callback) => {
        if (!isUserCreateMode.value && !value) {
          callback()
          return
        }
        if (!value) {
          callback(new Error('请输入登录密码'))
          return
        }
        if (value.length < 6 || value.length > 72) {
          callback(new Error('密码长度需为 6 到 72 个字符'))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  status: [{ required: true, message: '请选择账号状态', trigger: 'change' }],
}

const passwordRules: FormRules<PasswordFormState> = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 72, message: '密码长度需为 6 到 72 个字符', trigger: 'blur' },
  ],
}

const roleEditorRules: FormRules<RoleEditorState> = {
  code: [
    { required: true, message: '请输入管理组编码', trigger: 'blur' },
    { min: 2, max: 64, message: '编码长度需为 2 到 64 个字符', trigger: 'blur' },
  ],
  name: [
    { required: true, message: '请输入管理组名称', trigger: 'blur' },
    { min: 1, max: 64, message: '名称长度需为 1 到 64 个字符', trigger: 'blur' },
  ],
  description: [{ max: 255, message: '说明长度不能超过 255 个字符', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

watch([canViewUsersTab, canViewRolesTab, canViewSessionsTab], () => {
  syncVisibleTab()
}, { immediate: true })

void initializePage()

async function initializePage() {
  initialLoading.value = true
  errorMessage.value = ''
  try {
    const tasks: Promise<unknown>[] = []
    if (canViewUsersTab.value && canViewUsersResource.value) {
      tasks.push(loadUsersData())
    }
    if (canReadRoleOptions.value) {
      tasks.push(loadRoleOptions())
    }
    if (canViewRolesTab.value && canViewRolesResource.value) {
      tasks.push(loadPermissionGroups(), loadRolesData())
    }
    if (canViewSessionsTab.value && canViewSessionsResource.value) {
      tasks.push(loadSessionsData())
    }
    await Promise.all(tasks)
    syncVisibleTab()
  } catch (error) {
    errorMessage.value = toErrorMessage(error, '管理员设置加载失败')
  } finally {
    initialLoading.value = false
  }
}

async function loadRoleOptions() {
  roleOptionsLoading.value = true
  try {
    const result = await getAdminRoles({ page: 1, per_page: 100 })
    roleOptions.value = result.list
  } finally {
    roleOptionsLoading.value = false
  }
}

async function loadPermissionGroups() {
  permissionGroups.value = await getAdminPermissions()
}

async function loadUsersData() {
  const result = await getAdminUsers({
    page: userPagination.page,
    per_page: userPagination.per_page,
    keyword: normalizeKeyword(userQueryForm.keyword),
    status: userQueryForm.status || undefined,
    role_id: userQueryForm.role_id,
  })
  users.value = result.list
  userPagination.total = result.total
  userPagination.page = result.page
  userPagination.per_page = result.per_page
  userPagination.last_page = result.last_page
}

async function loadRolesData() {
  const result = await getAdminRoles({
    page: rolePagination.page,
    per_page: rolePagination.per_page,
    keyword: normalizeKeyword(roleQueryForm.keyword),
    status: roleQueryForm.status || undefined,
  })
  roles.value = result.list
  rolePagination.total = result.total
  rolePagination.page = result.page
  rolePagination.per_page = result.per_page
  rolePagination.last_page = result.last_page
}

async function loadSessionsData() {
  const result = await getAdminSessions({
    page: sessionPagination.page,
    per_page: sessionPagination.per_page,
    keyword: normalizeKeyword(sessionQueryForm.keyword),
    status: sessionQueryForm.status || undefined,
  })
  sessions.value = result.list
  sessionPagination.total = result.total
  sessionPagination.page = result.page
  sessionPagination.per_page = result.per_page
  sessionPagination.last_page = result.last_page
}

async function reloadRoleDataForAllViews() {
  const tasks: Promise<unknown>[] = [loadRolesData()]
  if (canReadRoleOptions.value) {
    tasks.push(loadRoleOptions())
  }
  await Promise.all(tasks)
}

async function handleUserRefresh() {
  userRefreshing.value = true
  try {
    const tasks: Promise<unknown>[] = []
    if (canViewUsersResource.value) {
      tasks.push(loadUsersData())
    }
    if (canReadRoleOptions.value) {
      tasks.push(loadRoleOptions())
    }
    await Promise.all(tasks)
    ElMessage.success('管理员数据已刷新')
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '刷新失败'))
  } finally {
    userRefreshing.value = false
  }
}

async function handleRoleRefresh() {
  roleRefreshing.value = true
  try {
    await Promise.all([reloadRoleDataForAllViews(), loadPermissionGroups()])
    ElMessage.success('管理组数据已刷新')
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '刷新失败'))
  } finally {
    roleRefreshing.value = false
  }
}

async function handleSessionRefresh() {
  sessionRefreshing.value = true
  try {
    if (canViewSessionsResource.value) {
      await loadSessionsData()
    }
    ElMessage.success('管理员会话已刷新')
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '刷新失败'))
  } finally {
    sessionRefreshing.value = false
  }
}

async function searchUsers() {
  userTableLoading.value = true
  try {
    await loadUsersData()
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '管理员列表加载失败'))
  } finally {
    userTableLoading.value = false
  }
}

async function searchRoles() {
  roleTableLoading.value = true
  try {
    await loadRolesData()
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '管理组列表加载失败'))
  } finally {
    roleTableLoading.value = false
  }
}

async function searchSessions() {
  sessionTableLoading.value = true
  try {
    await loadSessionsData()
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '管理员会话加载失败'))
  } finally {
    sessionTableLoading.value = false
  }
}

function handleUserSearch() {
  userPagination.page = 1
  void searchUsers()
}

function handleUserResetFilters() {
  userQueryForm.keyword = ''
  userQueryForm.status = ''
  userQueryForm.role_id = undefined
  userPagination.page = 1
  void searchUsers()
}

function handleUserPageChange(page: number) {
  userPagination.page = page
  void searchUsers()
}

function handleUserPageSizeChange(size: number) {
  userPagination.per_page = size
  userPagination.page = 1
  void searchUsers()
}

function handleRoleSearch() {
  rolePagination.page = 1
  void searchRoles()
}

function handleRoleResetFilters() {
  roleQueryForm.keyword = ''
  roleQueryForm.status = ''
  rolePagination.page = 1
  void searchRoles()
}

function handleRolePageChange(page: number) {
  rolePagination.page = page
  void searchRoles()
}

function handleRolePageSizeChange(size: number) {
  rolePagination.per_page = size
  rolePagination.page = 1
  void searchRoles()
}

function handleSessionSearch() {
  sessionPagination.page = 1
  void searchSessions()
}

function handleSessionResetFilters() {
  sessionQueryForm.keyword = ''
  sessionQueryForm.status = ''
  sessionPagination.page = 1
  void searchSessions()
}

function handleSessionPageChange(page: number) {
  sessionPagination.page = page
  void searchSessions()
}

function handleSessionPageSizeChange(size: number) {
  sessionPagination.per_page = size
  sessionPagination.page = 1
  void searchSessions()
}

function openCreateUserDialog() {
  userEditorMode.value = 'create'
  editingUser.value = null
  resetUserEditorForm()
  userEditorVisible.value = true
}

function openEditUserDialog(user: AdminUserItem) {
  userEditorMode.value = 'edit'
  editingUser.value = user
  userEditorForm.username = user.username
  userEditorForm.email = user.email ?? ''
  userEditorForm.display_name = user.display_name
  userEditorForm.password = ''
  userEditorForm.status = normalizeUserStatus(user.status)
  userEditorForm.role_ids = [...user.role_ids]
  userEditorSnapshot.value = {
    email: user.email ?? '',
    display_name: user.display_name,
    status: normalizeUserStatus(user.status),
    role_ids: [...user.role_ids],
  }
  userEditorVisible.value = true
}

function handleUserEditorClosed() {
  resetUserEditorForm()
}

function resolveFirstVisibleTab(): AdminSettingsTabKey {
  if (canViewUsersTab.value) {
    return 'users'
  }
  if (canViewRolesTab.value) {
    return 'roles'
  }
  if (canViewSessionsTab.value) {
    return 'sessions'
  }
  return 'users'
}

function isTabVisible(tab: AdminSettingsTabKey) {
  if (tab === 'users') {
    return canViewUsersTab.value
  }
  if (tab === 'roles') {
    return canViewRolesTab.value
  }
  return canViewSessionsTab.value
}

function syncVisibleTab() {
  if (!isTabVisible(activeTab.value)) {
    activeTab.value = resolveFirstVisibleTab()
  }
}

async function submitUserEditor() {
  userSubmitting.value = true
  try {
    if (isUserCreateMode.value) {
      const payload: AdminUserCreateRequest = {
        username: userEditorForm.username.trim(),
        email: normalizeOptionalString(userEditorForm.email),
        display_name: userEditorForm.display_name.trim(),
        password: userEditorForm.password,
        status: userEditorForm.status,
        role_ids: uniqueSortedNumbers(userEditorForm.role_ids),
      }
      if (findInactiveRoleIds(payload.role_ids ?? []).length > 0) {
        ElMessage.warning('已停用角色不能分配给管理员')
        return
      }
      await createAdminUser(payload)
      ElMessage.success('管理员创建成功')
      userEditorVisible.value = false
      userPagination.page = 1
      await searchUsers()
      return
    }

    const user = editingUser.value
    if (!user) {
      return
    }
    const payload = buildUserUpdatePayload()
    if (Object.keys(payload).length === 0) {
      ElMessage.info('未检测到变更')
      userEditorVisible.value = false
      return
    }
    if (payload.role_ids && findInactiveRoleIds(payload.role_ids).length > 0) {
      ElMessage.warning('已停用角色不能继续分配，请先移除')
      return
    }

    await updateAdminUser(user.id, payload)
    ElMessage.success('管理员信息已更新')
    userEditorVisible.value = false
    await searchUsers()
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '保存失败'))
  } finally {
    userSubmitting.value = false
  }
}

function openPasswordDialog(user: AdminUserItem) {
  passwordTarget.value = user
  passwordVisible.value = true
  passwordForm.password = ''
}

function handlePasswordClosed() {
  passwordTarget.value = null
  passwordForm.password = ''
}

async function submitPasswordReset() {
  if (!passwordTarget.value) {
    return
  }

  passwordSubmitting.value = true
  try {
    await resetAdminUserPassword(passwordTarget.value.id, passwordForm.password)
    ElMessage.success('密码已重置')
    passwordVisible.value = false
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '密码重置失败'))
  } finally {
    passwordSubmitting.value = false
  }
}

async function toggleUserStatus(user: AdminUserItem) {
  const nextStatus: AdminStatus = user.status === 'active' ? 'disabled' : 'active'
  const nextLabel = formatStatusLabel(nextStatus)

  try {
    await ElMessageBox.confirm(`确认将管理员“${user.display_name}”设为${nextLabel}吗？`, '确认状态切换', {
      type: nextStatus === 'disabled' ? 'warning' : 'info',
      confirmButtonText: '确认',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  userStatusUpdatingId.value = user.id
  try {
    await updateAdminUser(user.id, { status: nextStatus })
    ElMessage.success(`管理员状态已更新为${nextLabel}`)
    await searchUsers()
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '状态更新失败'))
  } finally {
    userStatusUpdatingId.value = null
  }
}

function openCreateRoleDialog() {
  roleEditorMode.value = 'create'
  editingRole.value = null
  resetRoleEditorForm()
  roleEditorVisible.value = true
}

function openEditRoleDialog(role: AdminRoleItem) {
  roleEditorMode.value = 'edit'
  editingRole.value = role
  roleEditorForm.code = role.code
  roleEditorForm.name = role.name
  roleEditorForm.description = role.description ?? ''
  roleEditorForm.status = normalizeRoleStatus(role.status)
  roleEditorForm.permission_codes = [...role.permission_codes]
  roleEditorSnapshot.value = {
    name: role.name,
    description: role.description ?? '',
    status: normalizeRoleStatus(role.status),
    permission_codes: uniqueSortedStrings(role.permission_codes),
  }
  roleEditorVisible.value = true
}

function handleRoleEditorClosed() {
  resetRoleEditorForm()
}

async function submitRoleEditor() {
  roleSubmitting.value = true
  try {
    if (isRoleCreateMode.value) {
      const payload: AdminRoleCreateRequest = {
        code: roleEditorForm.code.trim(),
        name: roleEditorForm.name.trim(),
        description: normalizeOptionalString(roleEditorForm.description),
        status: roleEditorForm.status,
        permission_codes: uniqueSortedStrings(roleEditorForm.permission_codes),
      }
      await createAdminRole(payload)
      ElMessage.success('管理组创建成功')
      roleEditorVisible.value = false
      rolePagination.page = 1
      await reloadRoleDataForAllViews()
      return
    }

    const role = editingRole.value
    if (!role) {
      return
    }
    const payload = buildRoleUpdatePayload()
    if (Object.keys(payload).length === 0) {
      ElMessage.info('未检测到变更')
      roleEditorVisible.value = false
      return
    }

    await updateAdminRole(role.id, payload)
    ElMessage.success('管理组已更新')
    roleEditorVisible.value = false
    await reloadRoleDataForAllViews()
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '保存失败'))
  } finally {
    roleSubmitting.value = false
  }
}

async function toggleRoleStatus(role: AdminRoleItem) {
  const nextStatus: RoleStatus = role.status === 'active' ? 'disabled' : 'active'
  const nextLabel = formatStatusLabel(nextStatus)

  try {
    await ElMessageBox.confirm(`确认将管理组“${role.name}”设为${nextLabel}吗？`, '确认状态切换', {
      type: nextStatus === 'disabled' ? 'warning' : 'info',
      confirmButtonText: '确认',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  roleStatusUpdatingId.value = role.id
  try {
    await updateAdminRole(role.id, { status: nextStatus })
    ElMessage.success(`管理组状态已更新为${nextLabel}`)
    await reloadRoleDataForAllViews()
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '状态更新失败'))
  } finally {
    roleStatusUpdatingId.value = null
  }
}

async function revokeSession(session: AdminSessionItem) {
  if (session.is_current || session.session_id === currentSessionId.value) {
    ElMessage.warning('不能吊销当前会话')
    return
  }

  const targetLabel = session.admin_display_name || session.admin_username
  try {
    await ElMessageBox.confirm(`确认吊销管理员“${targetLabel}”的会话吗？`, '确认吊销会话', {
      type: 'warning',
      confirmButtonText: '确认吊销',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  sessionRevokingId.value = session.session_id
  try {
    await revokeAdminSession(session.session_id)
    ElMessage.success('管理员会话已吊销')
    await searchSessions()
  } catch (error) {
    ElMessage.error(toErrorMessage(error, '吊销会话失败'))
  } finally {
    sessionRevokingId.value = null
  }
}

function resetUserEditorForm() {
  Object.assign(userEditorForm, createDefaultUserEditorForm())
  userEditorSnapshot.value = null
}

function resetRoleEditorForm() {
  Object.assign(roleEditorForm, createDefaultRoleEditorForm())
  roleEditorSnapshot.value = null
}

function buildUserUpdatePayload(): AdminUserUpdateRequest {
  const snapshot = userEditorSnapshot.value
  if (!snapshot) {
    return {}
  }

  const payload: AdminUserUpdateRequest = {}
  const nextEmail = normalizeOptionalString(userEditorForm.email)
  const previousEmail = normalizeOptionalString(snapshot.email)
  if (nextEmail !== previousEmail) {
    payload.email = nextEmail
  }

  const nextDisplayName = userEditorForm.display_name.trim()
  if (nextDisplayName !== snapshot.display_name) {
    payload.display_name = nextDisplayName
  }

  if (userEditorForm.status !== snapshot.status) {
    payload.status = userEditorForm.status
  }

  const nextRoleIds = uniqueSortedNumbers(userEditorForm.role_ids)
  const previousRoleIds = uniqueSortedNumbers(snapshot.role_ids)
  if (!sameNumberList(nextRoleIds, previousRoleIds)) {
    payload.role_ids = nextRoleIds
  }

  return payload
}

function buildRoleUpdatePayload(): AdminRoleUpdateRequest {
  const snapshot = roleEditorSnapshot.value
  if (!snapshot) {
    return {}
  }

  const payload: AdminRoleUpdateRequest = {}
  const nextName = roleEditorForm.name.trim()
  if (nextName !== snapshot.name) {
    payload.name = nextName
  }

  const nextDescription = normalizeOptionalString(roleEditorForm.description)
  const previousDescription = normalizeOptionalString(snapshot.description)
  if (nextDescription !== previousDescription) {
    payload.description = nextDescription
  }

  if (roleEditorForm.status !== snapshot.status) {
    payload.status = roleEditorForm.status
  }

  const nextPermissionCodes = uniqueSortedStrings(roleEditorForm.permission_codes)
  if (!sameStringList(nextPermissionCodes, snapshot.permission_codes)) {
    payload.permission_codes = nextPermissionCodes
  }

  return payload
}

function createDefaultUserEditorForm(): UserEditorState {
  return {
    username: '',
    email: '',
    display_name: '',
    password: '',
    status: 'active',
    role_ids: [],
  }
}

function createDefaultRoleEditorForm(): RoleEditorState {
  return {
    code: '',
    name: '',
    description: '',
    status: 'active',
    permission_codes: [],
  }
}

function normalizeKeyword(value: string) {
  const trimmed = value.trim()
  return trimmed || undefined
}

function normalizeOptionalString(value: string | null | undefined) {
  const trimmed = value?.trim() ?? ''
  return trimmed ? trimmed : null
}

function uniqueSortedNumbers(values: number[]) {
  return Array.from(new Set(values.filter((value) => value > 0))).sort((left, right) => left - right)
}

function uniqueSortedStrings(values: string[]) {
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean))).sort()
}

function sameNumberList(left: number[], right: number[]) {
  if (left.length !== right.length) {
    return false
  }
  return left.every((value, index) => value === right[index])
}

function sameStringList(left: string[], right: string[]) {
  if (left.length !== right.length) {
    return false
  }
  return left.every((value, index) => value === right[index])
}

function findInactiveRoleIds(roleIds: number[]) {
  const activeIds = new Set(activeRoleOptions.value.map((role) => role.id))
  return uniqueSortedNumbers(roleIds).filter((roleId) => !activeIds.has(roleId))
}

function normalizeUserStatus(value: string): AdminStatus {
  return value === 'disabled' ? 'disabled' : 'active'
}

function normalizeRoleStatus(value: string): RoleStatus {
  return value === 'disabled' ? 'disabled' : 'active'
}

function formatStatusLabel(status: string) {
  return status === 'active' ? '启用' : '停用'
}

function toErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error && error.message.trim() ? error.message : fallback
}
</script>

<template>
  <div class="admin-settings-page">
    <div class="admin-settings-page__header">
      <div>
        <h2>管理员设置</h2>
        <p>在同一页面管理管理员账号、管理组权限和管理员会话。</p>
      </div>
    </div>

    <QueryState :loading="initialLoading" :error-message="errorMessage" @retry="initializePage">
      <el-tabs v-model="activeTab">
        <el-tab-pane v-if="canViewUsersTab" label="管理员账号" name="users">
          <AdminUsersTab
            :loading="userTableLoading || roleOptionsLoading"
            :refreshing="userRefreshing"
            :has-users="hasUsers"
            :can-view-users-resource="canViewUsersResource"
            :can-view-roles-tab="canViewRolesTab"
            :can-create-user="canCreateUser"
            :can-update-user="canUpdateUser"
            :can-reset-user-password="canResetUserPassword"
            :query-form="userQueryForm"
            :role-options="roleOptions"
            :users="users"
            :pagination="userPagination"
            :user-status-updating-id="userStatusUpdatingId"
            @search="handleUserSearch"
            @reset="handleUserResetFilters"
            @refresh="handleUserRefresh"
            @create="openCreateUserDialog"
            @edit="openEditUserDialog"
            @toggle-status="toggleUserStatus"
            @reset-password="openPasswordDialog"
            @page-change="handleUserPageChange"
            @page-size-change="handleUserPageSizeChange"
          />
        </el-tab-pane>

        <el-tab-pane v-if="canViewRolesTab" label="管理组权限" name="roles">
          <AdminRolesTab
            :loading="roleTableLoading"
            :refreshing="roleRefreshing"
            :has-roles="hasRoles"
            :can-view-roles-resource="canViewRolesResource"
            :can-create-role="canCreateRole"
            :can-update-role="canUpdateRole"
            :query-form="roleQueryForm"
            :roles="roles"
            :pagination="rolePagination"
            :role-status-updating-id="roleStatusUpdatingId"
            @search="handleRoleSearch"
            @reset="handleRoleResetFilters"
            @refresh="handleRoleRefresh"
            @create="openCreateRoleDialog"
            @edit="openEditRoleDialog"
            @toggle-status="toggleRoleStatus"
            @page-change="handleRolePageChange"
            @page-size-change="handleRolePageSizeChange"
          />
        </el-tab-pane>

        <el-tab-pane v-if="canViewSessionsTab" label="管理员会话" name="sessions">
          <AdminSessionsTab
            :loading="sessionTableLoading"
            :refreshing="sessionRefreshing"
            :has-sessions="hasSessions"
            :can-view-sessions-resource="canViewSessionsResource"
            :can-revoke-session="canRevokeSession"
            :query-form="sessionQueryForm"
            :sessions="sessions"
            :pagination="sessionPagination"
            :session-revoking-id="sessionRevokingId"
            @search="handleSessionSearch"
            @reset="handleSessionResetFilters"
            @refresh="handleSessionRefresh"
            @revoke="revokeSession"
            @page-change="handleSessionPageChange"
            @page-size-change="handleSessionPageSizeChange"
          />
        </el-tab-pane>
      </el-tabs>
    </QueryState>

    <UserEditorDialog
      v-model:visible="userEditorVisible"
      :title="userEditorTitle"
      :is-create-mode="isUserCreateMode"
      :form="userEditorForm"
      :rules="userEditorRules"
      :role-options="roleOptions"
      :can-read-role-options="canReadRoleOptions"
      :submitting="userSubmitting"
      @submit="submitUserEditor"
      @closed="handleUserEditorClosed"
    />

    <PasswordResetDialog
      v-model:visible="passwordVisible"
      :target-label="passwordTargetLabel"
      :form="passwordForm"
      :rules="passwordRules"
      :submitting="passwordSubmitting"
      @submit="submitPasswordReset"
      @closed="handlePasswordClosed"
    />

    <RoleEditorDialog
      v-model:visible="roleEditorVisible"
      :title="roleEditorTitle"
      :is-create-mode="isRoleCreateMode"
      :is-built-in-role="isBuiltInRole"
      :form="roleEditorForm"
      :rules="roleEditorRules"
      :permission-groups="permissionGroups"
      :submitting="roleSubmitting"
      @submit="submitRoleEditor"
      @closed="handleRoleEditorClosed"
    />
  </div>
</template>

<style scoped>
.admin-settings-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.admin-settings-page__header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.admin-settings-page__header p {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
</style>
