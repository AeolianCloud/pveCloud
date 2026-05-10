<script setup lang="ts">
import { RefreshOutline } from '@vicons/ionicons5'
import {
  NButton,
  NCard,
  NDataTable,
  NIcon,
  NInput,
  NSwitch,
  NTabPane,
  NTabs,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onMounted, ref } from 'vue'

import EmptyState from '../../components/EmptyState.vue'
import QueryState from '../../components/QueryState.vue'
import { usePermissionStore } from '../../store/modules/permission'
import {
  getSystemConfigs,
  updateSystemConfig,
  type SystemConfigItem,
  type SystemConfigGroup,
} from '../../api/system-config'
import { message } from '../../utils/feedback'

const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const groups = ref<SystemConfigGroup[]>([])
const permissionStore = usePermissionStore()

const activeGroup = ref('')
const canViewConfig = computed(() => permissionStore.hasPermission('page.system-settings.config'))
const canUpdateConfig = computed(() => permissionStore.hasPermission('system-config:update'))

const currentGroup = computed(() => groups.value.find((g) => g.group_name === activeGroup.value))

const editForm = ref<Record<number, string>>({})

function normalizeConfigValue(config: SystemConfigItem) {
  if (config.is_secret) return ''
  if (config.value_type === 'bool') {
    return (config.config_value ?? '').trim().toLowerCase() === 'true' ? 'true' : 'false'
  }
  return config.config_value ?? ''
}

function initEditForm(group: SystemConfigGroup) {
  const form: Record<number, string> = {}
  for (const config of group.items) {
    form[config.id] = normalizeConfigValue(config)
  }
  editForm.value = form
}

async function loadConfigs() {
  if (!canViewConfig.value) {
    groups.value = []
    errorMessage.value = ''
    return
  }
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await getSystemConfigs()
    groups.value = result
    if (result.length > 0 && !activeGroup.value) {
      activeGroup.value = result[0].group_name
      initEditForm(result[0])
    } else if (result.length > 0) {
      const current = result.find((g) => g.group_name === activeGroup.value)
      if (current) initEditForm(current)
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载失败'
  } finally {
    loading.value = false
  }
}

function handleGroupChange(name: string) {
  activeGroup.value = name
  const group = groups.value.find((g) => g.group_name === name)
  if (group) initEditForm(group)
}

function isDirty(config: SystemConfigItem): boolean {
  return (editForm.value[config.id] ?? '') !== normalizeConfigValue(config)
}

function isBoolConfig(config: SystemConfigItem) {
  return config.value_type === 'bool'
}

function boolValue(config: SystemConfigItem) {
  return (editForm.value[config.id] ?? 'false') === 'true'
}

function updateBoolValue(config: SystemConfigItem, value: boolean) {
  editForm.value[config.id] = value ? 'true' : 'false'
}

async function saveConfig(config: SystemConfigItem) {
  const newValue = editForm.value[config.id] ?? ''
  saving.value = true
  try {
    await updateSystemConfig(config.id, { config_value: newValue })
    if (currentGroup.value) {
      const idx = currentGroup.value.items.findIndex((c) => c.id === config.id)
      if (idx !== -1) {
        currentGroup.value.items[idx].config_value = config.is_secret ? null : newValue
        currentGroup.value.items[idx].has_value = config.is_secret ? true : Boolean(newValue.trim())
      }
    }
    if (config.is_secret) {
      editForm.value[config.id] = ''
    }
    message.success('保存成功')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    saving.value = false
  }
}

function valueTypeTag(type: string): 'default' | 'warning' | 'success' | 'info' {
  const map: Record<string, 'default' | 'warning' | 'success' | 'info'> = {
    string: 'default',
    int: 'warning',
    bool: 'success',
    json: 'info',
  }
  return map[type] ?? 'info'
}

const columns = computed<DataTableColumns<SystemConfigItem>>(() => [
  { key: 'config_key', title: '配置键', minWidth: 200 },
  {
    key: 'value',
    title: '值',
    minWidth: 240,
    render: (row) => {
      if (row.is_secret) {
        return h('div', { class: 'secret-cell' }, [
          h(NInput, {
            value: editForm.value[row.id],
            type: 'password',
            showPasswordOn: 'click',
            disabled: !canUpdateConfig.value,
            placeholder: row.has_value ? '留空则保留旧值，输入新值则覆盖' : '请输入敏感配置值',
            'onUpdate:value': (v: string) => (editForm.value[row.id] = v),
          }),
          h(
            NTag,
            { type: row.has_value ? 'error' : 'default', size: 'small' },
            { default: () => (row.has_value ? '已配置' : '未配置') },
          ),
        ])
      }
      if (isBoolConfig(row)) {
        return h(NSwitch, {
          value: boolValue(row),
          disabled: !canUpdateConfig.value,
          'onUpdate:value': (v: boolean) => updateBoolValue(row, v),
        })
      }
      return h(NInput, {
        value: editForm.value[row.id],
        type: row.value_type === 'json' ? 'textarea' : 'text',
        rows: 3,
        disabled: !canUpdateConfig.value,
        'onUpdate:value': (v: string) => (editForm.value[row.id] = v),
      })
    },
  },
  {
    key: 'value_type',
    title: '类型',
    width: 90,
    align: 'center',
    render: (row) => h(NTag, { type: valueTypeTag(row.value_type), size: 'small' }, { default: () => row.value_type }),
  },
  { key: 'description', title: '说明', minWidth: 160, ellipsis: { tooltip: true } },
  {
    key: 'actions',
    title: '操作',
    width: 100,
    align: 'center',
    render: (row) => {
      if (!canUpdateConfig.value) return null
      return h(
        NButton,
        {
          type: 'primary',
          size: 'small',
          loading: saving.value,
          disabled: !isDirty(row),
          onClick: () => saveConfig(row),
        },
        { default: () => '保存' },
      )
    },
  },
])

onMounted(() => {
  void loadConfigs()
})
</script>

<template>
  <div class="system-settings-page">
    <div class="system-settings-page__header">
      <h2>系统设置</h2>
      <NButton :loading="loading" @click="loadConfigs">
        <template #icon>
          <NIcon><RefreshOutline /></NIcon>
        </template>
        刷新
      </NButton>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadConfigs">
      <template v-if="!canViewConfig">
        <NCard>
          <EmptyState title="暂无权限" description="当前账号没有系统配置查看权限。" />
        </NCard>
      </template>

      <template v-else-if="groups.length === 0">
        <NCard>
          <EmptyState title="暂无配置" description="当前没有可展示的系统配置。" />
        </NCard>
      </template>

      <template v-else>
        <NTabs :value="activeGroup" type="line" @update:value="handleGroupChange">
          <NTabPane v-for="group in groups" :key="group.group_name" :name="group.group_name" :tab="group.group_name" />
        </NTabs>

        <NCard v-if="currentGroup" :bordered="false">
          <NDataTable
            :columns="columns"
            :data="currentGroup.items"
            :row-key="(row: SystemConfigItem) => row.id"
            striped
            :bordered="false"
          />
        </NCard>
      </template>
    </QueryState>
  </div>
</template>

<style scoped>
.system-settings-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.system-settings-page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.system-settings-page__header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.secret-cell {
  display: grid;
  gap: 8px;
}
</style>
