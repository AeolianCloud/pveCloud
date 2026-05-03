<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormRules } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'

import QueryState from '../../components/QueryState.vue'
import { usePermissionStore } from '../../store/modules/permission'
import {
  createWebUser,
  getWebUserSessions,
  getWebUsers,
  resetWebUserPassword,
  revokeWebUserSession,
  updateWebUser,
  type WebUserItem,
  type WebUserSessionItem,
} from '../../api/web-user'

type TabKey = 'users' | 'sessions'
type EditorMode = 'create' | 'edit'

interface PaginationState {
  page: number
  per_page: number
  total: number
  last_page: number
}

interface UserFormState {
  username: string
  email: string
  display_name: string
  password: string
  status: string
}

const permissionStore = usePermissionStore()
const activeTab = ref<TabKey>('users')
const initialLoading = ref(false)
const errorMessage = ref('')

const users = ref<WebUserItem[]>([])
const sessions = ref<WebUserSessionItem[]>([])
const userLoading = ref(false)
const sessionLoading = ref(false)
const userSubmitting = ref(false)
const passwordSubmitting = ref(false)
const revokingSessionId = ref('')

const userQuery = reactive({ keyword: '', status: '' })
const sessionQuery = reactive({ user_id: undefined as number | undefined, status: '' })
const userPagination = reactive<PaginationState>({ page: 1, per_page: 15, total: 0, last_page: 0 })
const sessionPagination = reactive<PaginationState>({ page: 1, per_page: 15, total: 0, last_page: 0 })

const editorVisible = ref(false)
const editorMode = ref<EditorMode>('create')
const editingUser = ref<WebUserItem | null>(null)
const userForm = reactive<UserFormState>(defaultUserForm())

const passwordVisible = ref(false)
const passwordTarget = ref<WebUserItem | null>(null)
const passwordForm = reactive({ password: '' })

const canViewUsersTab = computed(() => permissionStore.hasPermission('page.web-users'))
const canViewSessionsTab = computed(() => permissionStore.hasPermission('page.web-user-sessions'))
const canCreateUser = computed(() => permissionStore.hasPermission('web-user:create'))
const canUpdateUser = computed(() => permissionStore.hasPermission('web-user:update'))
const canResetPassword = computed(() => permissionStore.hasPermission('web-user:password-reset'))
const canRevokeSession = computed(() => permissionStore.hasPermission('web-user-session:revoke'))
const isCreateMode = computed(() => editorMode.value === 'create')
const editorTitle = computed(() => (isCreateMode.value ? '新建 Web 用户' : '编辑 Web 用户'))

const userRules: FormRules<UserFormState> = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 64, message: '用户名长度需为 3 到 64 个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效邮箱', trigger: 'blur' },
  ],
  password: [
    {
      validator: (_rule, value, callback) => {
        if (!isCreateMode.value && !value) return callback()
        if (!value) return callback(new Error('请输入密码'))
        if (value.length < 6 || value.length > 72) return callback(new Error('密码长度需为 6 到 72 个字符'))
        callback()
      },
      trigger: 'blur',
    },
  ],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

const passwordRules: FormRules<{ password: string }> = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 72, message: '密码长度需为 6 到 72 个字符', trigger: 'blur' },
  ],
}

watch([canViewUsersTab, canViewSessionsTab], syncVisibleTab, { immediate: true })

void initializePage()

async function initializePage() {
  initialLoading.value = true
  errorMessage.value = ''
  try {
    const tasks: Promise<unknown>[] = []
    if (canViewUsersTab.value) tasks.push(loadUsers())
    if (canViewSessionsTab.value) tasks.push(loadSessions())
    await Promise.all(tasks)
    syncVisibleTab()
  } catch (error) {
    errorMessage.value = toError(error, 'Web 用户管理加载失败')
  } finally {
    initialLoading.value = false
  }
}

async function loadUsers() {
  const result = await getWebUsers({
    page: userPagination.page,
    per_page: userPagination.per_page,
    keyword: normalizeKeyword(userQuery.keyword),
    status: userQuery.status || undefined,
  })
  users.value = result.list
  Object.assign(userPagination, { page: result.page, per_page: result.per_page, total: result.total, last_page: result.last_page })
}

async function loadSessions() {
  const result = await getWebUserSessions({
    page: sessionPagination.page,
    per_page: sessionPagination.per_page,
    user_id: sessionQuery.user_id,
    status: sessionQuery.status || undefined,
  })
  sessions.value = result.list
  Object.assign(sessionPagination, { page: result.page, per_page: result.per_page, total: result.total, last_page: result.last_page })
}

async function refreshUsers() {
  userLoading.value = true
  try {
    await loadUsers()
    ElMessage.success('Web 用户已刷新')
  } catch (error) {
    ElMessage.error(toError(error, '刷新失败'))
  } finally {
    userLoading.value = false
  }
}

async function refreshSessions() {
  sessionLoading.value = true
  try {
    await loadSessions()
    ElMessage.success('用户状态已刷新')
  } catch (error) {
    ElMessage.error(toError(error, '刷新失败'))
  } finally {
    sessionLoading.value = false
  }
}

function searchUsers() {
  userPagination.page = 1
  void refreshUsers()
}

function searchSessions() {
  sessionPagination.page = 1
  void refreshSessions()
}

function openCreateDialog() {
  editorMode.value = 'create'
  editingUser.value = null
  Object.assign(userForm, defaultUserForm())
  editorVisible.value = true
}

function openEditDialog(user: WebUserItem) {
  editorMode.value = 'edit'
  editingUser.value = user
  Object.assign(userForm, {
    username: user.username,
    email: user.email,
    display_name: user.display_name ?? '',
    password: '',
    status: user.status,
  })
  editorVisible.value = true
}

async function submitUser() {
  userSubmitting.value = true
  try {
    if (isCreateMode.value) {
      await createWebUser({ ...userForm, display_name: optionalText(userForm.display_name) })
      ElMessage.success('Web 用户已创建')
    } else if (editingUser.value) {
      await updateWebUser(editingUser.value.id, {
        email: userForm.email,
        display_name: optionalText(userForm.display_name),
        status: userForm.status,
      })
      ElMessage.success('Web 用户已更新')
    }
    editorVisible.value = false
    await loadUsers()
  } catch (error) {
    ElMessage.error(toError(error, '保存失败'))
  } finally {
    userSubmitting.value = false
  }
}

function openPasswordDialog(user: WebUserItem) {
  passwordTarget.value = user
  passwordForm.password = ''
  passwordVisible.value = true
}

async function submitPassword() {
  if (!passwordTarget.value) return
  passwordSubmitting.value = true
  try {
    await resetWebUserPassword(passwordTarget.value.id, passwordForm.password)
    ElMessage.success('密码已重置')
    passwordVisible.value = false
  } catch (error) {
    ElMessage.error(toError(error, '重置失败'))
  } finally {
    passwordSubmitting.value = false
  }
}

async function revokeSession(row: WebUserSessionItem) {
  try {
    await ElMessageBox.confirm(`确认吊销用户 ${row.user.username} 的当前会话？`, '吊销会话', { type: 'warning' })
    revokingSessionId.value = row.session_id
    await revokeWebUserSession(row.session_id)
    ElMessage.success('会话已吊销')
    await loadSessions()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(toError(error, '吊销失败'))
  } finally {
    revokingSessionId.value = ''
  }
}

function syncVisibleTab() {
  if (activeTab.value === 'users' && !canViewUsersTab.value && canViewSessionsTab.value) activeTab.value = 'sessions'
  if (activeTab.value === 'sessions' && !canViewSessionsTab.value && canViewUsersTab.value) activeTab.value = 'users'
}

function statusType(status: string) {
  return status === 'active' ? 'success' : status === 'disabled' || status === 'revoked' ? 'danger' : 'warning'
}

function defaultUserForm(): UserFormState {
  return { username: '', email: '', display_name: '', password: '', status: 'active' }
}

function normalizeKeyword(value: string) {
  const trimmed = value.trim()
  return trimmed || undefined
}

function optionalText(value: string) {
  const trimmed = value.trim()
  return trimmed || null
}

function toError(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback
}
</script>

<template>
  <div class="web-users-page">
    <div class="web-users-page__header">
      <h2>Web 用户管理</h2>
    </div>

    <QueryState :loading="initialLoading" :error-message="errorMessage" @retry="initializePage">
      <el-card>
        <el-tabs v-model="activeTab">
          <el-tab-pane v-if="canViewUsersTab" label="Web 用户" name="users">
            <div class="toolbar">
              <el-input v-model="userQuery.keyword" clearable placeholder="用户名 / 邮箱 / 显示名称" @keyup.enter="searchUsers" />
              <el-select v-model="userQuery.status" clearable placeholder="状态">
                <el-option label="启用" value="active" />
                <el-option label="禁用" value="disabled" />
              </el-select>
              <el-button type="primary" @click="searchUsers">查询</el-button>
              <el-button :icon="Refresh" :loading="userLoading" @click="refreshUsers">刷新</el-button>
              <el-button v-if="canCreateUser" type="success" @click="openCreateDialog">新建用户</el-button>
            </div>

            <el-table v-loading="userLoading" :data="users" stripe>
              <el-table-column label="用户名" prop="username" min-width="140" />
              <el-table-column label="邮箱" prop="email" min-width="190" />
              <el-table-column label="显示名称" min-width="140">
                <template #default="{ row }">{{ row.display_name || '-' }}</template>
              </el-table-column>
              <el-table-column label="状态" width="100" align="center">
                <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ row.status }}</el-tag></template>
              </el-table-column>
              <el-table-column label="创建时间" prop="created_at" min-width="180" />
              <el-table-column label="操作" width="190" fixed="right">
                <template #default="{ row }">
                  <el-button v-if="canUpdateUser" size="small" @click="openEditDialog(row)">编辑</el-button>
                  <el-button v-if="canResetPassword" size="small" type="warning" @click="openPasswordDialog(row)">重置密码</el-button>
                </template>
              </el-table-column>
            </el-table>

            <el-pagination v-model:current-page="userPagination.page" v-model:page-size="userPagination.per_page" background layout="total, sizes, prev, pager, next" :total="userPagination.total" :page-sizes="[15, 30, 50, 100]" @change="refreshUsers" />
          </el-tab-pane>

          <el-tab-pane v-if="canViewSessionsTab" label="用户状态" name="sessions">
            <div class="toolbar">
              <el-input-number v-model="sessionQuery.user_id" :min="1" controls-position="right" placeholder="用户 ID" />
              <el-select v-model="sessionQuery.status" clearable placeholder="状态">
                <el-option label="活跃" value="active" />
                <el-option label="已吊销" value="revoked" />
                <el-option label="已过期" value="expired" />
              </el-select>
              <el-button type="primary" @click="searchSessions">查询</el-button>
              <el-button :icon="Refresh" :loading="sessionLoading" @click="refreshSessions">刷新</el-button>
            </div>

            <el-table v-loading="sessionLoading" :data="sessions" stripe>
              <el-table-column label="用户" min-width="180">
                <template #default="{ row }">{{ row.user.username }} / {{ row.user.email }}</template>
              </el-table-column>
              <el-table-column label="状态" width="100" align="center">
                <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ row.status }}</el-tag></template>
              </el-table-column>
              <el-table-column label="签发时间" prop="issued_at" min-width="180" />
              <el-table-column label="过期时间" prop="expires_at" min-width="180" />
              <el-table-column label="最近 IP" min-width="140"><template #default="{ row }">{{ row.last_seen_ip || '-' }}</template></el-table-column>
              <el-table-column label="User-Agent" min-width="220" show-overflow-tooltip><template #default="{ row }">{{ row.user_agent || '-' }}</template></el-table-column>
              <el-table-column label="操作" width="110" fixed="right">
                <template #default="{ row }">
                  <el-button v-if="canRevokeSession && row.status === 'active'" size="small" type="danger" :loading="revokingSessionId === row.session_id" @click="revokeSession(row)">吊销</el-button>
                </template>
              </el-table-column>
            </el-table>

            <el-pagination v-model:current-page="sessionPagination.page" v-model:page-size="sessionPagination.per_page" background layout="total, sizes, prev, pager, next" :total="sessionPagination.total" :page-sizes="[15, 30, 50, 100]" @change="refreshSessions" />
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </QueryState>

    <el-dialog v-model="editorVisible" :title="editorTitle" width="520px">
      <el-form :model="userForm" :rules="userRules" label-width="92px">
        <el-form-item label="用户名" prop="username"><el-input v-model="userForm.username" :disabled="!isCreateMode" /></el-form-item>
        <el-form-item label="邮箱" prop="email"><el-input v-model="userForm.email" /></el-form-item>
        <el-form-item label="显示名称" prop="display_name"><el-input v-model="userForm.display_name" /></el-form-item>
        <el-form-item v-if="isCreateMode" label="密码" prop="password"><el-input v-model="userForm.password" type="password" show-password /></el-form-item>
        <el-form-item label="状态" prop="status"><el-select v-model="userForm.status"><el-option label="启用" value="active" /><el-option label="禁用" value="disabled" /></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="editorVisible = false">取消</el-button><el-button type="primary" :loading="userSubmitting" @click="submitUser">保存</el-button></template>
    </el-dialog>

    <el-dialog v-model="passwordVisible" title="重置密码" width="420px">
      <el-form :model="passwordForm" :rules="passwordRules" label-width="80px">
        <el-form-item label="新密码" prop="password"><el-input v-model="passwordForm.password" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer><el-button @click="passwordVisible = false">取消</el-button><el-button type="primary" :loading="passwordSubmitting" @click="submitPassword">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.web-users-page { display: flex; flex-direction: column; gap: 16px; }
.web-users-page__header { display: flex; align-items: center; justify-content: space-between; }
.web-users-page__header h2 { margin: 0; font-size: 18px; font-weight: 600; }
.toolbar { display: flex; flex-wrap: wrap; gap: 12px; margin-bottom: 16px; }
.toolbar .el-input { width: 260px; }
.toolbar .el-select { width: 140px; }
.el-pagination { margin-top: 16px; justify-content: flex-end; }
</style>
