<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import {
  type FormInstance,
  type FormRules,
} from 'element-plus'
import { ArrowRight, CaretBottom } from '@element-plus/icons-vue'

import type { AdminPermissionGroup } from '../../../api/admin-role'
import type { PermissionTreeNode, RoleEditorState } from '../types'

interface PermissionCatalogNode {
  id: string
  label: string
  metaLabel: string
  pathHint?: string
  permissionCodes?: string[]
  children?: PermissionCatalogNode[]
}

const permissionCatalog: PermissionCatalogNode[] = [
  {
    id: 'menu:dashboard',
    label: '控制台',
    metaLabel: '一级菜单',
    pathHint: '/dashboard',
    permissionCodes: ['page.dashboard', 'dashboard:*', 'dashboard:view'],
  },
  {
    id: 'menu:system-settings',
    label: '系统设置',
    metaLabel: '一级菜单',
    pathHint: '/system',
    children: [
      {
        id: 'page:system-config',
        label: '系统配置',
        metaLabel: '系统设置子页面',
        pathHint: '/system/settings',
        permissionCodes: ['page.system-settings.config', 'system-config:*', 'system-config:view', 'system-config:update'],
      },
      {
        id: 'page:admin-settings',
        label: '管理员设置',
        metaLabel: '系统设置子页面',
        pathHint: '/system/admin-users',
        children: [
          {
            id: 'tab:admin-users',
            label: '管理员账号',
            metaLabel: '管理员设置 Tab',
            pathHint: 'Tab / 管理员账号',
            permissionCodes: [
              'page.system-settings.admin-users',
              'admin-user:*',
              'admin-user:view',
              'admin-user:create',
              'admin-user:update',
              'admin-user:password-reset',
            ],
          },
          {
            id: 'tab:admin-roles',
            label: '管理员权限',
            metaLabel: '管理员设置 Tab',
            pathHint: 'Tab / 管理员权限',
            permissionCodes: [
              'page.system-settings.admin-roles',
              'admin-role:*',
              'admin-role:view',
              'admin-role:create',
              'admin-role:update',
            ],
          },
          {
            id: 'tab:admin-sessions',
            label: '管理员会话',
            metaLabel: '管理员设置 Tab',
            pathHint: 'Tab / 管理员会话',
            permissionCodes: [
              'page.system-settings.admin-sessions',
              'admin-session:*',
              'admin-session:view',
              'admin-session:revoke',
            ],
          },
        ],
      },
    ],
  },
]

const props = defineProps<{
  visible: boolean
  title: string
  isCreateMode: boolean
  isBuiltInRole: boolean
  form: RoleEditorState
  rules: FormRules<RoleEditorState>
  permissionGroups: AdminPermissionGroup[]
  submitting: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: []
  closed: []
}>()

const formRef = ref<FormInstance>()
const filterText = ref('')
const expandedNodeIds = ref<string[]>([])

interface PermissionTreeRow {
  node: PermissionTreeNode
  depth: number
  expanded: boolean
  isBranch: boolean
  checked: boolean
  indeterminate: boolean
  totalLeafCount: number
  checkedLeafCount: number
}

const flatPermissions = computed(() =>
  props.permissionGroups
    .flatMap((group) => group.permissions)
    .sort((left, right) => left.code.localeCompare(right.code, 'zh-CN')),
)

const permissionCodeSet = computed(() => new Set(flatPermissions.value.map((permission) => permission.code)))
const permissionCount = computed(() => props.form.permission_codes.length)
const permissionTreeData = computed<PermissionTreeNode[]>(() =>
  buildPermissionTree(flatPermissions.value, props.permissionGroups, props.isBuiltInRole),
)
const normalizedFilter = computed(() => filterText.value.trim().toLowerCase())
const expandedNodeIdSet = computed(() => new Set(expandedNodeIds.value))
const visibleRows = computed(() => buildVisibleRows(permissionTreeData.value))
const totalPermissionCount = computed(() => flatPermissions.value.length)
const unmatchedPermissionCount = computed(() =>
  permissionTreeData.value.find((node) => node.id === 'unmatched-permissions')?.count ?? 0,
)
const visiblePermissionCount = computed(() =>
  visibleRows.value.filter((row) => row.node.type === 'permission').length,
)
const dialogDescription = computed(() =>
  props.isCreateMode
    ? '定义管理组基础信息，并按系统真实菜单目录分配权限。'
    : '调整管理组资料、状态与目录化权限分配方案。',
)

watch(
  () => props.visible,
  (value) => {
    if (value) {
      filterText.value = ''
      void nextTick(() => {
        formRef.value?.clearValidate()
        syncSelectedPermissions()
        syncExpandedState(true)
      })
    }
  },
)

watch(
  () => props.permissionGroups,
  () => {
    if (!props.visible) {
      return
    }
    void nextTick(() => {
      syncSelectedPermissions()
      syncExpandedState(false)
    })
  },
  { deep: true },
)

function uniqueSortedStrings(values: string[]) {
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean))).sort()
}

function syncSelectedPermissions() {
  props.form.permission_codes = uniqueSortedStrings(
    props.form.permission_codes.filter((code) => permissionCodeSet.value.has(code)),
  )
}

function syncExpandedState(reset: boolean) {
  const branchIds = collectBranchIds(permissionTreeData.value)
  const nextExpanded = reset
    ? getInitialExpandedIds(permissionTreeData.value)
    : expandedNodeIds.value.filter((id) => branchIds.has(id))
  expandedNodeIds.value = uniqueSortedStrings(nextExpanded)
}

function handleCheckAll() {
  if (props.isBuiltInRole) {
    return
  }
  const keys = flatPermissions.value.map((permission) => permission.code)
  props.form.permission_codes = uniqueSortedStrings(keys)
}

function handleClear() {
  if (props.isBuiltInRole) {
    return
  }
  props.form.permission_codes = []
}

function toggleExpanded(nodeId: string) {
  const next = new Set(expandedNodeIds.value)
  if (next.has(nodeId)) {
    next.delete(nodeId)
  } else {
    next.add(nodeId)
  }
  expandedNodeIds.value = Array.from(next)
}

function toggleNodeSelection(node: PermissionTreeNode, checked: boolean) {
  if (props.isBuiltInRole) {
    return
  }

  const next = new Set(props.form.permission_codes)
  for (const code of collectLeafCodes(node)) {
    if (checked) {
      next.add(code)
    } else {
      next.delete(code)
    }
  }
  props.form.permission_codes = uniqueSortedStrings(Array.from(next))
}

async function handleSubmit() {
  if (!formRef.value) {
    return
  }
  await formRef.value.validate()
  emit('submit')
}

function buildPermissionTree(
  permissions: Array<AdminPermissionGroup['permissions'][number]>,
  _groups: AdminPermissionGroup[],
  disabled: boolean,
) {
  const permissionMap = new Map(permissions.map((permission) => [permission.code, permission]))
  const assignedCodes = new Set<string>()
  const roots = permissionCatalog
    .map((catalogNode, index) => buildCatalogTreeNode(catalogNode, permissionMap, assignedCodes, disabled, index))
    .filter(Boolean) as PermissionTreeNode[]

  const unmatchedPermissions = permissions.filter((permission) => !assignedCodes.has(permission.code))
  if (unmatchedPermissions.length > 0) {
    roots.push(buildUnmatchedTreeNode(unmatchedPermissions, disabled, roots.length))
  }

  for (const root of roots) {
    finalizeTree(root)
  }

  return roots
}

function buildCatalogTreeNode(
  catalogNode: PermissionCatalogNode,
  permissionMap: Map<string, AdminPermissionGroup['permissions'][number]>,
  assignedCodes: Set<string>,
  disabled: boolean,
  index: number,
): PermissionTreeNode | null {
  const children: PermissionTreeNode[] = []

  for (const [permissionIndex, code] of (catalogNode.permissionCodes ?? []).entries()) {
    const permission = permissionMap.get(code)
    if (!permission) {
      continue
    }
    assignedCodes.add(code)
    children.push(createPermissionLeafNode(permission, disabled, permissionIndex))
  }

  for (const [childIndex, childNode] of (catalogNode.children ?? []).entries()) {
    const builtChild = buildCatalogTreeNode(childNode, permissionMap, assignedCodes, disabled, childIndex)
    if (builtChild) {
      children.push(builtChild)
    }
  }

  if (children.length === 0) {
    return null
  }

  return {
    id: catalogNode.id,
    label: catalogNode.label,
    type: 'root',
    disabled,
    meta_label: catalogNode.metaLabel,
    path_hint: catalogNode.pathHint,
    keywords: [catalogNode.label, catalogNode.metaLabel, catalogNode.pathHint ?? ''],
    sort_order: index,
    children,
  }
}

function createPermissionLeafNode(
  permission: AdminPermissionGroup['permissions'][number],
  disabled: boolean,
  index: number,
): PermissionTreeNode {
  return {
    id: permission.code,
    label: permission.name,
    type: 'permission',
    code: permission.code,
    description: permission.description,
    disabled,
    meta_label: permission.code.startsWith('page.') ? '页面入口权限' : '资源操作权限',
    keywords: [permission.code, permission.name, permission.group_name],
    sort_order: index,
  }
}

function buildUnmatchedTreeNode(
  permissions: Array<AdminPermissionGroup['permissions'][number]>,
  disabled: boolean,
  index: number,
): PermissionTreeNode {
  return {
    id: 'unmatched-permissions',
    label: '未归类权限',
    type: 'root',
    disabled,
    meta_label: '需要补充目录映射',
    path_hint: '新增权限若出现在这里，说明还没挂到正确业务目录',
    keywords: ['未归类', '权限映射', '目录映射'],
    sort_order: index,
    children: permissions.map((permission, permissionIndex) =>
      createPermissionLeafNode(permission, disabled, permissionIndex),
    ),
  }
}

function finalizeTree(node: PermissionTreeNode) {
  if (node.children?.length) {
    for (const child of node.children) {
      finalizeTree(child)
    }
    node.children.sort(compareTreeNodes)
    node.count = node.children.reduce((sum, child) => sum + (child.type === 'permission' ? 1 : child.count ?? 0), 0)
  }
}

function compareTreeNodes(left: PermissionTreeNode, right: PermissionTreeNode) {
  const orderDelta = (left.sort_order ?? 999) - (right.sort_order ?? 999)
  if (orderDelta !== 0) {
    return orderDelta
  }
  return left.label.localeCompare(right.label, 'zh-CN')
}

function isPermissionNode(node: PermissionTreeNode) {
  return node.type === 'permission'
}

function isBranchNode(node: PermissionTreeNode) {
  return (node.children?.length ?? 0) > 0
}

function getPrimaryLabel(node: PermissionTreeNode) {
  if (isPermissionNode(node)) {
    return node.meta_label || node.label
  }
  return node.label
}

function getSecondaryLabel(node: PermissionTreeNode) {
  if (isPermissionNode(node)) {
    return node.meta_label && node.meta_label !== node.label ? node.label : ''
  }
  return node.meta_label || ''
}

function collectLeafCodes(node: PermissionTreeNode): string[] {
  if (isPermissionNode(node) && node.code) {
    return [node.code]
  }
  return (node.children ?? []).flatMap((child) => collectLeafCodes(child))
}

function nodeMatchesFilter(node: PermissionTreeNode) {
  if (!normalizedFilter.value) {
    return true
  }
  return [node.label, node.meta_label, node.code, node.description, node.path_hint, ...(node.keywords ?? [])]
    .filter(Boolean)
    .some((item) => String(item).toLowerCase().includes(normalizedFilter.value))
}

function hasVisibleDescendant(node: PermissionTreeNode): boolean {
  return (node.children ?? []).some((child) => shouldShowNode(child))
}

function shouldShowNode(node: PermissionTreeNode) {
  if (!normalizedFilter.value) {
    return true
  }
  return nodeMatchesFilter(node) || hasVisibleDescendant(node)
}

function buildVisibleRows(nodes: PermissionTreeNode[], depth = 0): PermissionTreeRow[] {
  const rows: PermissionTreeRow[] = []
  const selected = new Set(props.form.permission_codes)
  const forceExpand = Boolean(normalizedFilter.value)

  for (const node of nodes) {
    if (!shouldShowNode(node)) {
      continue
    }

    const leafCodes = collectLeafCodes(node)
    const checkedLeafCount = leafCodes.filter((code) => selected.has(code)).length
    const totalLeafCount = leafCodes.length
    const isBranch = isBranchNode(node)
    const expanded = forceExpand || expandedNodeIdSet.value.has(node.id)

    rows.push({
      node,
      depth,
      expanded,
      isBranch,
      checked: totalLeafCount > 0 && checkedLeafCount === totalLeafCount,
      indeterminate: checkedLeafCount > 0 && checkedLeafCount < totalLeafCount,
      totalLeafCount,
      checkedLeafCount,
    })

    if (isBranch && expanded) {
      rows.push(...buildVisibleRows(node.children ?? [], depth + 1))
    }
  }

  return rows
}

function collectBranchIds(nodes: PermissionTreeNode[]) {
  const branchIds = new Set<string>()
  for (const node of nodes) {
    if (isBranchNode(node)) {
      branchIds.add(node.id)
      for (const childId of collectBranchIds(node.children ?? [])) {
        branchIds.add(childId)
      }
    }
  }
  return branchIds
}

function getInitialExpandedIds(nodes: PermissionTreeNode[]) {
  const result: string[] = []
  for (const node of nodes) {
    if (!isBranchNode(node)) {
      continue
    }
    result.push(node.id)
    for (const child of node.children ?? []) {
      if (isBranchNode(child)) {
        result.push(child.id)
      }
    }
  }
  return result
}
</script>

<template>
  <el-dialog
    class="role-editor-dialog"
    :model-value="props.visible"
    :title="props.title"
    width="1120px"
    top="4vh"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
    @closed="emit('closed')"
  >
    <el-form ref="formRef" :model="props.form" :rules="props.rules" label-position="top" class="role-editor-dialog__form">
      <div class="role-editor-dialog__layout">
        <section class="role-editor-dialog__sidebar">
          <div class="role-editor-dialog__panel role-editor-dialog__panel--sticky">
            <div class="role-editor-dialog__section-head">
              <div>
                <h3>基础信息</h3>
                <p>{{ dialogDescription }}</p>
              </div>
              <el-tag size="small" effect="plain">{{ props.isCreateMode ? '新建模式' : '编辑模式' }}</el-tag>
            </div>

            <div class="role-editor-dialog__field-list">
              <el-form-item label="管理组编码" prop="code">
                <el-input
                  v-model="props.form.code"
                  :disabled="!props.isCreateMode"
                  placeholder="请输入唯一编码，例如 ops_manager"
                />
              </el-form-item>
              <el-form-item label="管理组名称" prop="name">
                <el-input v-model="props.form.name" placeholder="请输入管理组名称" />
              </el-form-item>
              <el-form-item label="说明" prop="description">
                <el-input
                  v-model="props.form.description"
                  type="textarea"
                  :rows="4"
                  placeholder="请输入管理组说明，可留空"
                />
              </el-form-item>
              <el-form-item label="状态" prop="status">
                <el-radio-group v-model="props.form.status" class="role-editor-dialog__status-group">
                  <el-radio value="active" :disabled="props.isBuiltInRole">启用</el-radio>
                  <el-radio value="disabled" :disabled="props.isBuiltInRole">停用</el-radio>
                </el-radio-group>
              </el-form-item>
            </div>

            <div class="role-editor-dialog__summary-card">
              <div class="role-editor-dialog__summary-item">
                <span>已选权限</span>
                <strong>{{ permissionCount }}</strong>
              </div>
              <div class="role-editor-dialog__summary-item">
                <span>可分配权限</span>
                <strong>{{ totalPermissionCount }}</strong>
              </div>
              <div class="role-editor-dialog__summary-item">
                <span>目录覆盖</span>
                <strong>{{ totalPermissionCount === 0 ? '0%' : `${Math.round((permissionCount / totalPermissionCount) * 100)}%` }}</strong>
              </div>
            </div>

            <el-alert
              v-if="props.isBuiltInRole"
              type="warning"
              :closable="false"
              title="内置超级管理员角色不可修改状态和权限分配。"
            />
          </div>
        </section>

        <section class="role-editor-dialog__workspace">
          <div class="role-editor-dialog__panel">
            <div class="role-editor-dialog__section-head role-editor-dialog__section-head--workspace">
              <div>
                <h3>权限工作台</h3>
                <p>按照真实菜单目录管理页面入口权限和资源操作权限，避免后续权限增长后归属混乱。</p>
              </div>
              <div class="role-editor-dialog__section-tags">
                <el-tag type="primary" effect="light">目录化分配</el-tag>
                <el-tag v-if="unmatchedPermissionCount > 0" type="danger" effect="light">
                  未归类 {{ unmatchedPermissionCount }} 项
                </el-tag>
              </div>
            </div>

            <div class="role-editor-dialog__overview">
              <div class="role-editor-dialog__overview-card">
                <span>已选权限</span>
                <strong>{{ permissionCount }}</strong>
                <small>当前角色已授权的权限总数</small>
              </div>
              <div class="role-editor-dialog__overview-card">
                <span>目录节点</span>
                <strong>{{ permissionTreeData.length }}</strong>
                <small>按控制台与系统设置等真实目录组织</small>
              </div>
              <div class="role-editor-dialog__overview-card">
                <span>当前筛选结果</span>
                <strong>{{ visiblePermissionCount }}</strong>
                <small>会随搜索即时刷新并自动展开命中目录</small>
              </div>
            </div>

            <el-alert
              v-if="unmatchedPermissionCount > 0"
              type="warning"
              :closable="false"
              title="发现未归类权限"
              description="这些权限尚未挂到明确业务目录，建议在目录映射中补齐，避免后续授权时找不到归属。"
            />

            <div class="role-editor-dialog__toolbar">
              <el-input
                v-model="filterText"
                clearable
                placeholder="搜索目录、权限名称、权限码"
                class="role-editor-dialog__search"
              />
              <div class="role-editor-dialog__toolbar-actions">
                <el-button link type="primary" :disabled="props.isBuiltInRole" @click="handleCheckAll">全选全部权限</el-button>
                <el-button link :disabled="props.isBuiltInRole || permissionCount === 0" @click="handleClear">清空选择</el-button>
              </div>
            </div>

            <div class="role-editor-dialog__permission-browser">
              <div class="role-editor-dialog__browser-head">
                <span>目录 / 权限项</span>
                <span>权限码 / 统计</span>
              </div>

              <div class="role-editor-dialog__permission-tree-wrap">
                <div v-if="visibleRows.length === 0" class="role-editor-dialog__permission-empty">
                  未找到相关权限节点
                </div>
                <div v-else class="role-editor-dialog__permission-tree">
                  <div
                    v-for="row in visibleRows"
                    :key="row.node.id"
                    class="role-editor-dialog__tree-row"
                    :class="{
                      'role-editor-dialog__tree-row--branch': row.isBranch,
                      'role-editor-dialog__tree-row--leaf': !row.isBranch,
                    }"
                    :style="{ paddingLeft: `${row.depth * 20}px` }"
                  >
                    <button
                      v-if="row.isBranch"
                      type="button"
                      class="role-editor-dialog__tree-toggle"
                      :aria-label="row.expanded ? '收起节点' : '展开节点'"
                      @click="toggleExpanded(row.node.id)"
                    >
                      <el-icon>
                        <CaretBottom v-if="row.expanded" />
                        <ArrowRight v-else />
                      </el-icon>
                    </button>
                    <span v-else class="role-editor-dialog__tree-toggle-placeholder" />

                    <el-checkbox
                      :model-value="row.checked"
                      :indeterminate="row.indeterminate"
                      :disabled="props.isBuiltInRole || row.node.disabled"
                      @update:model-value="toggleNodeSelection(row.node, Boolean($event))"
                    />

                    <div
                      class="role-editor-dialog__permission-node"
                      :class="{
                        'role-editor-dialog__permission-node--branch': row.isBranch,
                        'role-editor-dialog__permission-node--leaf': !row.isBranch,
                      }"
                      :title="[row.node.label, row.node.meta_label, row.node.code, row.node.description, row.node.path_hint].filter(Boolean).join(' / ')"
                    >
                      <div class="role-editor-dialog__permission-main">
                        <div class="role-editor-dialog__permission-title-row">
                          <span class="role-editor-dialog__permission-label">{{ getPrimaryLabel(row.node) }}</span>
                          <span v-if="!isPermissionNode(row.node) && row.node.path_hint" class="role-editor-dialog__permission-path-tag">
                            {{ row.node.path_hint }}
                          </span>
                        </div>
                        <span v-if="getSecondaryLabel(row.node)" class="role-editor-dialog__permission-meta">{{ getSecondaryLabel(row.node) }}</span>
                        <span v-if="row.node.type === 'permission' && row.node.description" class="role-editor-dialog__permission-description">
                          {{ row.node.description }}
                        </span>
                      </div>
                      <span v-if="row.node.type === 'permission' && row.node.code" class="role-editor-dialog__permission-code">
                        {{ row.node.code }}
                      </span>
                      <el-tag v-else size="small" effect="plain">
                        {{ row.checkedLeafCount }}/{{ row.totalLeafCount }}
                      </el-tag>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </el-form>

    <template #footer>
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="props.submitting" @click="handleSubmit">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.role-editor-dialog :deep(.el-dialog__body) {
  padding-top: 18px;
}

.role-editor-dialog__form,
.role-editor-dialog__workspace,
.role-editor-dialog__sidebar,
.role-editor-dialog__permission-browser {
  display: flex;
  flex-direction: column;
}

.role-editor-dialog__layout {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 20px;
  align-items: start;
}

.role-editor-dialog__panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 18px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 18px;
  background:
    linear-gradient(180deg, var(--el-bg-color-overlay) 0%, var(--el-fill-color-extra-light) 100%);
  box-shadow: 0 10px 24px rgb(0 0 0 / 0.04);
}

.role-editor-dialog__panel--sticky {
  position: sticky;
  top: 0;
}

.role-editor-dialog__section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.role-editor-dialog__section-head h3 {
  margin: 0;
  font-size: 18px;
  line-height: 1.3;
  color: var(--el-text-color-primary);
}

.role-editor-dialog__section-head p {
  margin: 6px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.role-editor-dialog__section-head--workspace {
  align-items: center;
}

.role-editor-dialog__section-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.role-editor-dialog__field-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.role-editor-dialog__status-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
}

.role-editor-dialog__summary-card {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.role-editor-dialog__summary-item,
.role-editor-dialog__overview-card {
  padding: 14px 12px;
  border-radius: 14px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-extra-light);
}

.role-editor-dialog__summary-item span,
.role-editor-dialog__overview-card span {
  display: block;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.role-editor-dialog__summary-item strong,
.role-editor-dialog__overview-card strong {
  display: block;
  margin-top: 8px;
  font-size: 24px;
  line-height: 1;
  color: var(--el-text-color-primary);
}

.role-editor-dialog__overview-card small {
  display: block;
  margin-top: 8px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

.role-editor-dialog__overview {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.role-editor-dialog__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.role-editor-dialog__search {
  max-width: 360px;
}

.role-editor-dialog__toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
}

.role-editor-dialog__browser-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 4px;
  font-size: 12px;
  letter-spacing: 0.02em;
  color: var(--el-text-color-secondary);
}

.role-editor-dialog__permission-tree-wrap {
  border: 1px solid var(--el-border-color-light);
  border-radius: 16px;
  background: var(--el-bg-color);
  padding: 10px;
  max-height: 520px;
  overflow: auto;
}

.role-editor-dialog__permission-tree {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.role-editor-dialog__permission-empty {
  padding: 24px 12px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.role-editor-dialog__tree-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  min-height: 52px;
  padding-top: 6px;
  padding-bottom: 6px;
  padding-right: 8px;
  border-radius: 10px;
  transition: background-color 0.2s ease;
}

.role-editor-dialog__tree-row:hover {
  background: var(--el-fill-color-extra-light);
}

.role-editor-dialog__tree-toggle,
.role-editor-dialog__tree-toggle-placeholder {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  margin-top: 2px;
}

.role-editor-dialog__tree-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  padding: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--el-text-color-secondary);
  cursor: pointer;
}

.role-editor-dialog__tree-toggle:hover {
  background: var(--el-fill-color-light);
  color: var(--el-text-color-primary);
}

.role-editor-dialog__tree-toggle:focus-visible {
  outline: 2px solid var(--el-color-primary-light-5);
  outline-offset: 1px;
}

.role-editor-dialog__permission-node {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.role-editor-dialog__permission-node--branch .role-editor-dialog__permission-label {
  font-weight: 600;
}

.role-editor-dialog__permission-main {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.role-editor-dialog__permission-title-row {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-editor-dialog__permission-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.role-editor-dialog__permission-meta,
.role-editor-dialog__permission-description {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.4;
}

.role-editor-dialog__permission-path-tag,
.role-editor-dialog__permission-code {
  flex-shrink: 0;
  font-size: 12px;
  line-height: 1.4;
  font-family: ui-monospace, SFMono-Regular, SFMono-Regular, Consolas, 'Liberation Mono', Menlo, monospace;
}

.role-editor-dialog__permission-path-tag {
  max-width: 260px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--el-fill-color-extra-light);
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.role-editor-dialog__permission-code {
  color: var(--el-text-color-secondary);
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-top: 2px;
}

@media (max-width: 1100px) {
  .role-editor-dialog__layout {
    grid-template-columns: 1fr;
  }

  .role-editor-dialog__panel--sticky {
    position: static;
  }

  .role-editor-dialog__overview,
  .role-editor-dialog__summary-card {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .role-editor-dialog__toolbar,
  .role-editor-dialog__section-head,
  .role-editor-dialog__section-head--workspace {
    flex-direction: column;
    align-items: stretch;
  }

  .role-editor-dialog__search {
    max-width: none;
  }

  .role-editor-dialog__browser-head {
    display: none;
  }

  .role-editor-dialog__permission-code,
  .role-editor-dialog__permission-path-tag {
    max-width: 180px;
  }
}
</style>
