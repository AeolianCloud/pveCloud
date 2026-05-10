<script setup lang="ts">
import { NAlert, NButton, NForm, NFormItem, NInput, NModal, NRadio, NRadioGroup, NSelect } from 'naive-ui'
import { computed, nextTick, ref, watch } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'

import type { AdminRoleItem } from '../../../api/admin-role'
import type { UserEditorState } from '../types'

const props = defineProps<{
  visible: boolean
  title: string
  isCreateMode: boolean
  form: UserEditorState
  rules: FormRules
  roleOptions: AdminRoleItem[]
  canReadRoleOptions: boolean
  submitting: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: []
  closed: []
}>()

const formRef = ref<FormInst | null>(null)

watch(
  () => props.visible,
  (value) => {
    if (value) {
      void nextTick(() => formRef.value?.restoreValidation())
    }
  },
)

const roleSelectOptions = computed(() =>
  props.roleOptions.map((role) => ({
    label: role.status === 'active' ? role.name : `${role.name}（已停用）`,
    value: role.id,
    disabled: role.status !== 'active',
  })),
)

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  emit('submit')
}

function handleAfterLeave() {
  emit('closed')
}
</script>

<template>
  <NModal
    :show="props.visible"
    preset="card"
    :title="props.title"
    style="width: 640px"
    :mask-closable="false"
    :on-after-leave="handleAfterLeave"
    @update:show="emit('update:visible', $event)"
  >
    <NForm ref="formRef" :model="props.form" :rules="props.rules as any" label-placement="top">
      <NFormItem label="登录账号" path="username">
        <NInput
          v-model:value="props.form.username"
          :disabled="!props.isCreateMode"
          placeholder="请输入 3 到 64 位账号"
        />
      </NFormItem>
      <NFormItem label="显示名称" path="display_name">
        <NInput v-model:value="props.form.display_name" placeholder="请输入管理员显示名称" />
      </NFormItem>
      <NFormItem label="邮箱" path="email">
        <NInput v-model:value="props.form.email" placeholder="请输入邮箱，可留空" />
      </NFormItem>
      <NFormItem v-if="props.isCreateMode" label="登录密码" path="password">
        <NInput
          v-model:value="props.form.password"
          type="password"
          show-password-on="click"
          placeholder="请输入 6 到 72 位密码"
        />
      </NFormItem>
      <NFormItem label="账号状态" path="status">
        <NRadioGroup v-model:value="props.form.status">
          <NRadio value="active">启用</NRadio>
          <NRadio value="disabled">停用</NRadio>
        </NRadioGroup>
      </NFormItem>
      <NFormItem v-if="props.canReadRoleOptions" label="角色分配" path="role_ids">
        <NSelect
          v-model:value="props.form.role_ids"
          :options="roleSelectOptions"
          multiple
          filterable
          placeholder="请选择要分配的角色"
        />
      </NFormItem>
      <NFormItem v-if="props.canReadRoleOptions" :show-label="false">
        <NAlert type="info" :show-icon="true">
          仅启用中的角色可分配给管理员。已停用角色会保留显示，但不能再次分配。
        </NAlert>
      </NFormItem>
    </NForm>

    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 8px;">
        <NButton @click="emit('update:visible', false)">取消</NButton>
        <NButton type="primary" :loading="props.submitting" @click="handleSubmit">保存</NButton>
      </div>
    </template>
  </NModal>
</template>
