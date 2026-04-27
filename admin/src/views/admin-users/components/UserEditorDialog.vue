<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

import type { AdminRoleItem } from '../../../api/admin-role'
import type { UserEditorState } from '../types'

const props = defineProps<{
  visible: boolean
  title: string
  isCreateMode: boolean
  form: UserEditorState
  rules: FormRules<UserEditorState>
  roleOptions: AdminRoleItem[]
  canReadRoleOptions: boolean
  submitting: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: []
  closed: []
}>()

const formRef = ref<FormInstance>()

watch(
  () => props.visible,
  (value) => {
    if (value) {
      void nextTick(() => {
        formRef.value?.clearValidate()
      })
    }
  },
)

function formatRoleOptionLabel(role: AdminRoleItem) {
  return role.status === 'active' ? role.name : `${role.name}（已停用）`
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
    width="640px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
    @closed="emit('closed')"
  >
    <el-form ref="formRef" :model="props.form" :rules="props.rules" label-width="96px">
      <el-form-item label="登录账号" prop="username">
        <el-input
          v-model="props.form.username"
          :disabled="!props.isCreateMode"
          placeholder="请输入 3 到 64 位账号"
        />
      </el-form-item>
      <el-form-item label="显示名称" prop="display_name">
        <el-input v-model="props.form.display_name" placeholder="请输入管理员显示名称" />
      </el-form-item>
      <el-form-item label="邮箱" prop="email">
        <el-input v-model="props.form.email" placeholder="请输入邮箱，可留空" />
      </el-form-item>
      <el-form-item v-if="props.isCreateMode" label="登录密码" prop="password">
        <el-input
          v-model="props.form.password"
          type="password"
          show-password
          placeholder="请输入 6 到 72 位密码"
        />
      </el-form-item>
      <el-form-item label="账号状态" prop="status">
        <el-radio-group v-model="props.form.status">
          <el-radio value="active">启用</el-radio>
          <el-radio value="disabled">停用</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="props.canReadRoleOptions" label="角色分配" prop="role_ids">
        <el-select
          v-model="props.form.role_ids"
          multiple
          filterable
          collapse-tags
          collapse-tags-tooltip
          placeholder="请选择要分配的角色"
        >
          <el-option
            v-for="role in props.roleOptions"
            :key="role.id"
            :label="formatRoleOptionLabel(role)"
            :value="role.id"
            :disabled="role.status !== 'active'"
          />
        </el-select>
      </el-form-item>
      <el-form-item v-if="props.canReadRoleOptions">
        <el-alert
          type="info"
          :closable="false"
          title="仅启用中的角色可分配给管理员。已停用角色会保留显示，但不能再次分配。"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="props.submitting" @click="handleSubmit">保存</el-button>
    </template>
  </el-dialog>
</template>
