<script setup lang="ts">
// system/MenusView.vue
// 菜单管理页（仅 super_admin）。
//
// 功能范围（保持克制，先把闭环做通）：
// - 展示完整菜单树（树形表格）
// - 新增/编辑：title、path、permission、super_admin_only、icon、sort、visible、parent_id
// - 删除：软删除
//
// 说明：
// - path 为空表示目录节点（不跳转，仅用于分组/展开）
// - permission 仅控制“侧边栏可见性裁剪”，不等同接口鉴权；接口安全仍由后端中间件保证
import { h, ref, reactive, onMounted, computed, watch } from 'vue'
import { useMessage, useDialog, NButton, NSpace, NSwitch, NTag } from 'naive-ui'
import { AddOutline } from '@vicons/ionicons5'
import { listMenus, createMenu, updateMenu, deleteMenu } from '@/api/menu'
import type { AdminMenuNode } from '@/types'

const message = useMessage()
const dialog = useDialog()

// ── 列表（树） ────────────────────────────────────────────
const loading = ref(false)
const treeData = ref<AdminMenuNode[]>([])
const keyword = ref('')
const selectedId = ref<number | null>(null)
const selectedNode = computed(() => {
  if (!selectedId.value) return null
  return flatten(treeData.value).find((n) => n.id === selectedId.value) ?? null
})

async function loadData() {
  loading.value = true
  try {
    const res = await listMenus()
    treeData.value = res.data.data
    // 数据刷新后尽量保持选中项（若被删除则回落到空）
    if (selectedId.value && !flatten(treeData.value).some((n) => n.id === selectedId.value)) {
      selectedId.value = null
    }
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function flatten(nodes: AdminMenuNode[]): AdminMenuNode[] {
  const out: AdminMenuNode[] = []
  const walk = (arr: AdminMenuNode[]) => {
    arr.forEach((n) => {
      out.push(n)
      if (n.children?.length) walk(n.children)
    })
  }
  walk(nodes)
  return out
}

function normalize(s: string) {
  return (s || '').trim().toLowerCase()
}

// 过滤后的树（左侧展示用）
const filteredTree = computed(() => {
  const kw = normalize(keyword.value)
  if (!kw) return treeData.value

  // 命中规则：title/path/permission 任意包含关键词
  const match = (n: AdminMenuNode) => {
    const t = normalize(n.title)
    const p = normalize(n.path ?? '')
    const perm = normalize(n.permission ?? '')
    return t.includes(kw) || p.includes(kw) || perm.includes(kw)
  }

  // 递归过滤：保留命中节点及其祖先链（通过保留其子树命中）
  const filterNodes = (nodes: AdminMenuNode[]): AdminMenuNode[] => {
    const out: AdminMenuNode[] = []
    nodes.forEach((n) => {
      const children = n.children?.length ? filterNodes(n.children) : []
      if (match(n) || children.length) {
        out.push({ ...n, children })
      }
    })
    return out
  }

  return filterNodes(treeData.value)
})

// 左侧树选项（Naive UI NTree）
type MenuTreeOption = {
  key: number
  label: string
  children?: MenuTreeOption[]
}

const treeOptions = computed<MenuTreeOption[]>(() => {
  const mapNodes = (nodes: AdminMenuNode[]): MenuTreeOption[] =>
    nodes.map((n) => ({
      key: n.id,
      label: n.title,
      children: n.children?.length ? mapNodes(n.children) : undefined,
    }))
  return mapNodes(filteredTree.value)
})

// nodeByID 用于渲染树节点 label 时回查原始数据。
//
// 说明：
// - naive-ui 的 TreeOption 可能会被内部归一化处理，render-label 收到的 option 不保证保留自定义字段。
// - 因此不要把 path/permission 等业务字段塞进 option 里依赖它们，统一通过 key 回查。
const nodeByID = computed(() => {
  const m = new Map<number, AdminMenuNode>()
  flatten(treeData.value).forEach((n) => m.set(n.id, n))
  return m
})

function renderTreeLabel(info: unknown) {
  // naive-ui 可能传入 { option } 结构，这里做兼容，避免渲染期直接报错导致整页空白。
  const option = (info && typeof info === 'object' && 'option' in (info as any))
    ? (info as any).option as MenuTreeOption
    : info as MenuTreeOption

  const node = nodeByID.value.get(Number(option?.key))
  if (!node) {
    return h('span', null, option?.label ?? '')
  }

  const tags = []
  if (!node.path) tags.push(h(NTag, { size: 'small' }, { default: () => '目录' }))
  if (node.super_admin_only) tags.push(h(NTag, { size: 'small', type: 'warning' }, { default: () => '仅超管' }))
  if (node.visible === 0) tags.push(h(NTag, { size: 'small', type: 'default' }, { default: () => '隐藏' }))

  const sub = []
  if (node.path) sub.push(node.path)
  if (node.permission) sub.push(node.permission)

  return h('div', { class: 'tree-label' }, [
    h('div', { class: 'tree-label-main' }, [
      h('span', { class: 'tree-label-title' }, node.title),
      ...tags,
    ]),
    h('div', { class: 'tree-label-sub' }, sub.length ? sub.join(' · ') : ''),
  ])
}

// 父级菜单选项：只要是目录节点（path 为空）就允许作为 parent
// 同时排除自身及其子孙，避免成环。
function collectDescendantIds(rootId: number): Set<number> {
  const ids = new Set<number>()
  const nodeMap = new Map<number, AdminMenuNode>()
  flatten(treeData.value).forEach((n) => nodeMap.set(n.id, n))

  // 构建 parent -> children ids 映射，便于遍历子孙
  const childrenMap = new Map<number, number[]>()
  flatten(treeData.value).forEach((n) => {
    const arr = childrenMap.get(n.parent_id) ?? []
    arr.push(n.id)
    childrenMap.set(n.parent_id, arr)
  })

  const stack: number[] = [rootId]
  while (stack.length) {
    const cur = stack.pop()!
    const kids = childrenMap.get(cur) ?? []
    kids.forEach((kid) => {
      if (!ids.has(kid)) {
        ids.add(kid)
        stack.push(kid)
      }
    })
  }

  // nodeMap 未使用但保留：后续如要展示层级标题可用
  void nodeMap
  return ids
}

const parentOptions = computed(() => {
  const options = [{ label: '顶级菜单', value: 0 }]
  const forbid = selectedId.value ? collectDescendantIds(selectedId.value) : new Set<number>()
  if (selectedId.value) forbid.add(selectedId.value)

  flatten(treeData.value).forEach((n) => {
    if (!n.path) {
      if (forbid.has(n.id)) return
      options.push({ label: n.title, value: n.id })
    }
  })
  return options
})

// ── 右侧编辑（新建/编辑）───────────────────────────────────
const formRef = ref()
const saving = ref(false)
const mode = ref<'create' | 'edit' | 'empty'>('empty')

const form = reactive({
  parent_id: 0,
  title: '',
  path: '',
  permission: '',
  super_admin_only: 0 as 0 | 1,
  icon: '',
  sort: 0,
  visible: 1 as 0 | 1,
})

const rules = {
  title: [{ required: true, message: '请输入菜单标题', trigger: 'blur' }],
  visible: [{ required: true, type: 'number', trigger: 'change' }],
  super_admin_only: [{ required: true, type: 'number', trigger: 'change' }],
}

function startCreate(parentID: number) {
  mode.value = 'create'
  Object.assign(form, {
    parent_id: parentID,
    title: '',
    path: '',
    permission: '',
    super_admin_only: 0,
    icon: '',
    sort: 0,
    visible: 1,
  })
}

function startEdit(row: AdminMenuNode) {
  mode.value = 'edit'
  selectedId.value = row.id
  Object.assign(form, {
    parent_id: row.parent_id,
    title: row.title,
    path: row.path ?? '',
    permission: row.permission ?? '',
    super_admin_only: (row.super_admin_only ? 1 : 0) as 0 | 1,
    icon: row.icon ?? '',
    sort: row.sort ?? 0,
    visible: (row.visible ? 1 : 0) as 0 | 1,
  })
}

async function handleSubmit() {
  await formRef.value?.validate()
  saving.value = true
  try {
    const payload = {
      parent_id: form.parent_id,
      title: form.title,
      // 约定：空字符串 -> 后端视为目录节点（path=NULL）
      path: form.path,
      permission: form.permission,
      super_admin_only: form.super_admin_only,
      icon: form.icon,
      sort: form.sort,
      visible: form.visible,
    }
    if (mode.value === 'create') {
      const res = await createMenu(payload)
      message.success('创建成功')
      // 创建后选中该节点并进入编辑模式
      selectedId.value = (res.data.data as any).id ?? null
      mode.value = 'edit'
    } else {
      await updateMenu(selectedId.value!, payload)
      message.success('更新成功')
    }
    await loadData()
  } catch (err: unknown) {
    message.error(err instanceof Error ? err.message : '操作失败')
  } finally {
    saving.value = false
  }
}

// ── 删除 ────────────────────────────────────────────────
function handleDelete(row: AdminMenuNode) {
  dialog.warning({
    title: '确认删除',
    content: `确认删除菜单「${row.title}」？（软删除，可通过数据库恢复）`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteMenu(row.id)
        message.success('已删除')
        if (selectedId.value === row.id) {
          selectedId.value = null
          mode.value = 'empty'
        }
        await loadData()
      } catch (err: unknown) {
        message.error(err instanceof Error ? err.message : '删除失败')
      }
    },
  })
}

function handleCreateTop() {
  selectedId.value = null
  startCreate(0)
}

function handleCreateChild() {
  if (!selectedNode.value) {
    message.warning('请先选择一个菜单作为父级')
    return
  }
  if (selectedNode.value.path) {
    message.warning('只能在目录节点下新建子菜单')
    return
  }
  startCreate(selectedNode.value.id)
}

function handleDeleteSelected() {
  if (!selectedNode.value) {
    message.warning('请先选择要删除的菜单')
    return
  }
  handleDelete(selectedNode.value)
}

function onSelectKey(key: number) {
  // 点击树节点即进入编辑
  const node = flatten(treeData.value).find((n) => n.id === key)
  if (node) startEdit(node)
}

function handleTreeSelectedKeys(keys: Array<string | number>) {
  const first = keys && keys.length > 0 ? keys[0] : null
  if (first === null || first === undefined) return
  onSelectKey(Number(first))
}

// 选中节点变化时，自动把详情带到右侧（实现“点击即编辑”）
watch(selectedNode, (n) => {
  if (!n) {
    if (mode.value === 'edit') mode.value = 'empty'
    return
  }
  // 如果当前处于 create，不强行覆盖用户正在输入的内容
  if (mode.value !== 'create') {
    startEdit(n)
  }
})

onMounted(() => {
  loadData()
  mode.value = 'empty'
})
</script>

<template>
  <div class="page-container">
    <div class="toolbar">
      <div class="toolbar-left">
        <div class="toolbar-tip">
          <span class="tip-title">提示：</span>
          <span class="tip-text">点击左侧菜单行即可编辑；path 留空表示目录；permission 只控制侧边栏可见性。</span>
        </div>
      </div>
      <n-space>
        <n-button @click="loadData">刷新</n-button>
        <n-button type="primary" @click="handleCreateTop">
          <template #icon><n-icon><AddOutline /></n-icon></template>
          新建顶级
        </n-button>
        <n-button :disabled="!selectedNode || !!selectedNode.path" @click="handleCreateChild">
          新建子级
        </n-button>
        <n-button type="error" :disabled="!selectedNode" @click="handleDeleteSelected">
          删除
        </n-button>
      </n-space>
    </div>

    <div class="split">
      <!-- 左：菜单树 -->
      <n-card :bordered="false" class="panel left">
        <div class="left-head">
          <n-input
            v-model:value="keyword"
            clearable
            placeholder="搜索标题 / 路径 / 权限"
            style="width: 320px;"
          />
          <div class="left-meta">
            <span class="meta-text">共 {{ flatten(treeData).length }} 项</span>
          </div>
        </div>

        <n-spin :show="loading">
          <n-tree
            block-line
            selectable
            :data="treeOptions"
            :default-expand-all="true"
            :selected-keys="selectedId ? [selectedId] : []"
            :render-label="renderTreeLabel"
            @update:selected-keys="handleTreeSelectedKeys"
          />
        </n-spin>
      </n-card>

      <!-- 右：详情编辑 -->
      <n-card :bordered="false" class="panel right">
        <template #header>
          <div class="right-title">
            <span v-if="mode === 'edit'">编辑菜单：{{ selectedNode?.title }}</span>
            <span v-else-if="mode === 'create'">新建菜单</span>
            <span v-else>菜单详情</span>
          </div>
        </template>

        <div v-if="mode === 'empty'" class="empty-tip">
          请选择左侧菜单进行编辑，或点击“新建顶级/新建子级”创建菜单。
        </div>

        <n-form
          v-else
          ref="formRef"
          :model="form"
          :rules="rules"
          label-placement="left"
          label-width="110"
        >
          <n-form-item label="父级菜单" path="parent_id">
            <n-select v-model:value="form.parent_id" :options="parentOptions" placeholder="选择父级（目录）" />
          </n-form-item>
          <n-form-item label="菜单标题" path="title">
            <n-input v-model:value="form.title" placeholder="如 角色管理" />
          </n-form-item>
          <n-form-item label="路由路径" path="path">
            <n-input v-model:value="form.path" placeholder="如 /system/roles（留空表示目录）" />
          </n-form-item>
          <n-form-item label="可见权限" path="permission">
            <n-input v-model:value="form.permission" placeholder="如 role:list（留空表示无需权限）" />
          </n-form-item>
          <n-form-item label="仅超管可见" path="super_admin_only">
            <n-switch
              :value="form.super_admin_only === 1"
              @update:value="(v) => (form.super_admin_only = v ? 1 : 0)"
            />
          </n-form-item>
          <n-form-item label="图标标识" path="icon">
            <n-input v-model:value="form.icon" placeholder="如 dashboard/system（可选）" />
          </n-form-item>
          <n-form-item label="排序" path="sort">
            <n-input-number v-model:value="form.sort" :min="0" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="是否显示" path="visible">
            <n-switch
              :value="form.visible === 1"
              @update:value="(v) => (form.visible = v ? 1 : 0)"
            />
          </n-form-item>

          <div class="right-actions">
            <n-space justify="end">
              <n-button @click="mode = 'empty'">取消</n-button>
              <n-button type="primary" :loading="saving" @click="handleSubmit">
                {{ mode === 'create' ? '创建' : '保存' }}
              </n-button>
            </n-space>
          </div>
        </n-form>
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  padding: 16px 20px;
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.toolbar-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #666;
}

.tip-title {
  font-weight: 600;
}

.split {
  display: flex;
  gap: 16px;
  min-height: 520px;
}

.panel {
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.left {
  flex: 3;
  min-width: 560px;
}

.right {
  flex: 2;
  min-width: 420px;
}

.left-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.left-meta .meta-text {
  font-size: 12px;
  color: #909399;
}

.tree-label {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 2px 0;
}

.tree-label-main {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tree-label-title {
  color: #18181c;
}

.tree-label-sub {
  font-size: 12px;
  color: #909399;
}

.right-title {
  font-weight: 600;
  color: #18181c;
}

.empty-tip {
  font-size: 13px;
  color: #909399;
  padding: 24px 0;
}

.right-actions {
  margin-top: 8px;
}

@media (max-width: 1100px) {
  .split {
    flex-direction: column;
  }
  .left,
  .right {
    min-width: auto;
  }
}
</style>
