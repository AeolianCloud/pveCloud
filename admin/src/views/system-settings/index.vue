<script setup lang="ts">
import { Refresh } from '@element-plus/icons-vue'
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'

import EmptyState from '../../components/EmptyState.vue'
import QueryState from '../../components/QueryState.vue'
import { usePermissionStore } from '../../store/modules/permission'
import {
  getSystemConfigs,
  updateSystemConfig,
  type SystemConfigItem,
  type SystemConfigGroup,
} from '../../api/system-config'

const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const groups = ref<SystemConfigGroup[]>([])
const permissionStore = usePermissionStore()

const activeGroup = ref('')
const canViewConfig = computed(() => permissionStore.hasPermission('system-config:view'))
const canUpdateConfig = computed(() => permissionStore.hasPermission('system-config:update'))

const currentGroup = computed(() =>
  groups.value.find((g) => g.group_name === activeGroup.value),
)

const editForm = ref<Record<number, string>>({})

function initEditForm(group: SystemConfigGroup) {
  const form: Record<number, string> = {}
  for (const config of group.items) {
    form[config.id] = config.is_secret ? '' : (config.config_value ?? '')
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
  if (config.is_secret) return false
  return (editForm.value[config.id] ?? '') !== (config.config_value ?? '')
}

async function saveConfig(config: SystemConfigItem) {
  const newValue = editForm.value[config.id] ?? ''
  saving.value = true
  try {
    await updateSystemConfig(config.id, { config_value: newValue })
    if (currentGroup.value) {
      const idx = currentGroup.value.items.findIndex((c) => c.id === config.id)
      if (idx !== -1) {
        currentGroup.value.items[idx].config_value = newValue
      }
    }
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    saving.value = false
  }
}

function valueTypeTag(type: string) {
  const map: Record<string, string> = { string: '', int: 'warning', bool: 'success', json: 'info' }
  return map[type] ?? 'info'
}

onMounted(() => {
  void loadConfigs()
})
</script>

<template>
  <div class="system-settings-page">
    <div class="system-settings-page__header">
      <h2>系统设置</h2>
      <el-button :icon="Refresh" :loading="loading" @click="loadConfigs">刷新</el-button>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadConfigs">
      <template v-if="!canViewConfig">
        <el-card>
          <EmptyState title="暂无权限" description="当前账号没有系统配置查看权限。" />
        </el-card>
      </template>

      <template v-else-if="groups.length === 0">
        <el-card>
          <EmptyState title="暂无配置" description="当前没有可展示的系统配置。" />
        </el-card>
      </template>

      <template v-else>
        <el-tabs v-model="activeGroup" @tab-change="handleGroupChange">
          <el-tab-pane v-for="group in groups" :key="group.group_name" :label="group.group_name" :name="group.group_name" />
        </el-tabs>

        <el-card v-if="currentGroup">
          <el-table :data="currentGroup.items" stripe>
            <el-table-column label="配置键" prop="config_key" min-width="200" />
            <el-table-column label="值" min-width="240">
              <template #default="{ row }">
                <template v-if="row.is_secret">
                  <el-tag type="danger" size="small">敏感配置</el-tag>
                </template>
                <template v-else>
                  <el-input
                    v-model="editForm[row.id]"
                    size="small"
                    :type="row.value_type === 'json' ? 'textarea' : 'text'"
                    :rows="3"
                    :disabled="!canUpdateConfig"
                  />
                </template>
              </template>
            </el-table-column>
            <el-table-column label="类型" width="90" align="center">
              <template #default="{ row }">
                <el-tag :type="valueTypeTag(row.value_type)" size="small">{{ row.value_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="说明" prop="description" min-width="160" show-overflow-tooltip />
            <el-table-column label="操作" width="100" align="center">
              <template #default="{ row }">
                <el-button
                  v-if="!row.is_secret && canUpdateConfig"
                  type="primary"
                  size="small"
                  :loading="saving"
                  :disabled="!isDirty(row)"
                  @click="saveConfig(row)"
                >
                  保存
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
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
</style>
