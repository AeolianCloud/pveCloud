<script setup lang="ts">
import { CaretDownOutline, ChevronForwardOutline } from '@vicons/ionicons5'
import {
  NAlert,
  NButton,
  NCheckbox,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NModal,
  NRadio,
  NRadioGroup,
  NTag,
} from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import { computed, nextTick, ref, watch } from 'vue'

import type { AdminPermissionItem } from '../../../api/admin-role'
import type { PermissionTreeNode, RoleEditorState } from '../types'

const props = defineProps<{
  visible: boolean
  title: string
  isCreateMode: boolean
  isBuiltInRole: boolean
  form: RoleEditorState
  rules: FormRules
  permissionTree: AdminPermissionItem[]
  submitting: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: []
  closed: []
}>()

const formRef = ref<FormInst | null>(null)
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

const flatPermissions = computed(() => flattenPermissionTree(props.permissionTree))
const permissionCodeSet = computed(() => new Set(flatPermissions.value.map((p) => p.code)))
const permissionCount = computed(() => props.form.permission_codes.length)
const permissionTreeData = computed<PermissionTreeNode[]>(() =>
  buildPermissionTree(props.permissionTree, props.isBuiltInRole),
)
const normalizedFilter = computed(() => filterText.value.trim().toLowerCase())
const expandedNodeIdSet = computed(() => new Set(expandedNodeIds.value))
const visibleRows = computed(() => buildVisibleRows(permissionTreeData.value))
const totalPermissionCount = computed(() => flatPermissions.value.length)
const visiblePermissionCount = computed(
  () => visibleRows.value.filter((row) => row.node.type === 'action').length,
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
        formRef.value?.restoreValidation()
        syncSelectedPermissions()
        syncExpandedState(true)
      })
    }
  },
)

watch(
  () => props.permissionTree,
  () => {
    if (!props.visible) return
    void nextTick(() => {
      syncSelectedPermissions()
      syncExpandedState(false)
    })
  },
  { deep: true },
)

function uniqueSortedStrings(values: string[]) {
  return Array.from(new Set(values.map((v) => v.trim()).filter(Boolean))).sort()
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
  if (props.isBuiltInRole) return
  props.form.permission_codes = uniqueSortedStrings(flatPermissions.value.map((p) => p.code))
}

function handleClear() {
  if (props.isBuiltInRole) return
  props.form.permission_codes = []
}

function toggleExpanded(nodeId: string) {
  const next = new Set(expandedNodeIds.value)
  if (next.has(nodeId)) next.delete(nodeId)
  else next.add(nodeId)
  expandedNodeIds.value = Array.from(next)
}

function toggleNodeSelection(node: PermissionTreeNode, checked: boolean) {
  if (props.isBuiltInRole) return
  const next = new Set(props.form.permission_codes)
  for (const code of collectLeafCodes(node)) {
    if (checked) next.add(code)
    else next.delete(code)
  }
  props.form.permission_codes = uniqueSortedStrings(Array.from(next))
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  emit('submit')
}

function handleAfterLeave() {
  emit('closed')
}

function flattenPermissionTree(nodes: AdminPermissionItem[]): AdminPermissionItem[] {
  return nodes.flatMap((node) => [node, ...flattenPermissionTree(node.children ?? [])])
}

function buildPermissionTree(nodes: AdminPermissionItem[], disabled: boolean): PermissionTreeNode[] {
  const roots = nodes.map((node) => buildPermissionNode(node, disabled))
  for (const root of roots) finalizeTree(root)
  return roots
}

function buildPermissionNode(p: AdminPermissionItem, disabled: boolean): PermissionTreeNode {
  return {
    id: p.code,
    label: p.name,
    type: p.type,
    code: p.code,
    description: p.description,
    disabled,
    meta_label: p.type === 'menu' ? '菜单权限' : '操作权限',
    path_hint: p.path ?? undefined,
    keywords: [p.code, p.name, p.group_name, p.path ?? ''],
    sort_order: p.sort_order,
    children: (p.children ?? []).map((c) => buildPermissionNode(c, disabled)),
  }
}

function finalizeTree(node: PermissionTreeNode) {
  if (node.children?.length) {
    for (const child of node.children) finalizeTree(child)
    node.children.sort(compareTreeNodes)
    node.count = node.children.reduce(
      (sum, child) => sum + (child.type === 'action' ? 1 : child.count ?? 0),
      0,
    )
  }
}

function compareTreeNodes(left: PermissionTreeNode, right: PermissionTreeNode) {
  const orderDelta = (left.sort_order ?? 999) - (right.sort_order ?? 999)
  if (orderDelta !== 0) return orderDelta
  return left.label.localeCompare(right.label, 'zh-CN')
}

function isPermissionNode(node: PermissionTreeNode) {
  return node.type === 'action'
}

function isBranchNode(node: PermissionTreeNode) {
  return (node.children?.length ?? 0) > 0
}

function getPrimaryLabel(node: PermissionTreeNode) {
  if (isPermissionNode(node)) return node.meta_label || node.label
  return node.label
}

function getSecondaryLabel(node: PermissionTreeNode) {
  if (isPermissionNode(node)) {
    return node.meta_label && node.meta_label !== node.label ? node.label : ''
  }
  return node.meta_label || ''
}

function collectLeafCodes(node: PermissionTreeNode): string[] {
  if (isPermissionNode(node) && node.code) return [node.code]
  return (node.children ?? []).flatMap(collectLeafCodes)
}

function nodeMatchesFilter(node: PermissionTreeNode) {
  if (!normalizedFilter.value) return true
  return [node.label, node.meta_label, node.code, node.description, node.path_hint, ...(node.keywords ?? [])]
    .filter(Boolean)
    .some((item) => String(item).toLowerCase().includes(normalizedFilter.value))
}

function hasVisibleDescendant(node: PermissionTreeNode): boolean {
  return (node.children ?? []).some(shouldShowNode)
}

function shouldShowNode(node: PermissionTreeNode) {
  if (!normalizedFilter.value) return true
  return nodeMatchesFilter(node) || hasVisibleDescendant(node)
}

function buildVisibleRows(nodes: PermissionTreeNode[], depth = 0): PermissionTreeRow[] {
  const rows: PermissionTreeRow[] = []
  const selected = new Set(props.form.permission_codes)
  const forceExpand = Boolean(normalizedFilter.value)
  for (const node of nodes) {
    if (!shouldShowNode(node)) continue
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
      for (const childId of collectBranchIds(node.children ?? [])) branchIds.add(childId)
    }
  }
  return branchIds
}

function getInitialExpandedIds(nodes: PermissionTreeNode[]) {
  const result: string[] = []
  for (const node of nodes) {
    if (!isBranchNode(node)) continue
    result.push(node.id)
    for (const child of node.children ?? []) {
      if (isBranchNode(child)) result.push(child.id)
    }
  }
  return result
}
</script>

<template>
  <NModal
    :show="props.visible"
    preset="card"
    :title="props.title"
    style="width: 1120px; max-width: 96vw"
    :mask-closable="false"
    :on-after-leave="handleAfterLeave"
    @update:show="emit('update:visible', $event)"
  >
    <NForm ref="formRef" :model="props.form" :rules="props.rules as any" label-placement="top" class="role-editor-dialog__form">
      <div class="role-editor-dialog__layout">
        <section class="role-editor-dialog__sidebar">
          <div class="role-editor-dialog__panel">
            <div class="role-editor-dialog__section-head">
              <div>
                <h3>基础信息</h3>
                <p>{{ dialogDescription }}</p>
              </div>
              <NTag size="small">{{ props.isCreateMode ? '新建模式' : '编辑模式' }}</NTag>
            </div>

            <div class="role-editor-dialog__field-list">
              <NFormItem label="管理组编码" path="code">
                <NInput
                  v-model:value="props.form.code"
                  :disabled="!props.isCreateMode"
                  placeholder="请输入唯一编码，例如 ops_manager"
                />
              </NFormItem>
              <NFormItem label="管理组名称" path="name">
                <NInput v-model:value="props.form.name" placeholder="请输入管理组名称" />
              </NFormItem>
              <NFormItem label="说明" path="description">
                <NInput
                  v-model:value="props.form.description"
                  type="textarea"
                  :rows="4"
                  placeholder="请输入管理组说明，可留空"
                />
              </NFormItem>
              <NFormItem label="状态" path="status">
                <NRadioGroup v-model:value="props.form.status">
                  <NRadio value="active" :disabled="props.isBuiltInRole">启用</NRadio>
                  <NRadio value="disabled" :disabled="props.isBuiltInRole">停用</NRadio>
                </NRadioGroup>
              </NFormItem>
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
                <strong>
                  {{ totalPermissionCount === 0 ? '0%' : `${Math.round((permissionCount / totalPermissionCount) * 100)}%` }}
                </strong>
              </div>
            </div>

            <NAlert v-if="props.isBuiltInRole" type="warning" :show-icon="true">
              内置超级管理员角色不可修改状态和权限分配。
            </NAlert>
          </div>
        </section>

        <section class="role-editor-dialog__workspace">
          <div class="role-editor-dialog__panel">
            <div class="role-editor-dialog__section-head">
              <div>
                <h3>权限工作台</h3>
                <p>按真实菜单目录管理页面入口权限和资源操作权限。</p>
              </div>
              <div class="role-editor-dialog__section-tags">
                <NTag type="primary" size="small">目录化分配</NTag>
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

            <div class="role-editor-dialog__toolbar">
              <NInput
                v-model:value="filterText"
                clearable
                placeholder="搜索目录、权限名称、权限码"
                class="role-editor-dialog__search"
              />
              <div class="role-editor-dialog__toolbar-actions">
                <NButton text type="primary" :disabled="props.isBuiltInRole" @click="handleCheckAll">全选全部权限</NButton>
                <NButton text :disabled="props.isBuiltInRole || permissionCount === 0" @click="handleClear">清空选择</NButton>
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
                      <NIcon :size="14">
                        <CaretDownOutline v-if="row.expanded" />
                        <ChevronForwardOutline v-else />
                      </NIcon>
                    </button>
                    <span v-else class="role-editor-dialog__tree-toggle-placeholder" />

                    <NCheckbox
                      :checked="row.checked"
                      :indeterminate="row.indeterminate"
                      :disabled="props.isBuiltInRole || row.node.disabled"
                      @update:checked="toggleNodeSelection(row.node, $event)"
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
                          <span
                            v-if="!isPermissionNode(row.node) && row.node.path_hint"
                            class="role-editor-dialog__permission-path-tag"
                          >{{ row.node.path_hint }}</span>
                        </div>
                        <span v-if="getSecondaryLabel(row.node)" class="role-editor-dialog__permission-meta">{{
                          getSecondaryLabel(row.node)
                        }}</span>
                        <span
                          v-if="row.node.type === 'action' && row.node.description"
                          class="role-editor-dialog__permission-description"
                        >{{ row.node.description }}</span>
                      </div>
                      <span
                        v-if="row.node.type === 'action' && row.node.code"
                        class="role-editor-dialog__permission-code"
                      >{{ row.node.code }}</span>
                      <NTag v-else size="small">{{ row.checkedLeafCount }}/{{ row.totalLeafCount }}</NTag>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </NForm>

    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <NButton @click="emit('update:visible', false)">取消</NButton>
        <NButton type="primary" :loading="props.submitting" @click="handleSubmit">保存</NButton>
      </div>
    </template>
  </NModal>
</template>

<style scoped>
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
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.04);
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
  color: rgba(15, 23, 42, 0.92);
}

.role-editor-dialog__section-head p {
  margin: 6px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: rgba(15, 23, 42, 0.55);
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

.role-editor-dialog__summary-card,
.role-editor-dialog__overview {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.role-editor-dialog__summary-item,
.role-editor-dialog__overview-card {
  padding: 14px 12px;
  border-radius: 14px;
  background: rgba(15, 23, 42, 0.03);
}

.role-editor-dialog__summary-item span,
.role-editor-dialog__overview-card span {
  display: block;
  font-size: 12px;
  color: rgba(15, 23, 42, 0.55);
}

.role-editor-dialog__summary-item strong,
.role-editor-dialog__overview-card strong {
  display: block;
  margin-top: 8px;
  font-size: 24px;
  line-height: 1;
  color: rgba(15, 23, 42, 0.92);
}

.role-editor-dialog__overview-card small {
  display: block;
  margin-top: 8px;
  line-height: 1.5;
  color: rgba(15, 23, 42, 0.55);
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
  color: rgba(15, 23, 42, 0.55);
}

.role-editor-dialog__permission-tree-wrap {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 16px;
  background: #fff;
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
  color: rgba(15, 23, 42, 0.55);
  font-size: 13px;
}

.role-editor-dialog__tree-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  min-height: 52px;
  padding: 6px 8px;
  border-radius: 10px;
  transition: background-color 0.2s ease;
}

.role-editor-dialog__tree-row:hover {
  background: rgba(15, 23, 42, 0.04);
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
  color: rgba(15, 23, 42, 0.55);
  cursor: pointer;
}

.role-editor-dialog__tree-toggle:hover {
  background: rgba(15, 23, 42, 0.06);
  color: rgba(15, 23, 42, 0.92);
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
  color: rgba(15, 23, 42, 0.92);
}

.role-editor-dialog__permission-meta,
.role-editor-dialog__permission-description {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
  line-height: 1.4;
}

.role-editor-dialog__permission-path-tag,
.role-editor-dialog__permission-code {
  flex-shrink: 0;
  font-size: 12px;
  line-height: 1.4;
  font-family: ui-monospace, SFMono-Regular, Consolas, 'Liberation Mono', Menlo, monospace;
}

.role-editor-dialog__permission-path-tag {
  max-width: 260px;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.08);
  color: rgba(37, 99, 235, 0.92);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.role-editor-dialog__permission-code {
  color: rgba(15, 23, 42, 0.6);
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1100px) {
  .role-editor-dialog__layout {
    grid-template-columns: 1fr;
  }

  .role-editor-dialog__overview,
  .role-editor-dialog__summary-card {
    grid-template-columns: 1fr;
  }
}
</style>
