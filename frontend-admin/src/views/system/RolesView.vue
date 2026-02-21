<script setup lang="ts">
import { h, ref, reactive, computed, onMounted } from 'vue'
import { useMessage, useDialog, NButton, NSpace, NTag } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { KeyOutline } from '@vicons/ionicons5'
import { listRoles, createRole, updateRole, deleteRole, assignPermissions } from '@/api/role'
import { listPermissions } from '@/api/permission'
import type { AdminRole } from '@/types'
import type { GroupedPermissions } from '@/api/permission'
import { useAuthStore } from '@/store/auth'

const message = useMessage()
const dialog = useDialog()
const authStore = useAuthStore()

// ── 列表 ──────────────────────────────────────────────────
const loading = ref(false)
const tableData = ref<AdminRole[]>([])
const total = ref(0)

const query = reactive({ keyword: '', page_num: 1, page_size: 20 })

async function loadData() {
  loading.value = true
  try {
    const res = await listRoles(query)
    const { list, total: t } = res.data.data
    tableData.value = list
    total.value = t
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() { query.page_num = 1; loadData() }
function handleReset() { query.keyword = ''; query.page_num = 1; loadData() }

// ── 新建 / 编辑弹窗 ───────────────────────────────────────
const modalVisible = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const modalLoading = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref()

const form = reactive({ name: '', label: '', description: '', sort: 0 })

const rules = {
  name: [
    { required: true, message: '请输入角色标识', trigger: 'blur' },
    { min: 2, max: 64, message: '长度 2~64 位', trigger: 'blur' },
    {
      pattern: /^[a-z0-9_]+$/,
      message: '只允许小写字母、数字和下划线',
      trigger: 'blur',
    },
  ],
  label: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

function openCreate() {
  modalMode.value = 'create'
  editingId.value = null
  Object.assign(form, { name: '', label: '', description: '', sort: 0 })
  modalVisible.value = true
}

function openEdit(row: AdminRole) {
  modalMode.value = 'edit'
  editingId.value = row.id
  Object.assign(form, { name: row.name, label: row.label, description: row.description, sort: row.sort })
  modalVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  modalLoading.value = true
  try {
    if (modalMode.value === 'create') {
      await createRole({ name: form.name, label: form.label, description: form.description, sort: form.sort })
      message.success('创建成功')
    } else {
      await updateRole(editingId.value!, { label: form.label, description: form.description, sort: form.sort })
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

// ── 删除 ──────────────────────────────────────────────────
function handleDelete(row: AdminRole) {
  dialog.warning({
    title: '确认删除',
    content: `确认删除角色「${row.label}」？关联该角色的用户将失去对应权限。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteRole(row.id)
        message.success('已删除')
        loadData()
      } catch (err: unknown) {
        message.error(err instanceof Error ? err.message : '删除失败')
      }
    },
  })
}

// ── 权限分配抽屉 ──────────────────────────────────────────
const drawerVisible = ref(false)
const drawerRole = ref<AdminRole | null>(null)
const drawerLoading = ref(false)
const permGroups = ref<GroupedPermissions[]>([])
// 当前已勾选的权限 ID 集合
const checkedPermIds = ref<Set<number>>(new Set())

async function openPermDrawer(row: AdminRole) {
  drawerRole.value = row
  drawerVisible.value = true
  drawerLoading.value = true
  try {
    // 并行加载全部权限 + 角色当前权限
    const [permRes, roleRes] = await Promise.all([
      listPermissions(),
      listRoles({ page_num: 1, page_size: 1, keyword: '' }),  // 仅用于获取该角色最新权限
    ])
    permGroups.value = permRes.data.data

    // 从列表中重新查找该角色的最新权限（含 permissions 字段）
    const fullRes = await listRoles({ page_num: 1, page_size: 9999 })
    const found = fullRes.data.data.list.find((r) => r.id === row.id)
    checkedPermIds.value = new Set((found?.permissions ?? []).map((p) => p.id))
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '加载权限失败')
  } finally {
    drawerLoading.value = false
  }
}

// 某个 group 下是否全选
function isGroupChecked(group: GroupedPermissions): boolean {
  return group.permissions.every((p) => checkedPermIds.value.has(p.id))
}

// 某个 group 下是否部分选中
function isGroupIndeterminate(group: GroupedPermissions): boolean {
  const checked = group.permissions.filter((p) => checkedPermIds.value.has(p.id)).length
  return checked > 0 && checked < group.permissions.length
}

// 切换整个 group
function toggleGroup(group: GroupedPermissions, checked: boolean) {
  group.permissions.forEach((p) => {
    if (checked) checkedPermIds.value.add(p.id)
    else checkedPermIds.value.delete(p.id)
  })
}

// 切换单个权限
function togglePerm(id: number, checked: boolean) {
  if (checked) checkedPermIds.value.add(id)
  else checkedPermIds.value.delete(id)
}

// 已选数量
const checkedCount = computed(() => checkedPermIds.value.size)

async function handleSavePermissions() {
  if (!drawerRole.value) return
  drawerLoading.value = true
  try {
    await assignPermissions(drawerRole.value.id, Array.from(checkedPermIds.value))
    message.success('权限保存成功')
    drawerVisible.value = false
    loadData()
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '保存失败')
  } finally {
    drawerLoading.value = false
  }
}

// group 中文映射
const groupLabelMap: Record<string, string> = {
  admin: '管理员账号',
  role:  '角色管理',
  log:   '登录日志',
  op:    '操作日志',
}
function groupLabel(g: string) { return groupLabelMap[g] ?? g }

// ── 表格列 ────────────────────────────────────────────────
const columns: DataTableColumns<AdminRole> = [
  { title: 'ID',    key: 'id',    width: 80 },
  { title: '标识',  key: 'name',  width: 160 },
  { title: '名称',  key: 'label', width: 120 },
  {
    title: '描述',
    key: 'description',
    render: (row) => row.description || '—',
  },
  { title: '排序', key: 'sort', width: 80 },
  {
    title: '权限数',
    key: 'permissions',
    width: 90,
    render: (row) =>
      h(NTag, { size: 'small', type: 'info', bordered: false }, {
        default: () => `${row.permissions?.length ?? 0} 项`,
      }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    fixed: 'right',
    render: (row) => {
      const btns = []
      if (authStore.hasPermission('role:update')) {
        btns.push(h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }))
      }
      if (authStore.hasPermission('role:assign')) {
        btns.push(h(
          NButton,
          { size: 'small', type: 'primary', onClick: () => openPermDrawer(row) },
          {
            default: () => '分配权限',
            icon: () => h('span', { class: 'n-icon' }, [h(KeyOutline)]),
          },
        ))
      }
      if (authStore.hasPermission('role:delete')) {
        btns.push(h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' }))
      }
      return h(NSpace, null, { default: () => btns })
    },
  },
]

onMounted(loadData)
</script>

<template>
  <div class="roles-page">
    <!-- 搜索栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <n-input
          v-model:value="query.keyword"
          placeholder="搜索角色标识或名称"
          clearable
          style="width: 240px;"
          @keyup.enter="handleSearch"
        />
        <n-button @click="handleSearch">搜索</n-button>
        <n-button @click="handleReset">重置</n-button>
      </div>
      <n-button v-permission="'role:create'" type="primary" @click="openCreate">
        <template #icon><n-icon><KeyOutline /></n-icon></template>
        新建角色
      </n-button>
    </div>

    <!-- 数据表格 -->
    <n-card :bordered="false" class="table-card">
      <n-data-table
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :pagination="false"
        :scroll-x="900"
        size="small"
        striped
      />
      <div class="pagination">
        <n-pagination
          v-model:page="query.page_num"
          v-model:page-size="query.page_size"
          :item-count="total"
          :page-sizes="[20, 50]"
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
      :title="modalMode === 'create' ? '新建角色' : '编辑角色'"
      preset="card"
      style="width: 460px;"
      :mask-closable="false"
    >
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80">
        <n-form-item label="角色标识" path="name">
          <n-input
            v-model:value="form.name"
            placeholder="如 admin、operator"
            :disabled="modalMode === 'edit'"
          />
        </n-form-item>
        <n-form-item label="角色名称" path="label">
          <n-input v-model:value="form.label" placeholder="如 普通管理员" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="form.description" type="textarea" placeholder="角色说明（选填）" :rows="2" />
        </n-form-item>
        <n-form-item label="排序" path="sort">
          <n-input-number v-model:value="form.sort" :min="0" style="width: 100%;" />
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

    <!-- 权限分配抽屉 -->
    <n-drawer
      v-model:show="drawerVisible"
      :width="480"
      placement="right"
      :mask-closable="false"
    >
      <n-drawer-content :title="`分配权限：${drawerRole?.label ?? ''}`" closable>
        <n-spin :show="drawerLoading">
          <div v-if="permGroups.length === 0 && !drawerLoading" class="perm-empty">
            暂无权限数据，请先执行 seed.sql 写入初始权限
          </div>

          <div v-for="group in permGroups" :key="group.group" class="perm-group">
            <!-- 分组标题 + 全选 -->
            <div class="perm-group-header">
              <n-checkbox
                :checked="isGroupChecked(group)"
                :indeterminate="isGroupIndeterminate(group)"
                @update:checked="(v) => toggleGroup(group, v)"
              >
                <span class="perm-group-label">{{ groupLabel(group.group) }}</span>
              </n-checkbox>
            </div>

            <!-- 权限列表 -->
            <div class="perm-list">
              <n-checkbox
                v-for="perm in group.permissions"
                :key="perm.id"
                :checked="checkedPermIds.has(perm.id)"
                @update:checked="(v) => togglePerm(perm.id, v)"
              >
                <span class="perm-label">{{ perm.label }}</span>
                <span class="perm-name">{{ perm.name }}</span>
              </n-checkbox>
            </div>
          </div>
        </n-spin>

        <template #footer>
          <n-space justify="space-between" align="center" style="width: 100%;">
            <span class="checked-count">已选 {{ checkedCount }} 项权限</span>
            <n-space>
              <n-button @click="drawerVisible = false">取消</n-button>
              <n-button type="primary" :loading="drawerLoading" @click="handleSavePermissions">
                保存权限
              </n-button>
            </n-space>
          </n-space>
        </template>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<style scoped>
.roles-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 工具栏 */
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

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

/* 权限抽屉 */
.perm-empty {
  color: #909399;
  font-size: 13px;
  text-align: center;
  padding: 40px 0;
}

.perm-group {
  margin-bottom: 20px;
}

.perm-group-header {
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 10px;
}

.perm-group-label {
  font-weight: 600;
  font-size: 14px;
  color: #18181c;
}

.perm-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px 16px;
  padding-left: 4px;
}

.perm-label {
  font-size: 13px;
  color: #18181c;
}

.perm-name {
  font-size: 11px;
  color: #b0b0b0;
  margin-left: 4px;
}

.checked-count {
  font-size: 13px;
  color: #909399;
}
</style>
