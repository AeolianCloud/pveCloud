<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import {
  type FilterNodeMethodFunction,
  type FormInstance,
  type FormRules,
  type TreeInstance,
} from 'element-plus'

import type { AdminPermissionGroup } from '../../../api/admin-role'
import type { PermissionTreeNode, RoleEditorState } from '../types'

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
const treeRef = ref<TreeInstance>()
const filterText = ref('')

const permissionCount = computed(() => props.form.permission_codes.length)
const permissionTreeData = computed<PermissionTreeNode[]>(() =>
  props.permissionGroups.map((group) => ({
    id: `group:${group.group_name}`,
    label: group.group_name,
    type: 'group',
    count: group.permissions.length,
    disabled: props.isBuiltInRole,
    children: group.permissions.map((permission) => ({
      id: permission.code,
      label: permission.name,
      type: 'permission',
      code: permission.code,
      description: permission.description,
      disabled: props.isBuiltInRole,
    })),
  })),
)

const permissionTreeProps = {
  children: 'children',
  label: 'label',
  disabled: 'disabled',
}

watch(filterText, (value) => {
  treeRef.value?.filter(value)
})

watch(
  () => props.visible,
  (value) => {
    if (value) {
      filterText.value = ''
      void nextTick(() => {
        formRef.value?.clearValidate()
        syncTree()
      })
    }
  },
)

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

function uniqueSortedStrings(values: string[]) {
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean))).sort()
}

function syncTree() {
  const keys = props.form.permission_codes.filter((code) => !code.startsWith('group:'))
  treeRef.value?.setCheckedKeys(keys, false)
}

function handleTreeCheck(_data: PermissionTreeNode, checked: { checkedKeys: unknown[] }) {
  props.form.permission_codes = uniqueSortedStrings(
    checked.checkedKeys.filter((key): key is string => typeof key === 'string' && !key.startsWith('group:')),
  )
}

function handleCheckAll() {
  if (props.isBuiltInRole) {
    return
  }
  const keys = props.permissionGroups.flatMap((group) => group.permissions.map((permission) => permission.code))
  treeRef.value?.setCheckedKeys(keys, false)
  props.form.permission_codes = uniqueSortedStrings(keys)
}

function handleClear() {
  if (props.isBuiltInRole) {
    return
  }
  treeRef.value?.setCheckedKeys([], false)
  props.form.permission_codes = []
}

async function handleSubmit() {
  if (!formRef.value) {
    return
  }
  await formRef.value.validate()
  emit('submit')
}
</script>

<template>
  <el-dialog
    :model-value="props.visible"
    :title="props.title"
    width="760px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
    @closed="emit('closed')"
  >
    <el-form ref="formRef" :model="props.form" :rules="props.rules" label-width="96px">
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
          :rows="3"
          placeholder="请输入管理组说明，可留空"
        />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="props.form.status">
          <el-radio value="active" :disabled="props.isBuiltInRole">启用</el-radio>
          <el-radio value="disabled" :disabled="props.isBuiltInRole">停用</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="权限分配">
        <div class="role-editor-dialog__permission-panel">
          <div class="role-editor-dialog__permission-head">
            <span>已选 {{ permissionCount }} 项权限</span>
            <div class="role-editor-dialog__permission-tools">
              <el-button link type="primary" :disabled="props.isBuiltInRole" @click="handleCheckAll">全选</el-button>
              <el-button link :disabled="props.isBuiltInRole || permissionCount === 0" @click="handleClear">清空</el-button>
              <el-tag v-if="props.isBuiltInRole" type="warning" size="small">内置超级管理员角色不可修改权限</el-tag>
            </div>
          </div>
          <el-input
            v-model="filterText"
            clearable
            placeholder="筛选权限组、权限名称或权限码"
          />
          <div class="role-editor-dialog__permission-tree-wrap">
            <el-tree
              ref="treeRef"
              :data="permissionTreeData"
              :props="permissionTreeProps"
              node-key="id"
              show-checkbox
              default-expand-all
              :expand-on-click-node="false"
              :check-on-click-node="false"
              :check-on-click-leaf="false"
              :filter-node-method="filterPermissionNode"
              class="role-editor-dialog__permission-tree"
              @check="handleTreeCheck"
            >
              <template #default="{ data }">
                <div
                  class="role-editor-dialog__permission-node"
                  :class="{ 'role-editor-dialog__permission-node--group': data.type === 'group' }"
                  :title="data.type === 'permission' ? [data.label, data.code, data.description].filter(Boolean).join(' / ') : data.label"
                >
                  <span class="role-editor-dialog__permission-label">{{ data.label }}</span>
                  <span v-if="data.type === 'permission' && data.code" class="role-editor-dialog__permission-code">
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
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="props.submitting" @click="handleSubmit">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.role-editor-dialog__permission-panel {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.role-editor-dialog__permission-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.role-editor-dialog__permission-tools {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 12px;
}

.role-editor-dialog__permission-tree-wrap {
  border: 1px solid var(--el-border-color-light);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
  padding: 12px;
  max-height: 420px;
  overflow: auto;
}

.role-editor-dialog__permission-tree {
  background: transparent;
}

.role-editor-dialog__permission-tree :deep(.el-tree-node__content) {
  min-height: 40px;
  border-radius: 8px;
  padding-right: 8px;
}

.role-editor-dialog__permission-tree :deep(.el-tree-node__content:hover) {
  background: var(--el-fill-color-light);
}

.role-editor-dialog__permission-node {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.role-editor-dialog__permission-node--group {
  font-weight: 600;
}

.role-editor-dialog__permission-label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.role-editor-dialog__permission-code {
  flex-shrink: 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1;
}
</style>
