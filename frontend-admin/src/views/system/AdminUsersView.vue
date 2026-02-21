<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, useDialog, NTag, NButton, NSpace } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { AddOutline } from '@vicons/ionicons5'
import {
  listAdminUsers,
  createAdminUser,
  updateAdminUser,
  setAdminUserStatus,
  deleteAdminUser,
} from '@/api/adminUser'
import { listRoles } from '@/api/role'
import type { AdminUser } from '@/types'
import { useAuthStore } from '@/store/auth'

const message = useMessage()
const dialog = useDialog()
const authStore = useAuthStore()

// ── 列表状态 ──────────────────────────────────────────────
const loading = ref(false)
const tableData = ref<AdminUser[]>([])
const total = ref(0)

const query = reactive({
  keyword: '',
  page_num: 1,
  page_size: 20,
})

// ── 弹窗状态 ──────────────────────────────────────────────
const modalVisible = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const modalLoading = ref(false)
const editingId = ref<number | null>(null)

const formRef = ref()
const form = reactive({
  username: '',
  password: '',
  nickname: '',
  email: '',
  role_ids: [] as number[],
})

// 表单校验规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 64, message: '用户名长度 3~64 位', trigger: 'blur' },
  ],
  password: [
    {
      required: true,
      validator: (_rule: unknown, value: string) => {
        // 编辑模式下密码可为空（不修改）
        if (modalMode.value === 'edit' && !value) return true
        if (!value) return new Error('请输入密码')
        if (value.length < 6) return new Error('密码至少 6 位')
        return true
      },
      trigger: 'blur',
    },
  ],
  role_ids: [
    { type: 'array', min: 1, message: '至少选择一个角色', trigger: 'change' },
  ],
}

// ── 角色选项（从角色接口加载，保证显示全部角色）──────────
const roleOptions = ref<{ label: string; value: number }[]>([])

async function loadRoleOptions() {
  try {
    const res = await listRoles({ page_num: 1, page_size: 200 })
    roleOptions.value = res.data.data.list.map((r) => ({
      label: r.label,
      value: r.id,
    }))
  } catch {
    // 加载失败不阻塞主流程，选项保持空即可
  }
}

// ── 数据加载 ──────────────────────────────────────────────
async function loadData() {
  loading.value = true
  try {
    const res = await listAdminUsers(query)
    const { list, total: t } = res.data.data
    tableData.value = list
    total.value = t
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  query.page_num = 1
  loadData()
}

function handleReset() {
  query.keyword = ''
  query.page_num = 1
  loadData()
}

// ── 新建 / 编辑弹窗 ───────────────────────────────────────
function openCreate() {
  modalMode.value = 'create'
  editingId.value = null
  Object.assign(form, { username: '', password: '', nickname: '', email: '', role_ids: [] })
  modalVisible.value = true
}

function openEdit(row: AdminUser) {
  modalMode.value = 'edit'
  editingId.value = row.id
  Object.assign(form, {
    username: row.username,
    password: '',
    nickname: row.nickname,
    // email 可能为 null（后端允许 NULL），表单里用空字符串表示“未填写”
    email: row.email ?? '',
    role_ids: row.roles?.map((r) => r.id) ?? [],
  })
  modalVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await createAdminUser({
        username: form.username,
        password: form.password,
        nickname: form.nickname,
        email: form.email,
        role_ids: form.role_ids,
      })
      message.success('创建成功')
    } else {
      const payload: Record<string, unknown> = {
        nickname: form.nickname,
        email: form.email,
        role_ids: form.role_ids,
      }
      if (form.password) payload.password = form.password
      await updateAdminUser(editingId.value!, payload)
      message.success('更新成功')
    }
    modalVisible.value = false
    loadData()
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '操作失败')
  } finally {
    modalLoading.value = false
  }
}

// ── 状态切换 ──────────────────────────────────────────────
async function handleToggleStatus(row: AdminUser) {
  const next = row.status === 1 ? 0 : 1
  const label = next === 1 ? '启用' : '禁用'
  try {
    await setAdminUserStatus(row.id, next as 0 | 1)
    message.success(`已${label}`)
    loadData()
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '操作失败')
  }
}

// ── 删除 ──────────────────────────────────────────────────
function handleDelete(row: AdminUser) {
  dialog.warning({
    title: '确认删除',
    content: `确认删除管理员「${row.nickname || row.username}」？此操作不可恢复。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteAdminUser(row.id)
        message.success('已删除')
        loadData()
      } catch (err: unknown) {
        message.error(err instanceof Error ? err.message : '删除失败')
      }
    },
  })
}

// ── 表格列定义 ────────────────────────────────────────────
const columns: DataTableColumns<AdminUser> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '用户名', key: 'username', width: 140 },
  {
    title: '昵称',
    key: 'nickname',
    render: (row) => row.nickname || '—',
  },
  {
    title: '邮箱',
    key: 'email',
    render: (row) => row.email || '—',
  },
  {
    title: '角色',
    key: 'roles',
    render: (row) => {
      if (!row.roles?.length) return '—'
      return row.roles.map((r) => r.label).join('、')
    },
  },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small', bordered: false }, {
        default: () => (row.status === 1 ? '启用' : '禁用'),
      }),
  },
  {
    title: '最后登录',
    key: 'last_login_at',
    width: 160,
    render: (row) =>
      row.last_login_at
        ? new Date(row.last_login_at).toLocaleString('zh-CN', {
            month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit',
          })
        : '从未登录',
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 160,
    render: (row) =>
      new Date(row.created_at).toLocaleString('zh-CN', {
        year: 'numeric', month: '2-digit', day: '2-digit',
      }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    fixed: 'right',
    render: (row) => {
      // 按权限决定渲染哪些按钮
      const btns = []
      if (authStore.hasPermission('admin:update')) {
        btns.push(h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }))
      }
      if (authStore.hasPermission('admin:status')) {
        btns.push(h(
          NButton,
          { size: 'small', type: row.status === 1 ? 'warning' : 'success', onClick: () => handleToggleStatus(row) },
          { default: () => (row.status === 1 ? '禁用' : '启用') },
        ))
      }
      if (authStore.hasPermission('admin:delete')) {
        btns.push(h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' }))
      }
      return h(NSpace, null, { default: () => btns })
    },
  },
]

onMounted(() => {
  loadData()
  loadRoleOptions()
})
</script>

<template>
  <div class="admin-users">
    <!-- 搜索栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <n-input
          v-model:value="query.keyword"
          placeholder="搜索用户名或昵称"
          clearable
          style="width: 240px;"
          @keyup.enter="handleSearch"
        />
        <n-button @click="handleSearch">搜索</n-button>
        <n-button @click="handleReset">重置</n-button>
      </div>
      <n-button v-permission="'admin:create'" type="primary" @click="openCreate">
        <template #icon><n-icon><AddOutline /></n-icon></template>
        新建管理员
      </n-button>
    </div>

    <!-- 数据表格 -->
    <n-card :bordered="false" class="table-card">
      <n-data-table
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :pagination="false"
        :scroll-x="1000"
        size="small"
        striped
      />

      <!-- 分页 -->
      <div class="pagination">
        <n-pagination
          v-model:page="query.page_num"
          v-model:page-size="query.page_size"
          :item-count="total"
          :page-sizes="[20, 50, 100]"
          show-size-picker
          show-quick-jumper
          @update:page="loadData"
          @update:page-size="() => { query.page_num = 1; loadData() }"
        />
      </div>
    </n-card>

    <!-- 新建/编辑弹窗 -->
    <n-modal
      v-model:show="modalVisible"
      :title="modalMode === 'create' ? '新建管理员' : '编辑管理员'"
      preset="card"
      style="width: 480px;"
      :mask-closable="false"
    >
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80">
        <n-form-item label="用户名" path="username">
          <n-input
            v-model:value="form.username"
            placeholder="登录用户名"
            :disabled="modalMode === 'edit'"
          />
        </n-form-item>

        <n-form-item :label="modalMode === 'create' ? '密码' : '新密码'" path="password">
          <n-input
            v-model:value="form.password"
            type="password"
            show-password-on="click"
            :placeholder="modalMode === 'create' ? '至少 6 位' : '不填则不修改'"
          />
        </n-form-item>

        <n-form-item label="昵称" path="nickname">
          <n-input v-model:value="form.nickname" placeholder="显示名称（选填）" />
        </n-form-item>

        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="form.email" placeholder="邮箱地址（选填）" />
        </n-form-item>

        <n-form-item label="角色" path="role_ids">
          <n-select
            v-model:value="form.role_ids"
            :options="roleOptions"
            multiple
            placeholder="请选择角色"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="modalVisible = false">取消</n-button>
          <n-button type="primary" :loading="modalLoading" @click="handleSubmit">
            {{ modalMode === 'create' ? '创建' : '保存' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.admin-users {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 搜索工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  padding: 16px 20px;
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 表格卡片 */
.table-card {
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
