<script setup lang="ts">
import { EditPen, Key, Plus, Refresh, Search, SwitchButton } from '@element-plus/icons-vue'
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import {
  ElMessage,
  ElMessageBox,
  type FilterNodeMethodFunction,
  type FormInstance,
  type FormRules,
  type TreeInstance,
} from 'element-plus'

import EmptyState from '../../components/EmptyState.vue'
import QueryState from '../../components/QueryState.vue'
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
import { usePermissionStore } from '../../store/modules/permission'

type EditorMode = 'create' | 'edit'
type AdminStatus = 'active' | 'disabled'
type RoleStatus = 'active' | 'disabled'

interface UserQueryFormState {
  keyword: string
  status: '' | AdminStatus
  role_id: number | undefined
}

interface RoleQueryFormState {
  keyword: string
  status: '' | RoleStatus
}

interface UserEditorState {
  username: string
  email: string
  display_name: string
  password: string
  status: AdminStatus
  role_ids: number[]
}

interface UserEditorSnapshot {
  email: string
  display_name: string
  status: AdminStatus
  role_ids: number[]
}

interface PasswordFormState {
  password: string
}

interface RoleEditorState {
  code: string
  name: string
  description: string
  status: RoleStatus
  permission_codes: string[]
}

interface RoleEditorSnapshot {
  name: string
  description: string
  status: RoleStatus
  permission_codes: string[]
}

interface PermissionTreeNode {
  id: string
  label: string
  type: 'group' | 'permission'
  code?: string
  count?: number
  description?: string | null
  children?: PermissionTreeNode[]
  disabled?: boolean
}

const activeTab = ref<'users' | 'roles'>('users')
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

const users = ref<AdminUserItem[]>([])
const roleOptions = ref<AdminRoleItem[]>([])
const roles = ref<AdminRoleItem[]>([])
const permissionGroups = ref<AdminPermissionGroup[]>([])

const userPagination = reactive({
  page: 1,
  per_page: 15,
  total: 0,
  last_page: 0,
})

const rolePagination = reactive({
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

const userEditorVisible = ref(false)
const userEditorMode = ref<EditorMode>('create')
const editingUser = ref<AdminUserItem | null>(null)
const userEditorFormRef = ref<FormInstance>()
const userEditorForm = reactive<UserEditorState>(createDefaultUserEditorForm())
const userEditorSnapshot = ref<UserEditorSnapshot | null>(null)

const passwordVisible = ref(false)
const passwordTarget = ref<AdminUserItem | null>(null)
const passwordFormRef = ref<FormInstance>()
const passwordForm = reactive<PasswordFormState>({
  password: '',
})

const roleEditorVisible = ref(false)
const roleEditorMode = ref<EditorMode>('create')
const editingRole = ref<AdminRoleItem | null>(null)
const roleEditorFormRef = ref<FormInstance>()
const rolePermissionTreeRef = ref<TreeInstance>()
const permissionFilterText = ref('')
const roleEditorForm = reactive<RoleEditorState>(createDefaultRoleEditorForm())
const roleEditorSnapshot = ref<RoleEditorSnapshot | null>(null)

const hasUsers = computed(() => users.value.length > 0)
const hasRoles = computed(() => roles.value.length > 0)
const activeRoleOptions = computed(() => roleOptions.value.filter((role) => role.status === 'active'))
const canViewUsersTab = computed(() => permissionStore.hasPermission('page.system-settings.admin-users'))
const canViewRolesTab = computed(() => permissionStore.hasPermission('page.system-settings.admin-roles'))
const canViewUsersResource = computed(() => permissionStore.hasPermission('admin-user:view'))
const canViewRolesResource = computed(() => permissionStore.hasPermission('admin-role:view'))
const canReadRoleOptions = computed(() => permissionStore.hasPermission('admin-role:view'))
const canCreateUser = computed(() => permissionStore.hasPermission('admin-user:create'))
const canUpdateUser = computed(() => permissionStore.hasPermission('admin-user:update'))
const canResetUserPassword = computed(() => permissionStore.hasPermission('admin-user:password-reset'))
const canCreateRole = computed(() => permissionStore.hasPermission('admin-role:create'))
const canUpdateRole = computed(() => permissionStore.hasPermission('admin-role:update'))

const isUserCreateMode = computed(() => userEditorMode.value === 'create')
const userEditorTitle = computed(() => (isUserCreateMode.value ? '新建管理员' : '编辑管理员'))

const isRoleCreateMode = computed(() => roleEditorMode.value === 'create')
const roleEditorTitle = computed(() => (isRoleCreateMode.value ? '新建管理组' : '编辑管理组'))
const isBuiltInRole = computed(() => editingRole.value?.code === 'super_admin')
const rolePermissionCount = computed(() => roleEditorForm.permission_codes.length)
const permissionTreeData = computed<PermissionTreeNode[]>(() =>
  permissionGroups.value.map((group) => ({
    id: `group:${group.group_name}`,
    label: group.group_name,
    type: 'group',
    count: group.permissions.length,
    disabled: isBuiltInRole.value,
    children: group.permissions.map((permission) => ({
      id: permission.code,
      label: permission.name,
      type: 'permission',
      code: permission.code,
      description: permission.description,
      disabled: isBuiltInRole.value,
    })),
  })),
)

const permissionTreeProps = {
  children: 'children',
  label: 'label',
  disabled: 'disabled',
}

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
      tasks.push(loadPermissionGroups(), loadRolesData(), loadRoleOptions())
    }
    await Promise.all(tasks)
    syncVisibleTab()
  } catch (error) {
    errorMessage.value = toErrorMessage(error, '管理员设置加载失败')
  } finally {
    initialLoading.value = false
  }
}

watch(permissionFilterText, (value) => {
  rolePermissionTreeRef.value?.filter(value)
})

watch([canViewUsersTab, canViewRolesTab], () => {
  syncVisibleTab()
}, { immediate: true })

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

async function reloadRoleDataForAllViews() {
  await Promise.all([loadRoleOptions(), loadRolesData()])
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

function openCreateUserDialog() {
  userEditorMode.value = 'create'
  editingUser.value = null
  resetUserEditorForm()
  userEditorVisible.value = true
  void nextTick(() => {
    userEditorFormRef.value?.clearValidate()
  })
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
  void nextTick(() => {
    userEditorFormRef.value?.clearValidate()
  })
}

function handleUserEditorClosed() {
  resetUserEditorForm()
}

function syncVisibleTab() {
  if (activeTab.value === 'users' && !canViewUsersTab.value && canViewRolesTab.value) {
    activeTab.value = 'roles'
    return
  }
  if (activeTab.value === 'roles' && !canViewRolesTab.value && canViewUsersTab.value) {
    activeTab.value = 'users'
  }
}

async function submitUserEditor() {
  if (!userEditorFormRef.value) {
    return
  }
  await userEditorFormRef.value.validate()

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
  void nextTick(() => {
    passwordFormRef.value?.clearValidate()
  })
}

function handlePasswordClosed() {
  passwordTarget.value = null
  passwordForm.password = ''
}

async function submitPasswordReset() {
  if (!passwordFormRef.value || !passwordTarget.value) {
    return
  }
  await passwordFormRef.value.validate()

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
  void nextTick(() => {
    roleEditorFormRef.value?.clearValidate()
    syncRolePermissionTree()
  })
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
  void nextTick(() => {
    roleEditorFormRef.value?.clearValidate()
    syncRolePermissionTree()
  })
}

function handleRoleEditorClosed() {
  resetRoleEditorForm()
}

const filterPermissionNode: FilterNodeMethodFunction = (value, rawData) => {
  const data = rawData as PermissionTreeNode
  if (!value) {
    return true
  }
  const keyword = String(value).trim().toLowerCase()
  if (!keyword) {
    return true
  }
  return [data.label, data.code, data.description]
    .filter(Boolean)
    .some((item) => String(item).toLowerCase().includes(keyword))
}

function handlePermissionTreeCheck(_data: PermissionTreeNode, checked: { checkedKeys: unknown[] }) {
  roleEditorForm.permission_codes = uniqueSortedStrings(
    checked.checkedKeys.filter((key): key is string => typeof key === 'string' && !key.startsWith('group:')),
  )
}

function handleCheckAllPermissions() {
  if (isBuiltInRole.value) {
    return
  }
  const tree = rolePermissionTreeRef.value
  if (!tree) {
    return
  }
  const keys = permissionGroups.value.flatMap((group) => group.permissions.map((permission) => permission.code))
  tree.setCheckedKeys(keys, false)
  roleEditorForm.permission_codes = uniqueSortedStrings(keys)
}

function handleClearPermissions() {
  if (isBuiltInRole.value) {
    return
  }
  rolePermissionTreeRef.value?.setCheckedKeys([], false)
  roleEditorForm.permission_codes = []
}

async function submitRoleEditor() {
  if (!roleEditorFormRef.value) {
    return
  }
  await roleEditorFormRef.value.validate()

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

function resetUserEditorForm() {
  Object.assign(userEditorForm, createDefaultUserEditorForm())
  userEditorSnapshot.value = null
  userEditorFormRef.value?.clearValidate()
}

function resetRoleEditorForm() {
  Object.assign(roleEditorForm, createDefaultRoleEditorForm())
  roleEditorSnapshot.value = null
  permissionFilterText.value = ''
  roleEditorFormRef.value?.clearValidate()
  rolePermissionTreeRef.value?.setCheckedKeys([], false)
}

function syncRolePermissionTree() {
  const keys = roleEditorForm.permission_codes.filter((code) => !code.startsWith('group:'))
  rolePermissionTreeRef.value?.setCheckedKeys(keys, false)
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

function statusTagType(status: string) {
  return status === 'active' ? 'success' : 'info'
}

function formatRoleOptionLabel(role: AdminRoleItem) {
  return role.status === 'active' ? role.name : `${role.name}（已停用）`
}

function formatDateTime(value: string | null) {
  if (!value) {
    return '-'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(date)
}

function toErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error && error.message.trim() ? error.message : fallback
}

onMounted(() => {
  void initializePage()
})
</script>

<template>
  <div class="admin-settings-page">
    <div class="admin-settings-page__header">
      <div>
        <h2>管理员设置</h2>
        <p>在同一页面管理管理员账号、管理组和权限分配规则。</p>
      </div>
    </div>

    <QueryState :loading="initialLoading" :error-message="errorMessage" @retry="initializePage">
      <el-tabs v-model="activeTab">
        <el-tab-pane v-if="canViewUsersTab" label="管理员账号" name="users">
          <el-card v-loading="userTableLoading || roleOptionsLoading" shadow="never" class="admin-settings-page__card">
            <div class="admin-settings-page__toolbar">
              <el-form inline class="admin-settings-page__filters" @submit.prevent>
                <el-form-item label="关键字">
                  <el-input
                    v-model="userQueryForm.keyword"
                    clearable
                    placeholder="搜索账号、邮箱或显示名称"
                    @keyup.enter="handleUserSearch"
                  />
                </el-form-item>
                <el-form-item label="状态">
                  <el-select v-model="userQueryForm.status" clearable placeholder="全部状态">
                    <el-option label="启用" value="active" />
                    <el-option label="停用" value="disabled" />
                  </el-select>
                </el-form-item>
                <el-form-item v-if="canViewRolesTab" label="角色">
                  <el-select v-model="userQueryForm.role_id" clearable filterable placeholder="全部角色">
                    <el-option
                      v-for="role in roleOptions"
                      :key="role.id"
                      :label="formatRoleOptionLabel(role)"
                      :value="role.id"
                    />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" :icon="Search" @click="handleUserSearch">查询</el-button>
                  <el-button @click="handleUserResetFilters">重置</el-button>
                </el-form-item>
              </el-form>

              <div class="admin-settings-page__toolbar-actions">
                <el-button :icon="Refresh" :loading="userRefreshing" @click="handleUserRefresh">刷新</el-button>
                <el-button v-if="canCreateUser" type="primary" :icon="Plus" @click="openCreateUserDialog">新建管理员</el-button>
              </div>
            </div>

            <template v-if="!canViewUsersResource">
              <EmptyState title="暂无权限" description="当前账号没有管理员账号查看权限。" />
            </template>

            <div v-else-if="hasUsers" class="admin-settings-page__table">
              <el-table :data="users" stripe>
                <el-table-column label="账号" min-width="140">
                  <template #default="{ row }">
                    <div class="admin-settings-page__identity">
                      <span class="admin-settings-page__primary">{{ row.username }}</span>
                      <span class="admin-settings-page__secondary">{{ row.display_name }}</span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="邮箱" prop="email" min-width="220" show-overflow-tooltip>
                  <template #default="{ row }">
                    {{ row.email || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="状态" width="100" align="center">
                  <template #default="{ row }">
                    <el-tag :type="statusTagType(row.status)" size="small">
                      {{ formatStatusLabel(row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="角色" min-width="220">
                  <template #default="{ row }">
                    <div v-if="row.roles.length > 0" class="admin-settings-page__tags">
                      <el-tag v-for="role in row.roles" :key="role.id" size="small" effect="plain">
                        {{ role.name }}
                      </el-tag>
                    </div>
                    <span v-else class="admin-settings-page__secondary">未分配角色</span>
                  </template>
                </el-table-column>
                <el-table-column label="最后登录" min-width="220">
                  <template #default="{ row }">
                    <div class="admin-settings-page__meta">
                      <span>{{ formatDateTime(row.last_login_at) }}</span>
                      <span>{{ row.last_login_ip || '-' }}</span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="创建时间" min-width="180">
                  <template #default="{ row }">
                    {{ formatDateTime(row.created_at) }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="260" fixed="right">
                  <template #default="{ row }">
                    <div class="admin-settings-page__actions">
                      <el-button v-if="canUpdateUser" link type="primary" :icon="EditPen" @click="openEditUserDialog(row)">编辑</el-button>
                      <el-button
                        v-if="canUpdateUser"
                        link
                        :type="row.status === 'active' ? 'warning' : 'success'"
                        :icon="SwitchButton"
                        :loading="userStatusUpdatingId === row.id"
                        @click="toggleUserStatus(row)"
                      >
                        {{ row.status === 'active' ? '停用' : '启用' }}
                      </el-button>
                      <el-button v-if="canResetUserPassword" link type="danger" :icon="Key" @click="openPasswordDialog(row)">重置密码</el-button>
                    </div>
                  </template>
                </el-table-column>
              </el-table>

              <div class="admin-settings-page__pagination">
                <el-pagination
                  background
                  layout="total, sizes, prev, pager, next"
                  :current-page="userPagination.page"
                  :page-size="userPagination.per_page"
                  :page-sizes="[15, 30, 50, 100]"
                  :total="userPagination.total"
                  @current-change="handleUserPageChange"
                  @size-change="handleUserPageSizeChange"
                />
              </div>
            </div>

            <EmptyState
              v-else
              title="暂无管理员"
              :description="userQueryForm.keyword || userQueryForm.status || userQueryForm.role_id ? '未找到符合条件的管理员账号。' : '当前还没有可展示的管理员账号。'"
            />
          </el-card>
        </el-tab-pane>

        <el-tab-pane v-if="canViewRolesTab" label="管理组权限" name="roles">
          <el-card v-loading="roleTableLoading" shadow="never" class="admin-settings-page__card">
            <div class="admin-settings-page__toolbar">
              <el-form inline class="admin-settings-page__filters" @submit.prevent>
                <el-form-item label="关键字">
                  <el-input
                    v-model="roleQueryForm.keyword"
                    clearable
                    placeholder="搜索编码、名称或说明"
                    @keyup.enter="handleRoleSearch"
                  />
                </el-form-item>
                <el-form-item label="状态">
                  <el-select v-model="roleQueryForm.status" clearable placeholder="全部状态">
                    <el-option label="启用" value="active" />
                    <el-option label="停用" value="disabled" />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" :icon="Search" @click="handleRoleSearch">查询</el-button>
                  <el-button @click="handleRoleResetFilters">重置</el-button>
                </el-form-item>
              </el-form>

              <div class="admin-settings-page__toolbar-actions">
                <el-button :icon="Refresh" :loading="roleRefreshing" @click="handleRoleRefresh">刷新</el-button>
                <el-button v-if="canCreateRole" type="primary" :icon="Plus" @click="openCreateRoleDialog">新建管理组</el-button>
              </div>
            </div>

            <template v-if="!canViewRolesResource">
              <EmptyState title="暂无权限" description="当前账号没有管理组权限查看权限。" />
            </template>

            <div v-else-if="hasRoles" class="admin-settings-page__table">
              <el-table :data="roles" stripe>
                <el-table-column label="管理组" min-width="220">
                  <template #default="{ row }">
                    <div class="admin-settings-page__identity">
                      <span class="admin-settings-page__primary">{{ row.name }}</span>
                      <span class="admin-settings-page__secondary">{{ row.code }}</span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="说明" min-width="240" show-overflow-tooltip>
                  <template #default="{ row }">
                    {{ row.description || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="状态" width="100" align="center">
                  <template #default="{ row }">
                    <el-tag :type="statusTagType(row.status)" size="small">
                      {{ formatStatusLabel(row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="权限码" min-width="220">
                  <template #default="{ row }">
                    <div class="admin-settings-page__meta">
                      <el-tag size="small" effect="plain">{{ row.permission_codes.length }} 项权限</el-tag>
                      <span class="admin-settings-page__secondary">
                        {{ row.permission_codes.slice(0, 3).join(' / ') || '未分配权限' }}
                      </span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="更新时间" min-width="180">
                  <template #default="{ row }">
                    {{ formatDateTime(row.updated_at) }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="220" fixed="right">
                  <template #default="{ row }">
                    <div class="admin-settings-page__actions">
                      <el-button v-if="canUpdateRole" link type="primary" :icon="EditPen" @click="openEditRoleDialog(row)">编辑</el-button>
                      <el-button
                        v-if="canUpdateRole"
                        link
                        :type="row.status === 'active' ? 'warning' : 'success'"
                        :icon="SwitchButton"
                        :loading="roleStatusUpdatingId === row.id"
                        @click="toggleRoleStatus(row)"
                      >
                        {{ row.status === 'active' ? '停用' : '启用' }}
                      </el-button>
                    </div>
                  </template>
                </el-table-column>
              </el-table>

              <div class="admin-settings-page__pagination">
                <el-pagination
                  background
                  layout="total, sizes, prev, pager, next"
                  :current-page="rolePagination.page"
                  :page-size="rolePagination.per_page"
                  :page-sizes="[15, 30, 50, 100]"
                  :total="rolePagination.total"
                  @current-change="handleRolePageChange"
                  @size-change="handleRolePageSizeChange"
                />
              </div>
            </div>

            <EmptyState
              v-else
              title="暂无管理组"
              :description="roleQueryForm.keyword || roleQueryForm.status ? '未找到符合条件的管理组。' : '当前还没有可展示的管理组。'"
            />
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </QueryState>

    <el-dialog
      v-model="userEditorVisible"
      :title="userEditorTitle"
      width="640px"
      destroy-on-close
      @closed="handleUserEditorClosed"
    >
      <el-form ref="userEditorFormRef" :model="userEditorForm" :rules="userEditorRules" label-width="96px">
        <el-form-item label="登录账号" prop="username">
          <el-input
            v-model="userEditorForm.username"
            :disabled="!isUserCreateMode"
            placeholder="请输入 3 到 64 位账号"
          />
        </el-form-item>
        <el-form-item label="显示名称" prop="display_name">
          <el-input v-model="userEditorForm.display_name" placeholder="请输入管理员显示名称" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userEditorForm.email" placeholder="请输入邮箱，可留空" />
        </el-form-item>
        <el-form-item v-if="isUserCreateMode" label="登录密码" prop="password">
          <el-input
            v-model="userEditorForm.password"
            type="password"
            show-password
            placeholder="请输入 6 到 72 位密码"
          />
        </el-form-item>
        <el-form-item label="账号状态" prop="status">
          <el-radio-group v-model="userEditorForm.status">
            <el-radio value="active">启用</el-radio>
            <el-radio value="disabled">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="canReadRoleOptions" label="角色分配" prop="role_ids">
          <el-select
            v-model="userEditorForm.role_ids"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            placeholder="请选择要分配的角色"
          >
            <el-option
              v-for="role in roleOptions"
              :key="role.id"
              :label="formatRoleOptionLabel(role)"
              :value="role.id"
              :disabled="role.status !== 'active'"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="canReadRoleOptions">
          <el-alert
            type="info"
            :closable="false"
            title="仅启用中的角色可分配给管理员。已停用角色会保留显示，但不能再次分配。"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="userEditorVisible = false">取消</el-button>
        <el-button type="primary" :loading="userSubmitting" @click="submitUserEditor">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="passwordVisible"
      title="重置管理员密码"
      width="480px"
      destroy-on-close
      @closed="handlePasswordClosed"
    >
      <el-alert
        type="warning"
        :closable="false"
        class="admin-settings-page__dialog-alert"
        title="密码重置后会立即生效，请通过安全渠道告知管理员。"
      />
      <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-width="84px">
        <el-form-item label="管理员">
          <el-input :model-value="passwordTarget?.display_name || passwordTarget?.username || '-'" disabled />
        </el-form-item>
        <el-form-item label="新密码" prop="password">
          <el-input
            v-model="passwordForm.password"
            type="password"
            show-password
            placeholder="请输入 6 到 72 位新密码"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="passwordVisible = false">取消</el-button>
        <el-button type="danger" :loading="passwordSubmitting" @click="submitPasswordReset">确认重置</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="roleEditorVisible"
      :title="roleEditorTitle"
      width="760px"
      destroy-on-close
      @closed="handleRoleEditorClosed"
    >
      <el-form ref="roleEditorFormRef" :model="roleEditorForm" :rules="roleEditorRules" label-width="96px">
        <el-form-item label="管理组编码" prop="code">
          <el-input
            v-model="roleEditorForm.code"
            :disabled="!isRoleCreateMode"
            placeholder="请输入唯一编码，例如 ops_manager"
          />
        </el-form-item>
        <el-form-item label="管理组名称" prop="name">
          <el-input v-model="roleEditorForm.name" placeholder="请输入管理组名称" />
        </el-form-item>
        <el-form-item label="说明" prop="description">
          <el-input
            v-model="roleEditorForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入管理组说明，可留空"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="roleEditorForm.status">
            <el-radio value="active" :disabled="isBuiltInRole">启用</el-radio>
            <el-radio value="disabled" :disabled="isBuiltInRole">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="权限分配">
          <div class="admin-settings-page__permission-panel">
            <div class="admin-settings-page__permission-head">
              <span>已选 {{ rolePermissionCount }} 项权限</span>
              <div class="admin-settings-page__permission-tools">
                <el-button link type="primary" :disabled="isBuiltInRole" @click="handleCheckAllPermissions">全选</el-button>
                <el-button link :disabled="isBuiltInRole || rolePermissionCount === 0" @click="handleClearPermissions">清空</el-button>
                <el-tag v-if="isBuiltInRole" type="warning" size="small">内置超级管理员角色不可修改权限</el-tag>
              </div>
            </div>
            <el-input
              v-model="permissionFilterText"
              clearable
              placeholder="筛选权限组、权限名称或权限码"
            />
            <div class="admin-settings-page__permission-tree-wrap">
              <el-tree
                ref="rolePermissionTreeRef"
                :data="permissionTreeData"
                :props="permissionTreeProps"
                node-key="id"
                show-checkbox
                default-expand-all
                :expand-on-click-node="false"
                :check-on-click-node="false"
                :check-on-click-leaf="false"
                :filter-node-method="filterPermissionNode"
                class="admin-settings-page__permission-tree"
                @check="handlePermissionTreeCheck"
              >
                <template #default="{ data }">
                  <div
                    class="admin-settings-page__permission-node"
                    :class="{ 'admin-settings-page__permission-node--group': data.type === 'group' }"
                    :title="data.type === 'permission' ? [data.label, data.code, data.description].filter(Boolean).join(' / ') : data.label"
                  >
                    <span class="admin-settings-page__permission-label">{{ data.label }}</span>
                    <span v-if="data.type === 'permission' && data.code" class="admin-settings-page__permission-code">
                      {{ data.code }}
                    </span>
                    <el-tag v-else-if="data.type === 'group'" size="small" effect="plain">
                      {{ data.count }} 项
                    </el-tag>
                  </div>
                </template>
              </el-tree>
            </div>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="roleEditorVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleSubmitting" @click="submitRoleEditor">保存</el-button>
      </template>
    </el-dialog>
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

.admin-settings-page__card :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.admin-settings-page__toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.admin-settings-page__filters {
  display: flex;
  flex: 1;
  flex-wrap: wrap;
  gap: 8px 0;
}

.admin-settings-page__toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.admin-settings-page__table {
  min-height: 240px;
}

.admin-settings-page__identity,
.admin-settings-page__meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.admin-settings-page__primary {
  color: var(--el-text-color-primary);
  font-weight: 600;
}

.admin-settings-page__secondary {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.admin-settings-page__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.admin-settings-page__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 12px;
}

.admin-settings-page__pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
}

.admin-settings-page__dialog-alert {
  margin-bottom: 16px;
}

.admin-settings-page__permission-panel {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.admin-settings-page__permission-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.admin-settings-page__permission-tools {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 12px;
}

.admin-settings-page__permission-tree-wrap {
  border: 1px solid var(--el-border-color-light);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
  padding: 12px;
  max-height: 420px;
  overflow: auto;
}

.admin-settings-page__permission-tree {
  background: transparent;
}

.admin-settings-page__permission-tree :deep(.el-tree-node__content) {
  min-height: 40px;
  border-radius: 8px;
  padding-right: 8px;
}

.admin-settings-page__permission-tree :deep(.el-tree-node__content:hover) {
  background: var(--el-fill-color-light);
}

.admin-settings-page__permission-node {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.admin-settings-page__permission-node--group {
  font-weight: 600;
}

.admin-settings-page__permission-label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-settings-page__permission-code {
  flex-shrink: 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1;
}

@media (max-width: 960px) {
  .admin-settings-page__toolbar {
    flex-direction: column;
  }

  .admin-settings-page__toolbar-actions {
    width: 100%;
  }

  .admin-settings-page__pagination {
    justify-content: flex-start;
    overflow-x: auto;
  }
}
</style>
