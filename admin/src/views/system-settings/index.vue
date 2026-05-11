<script setup lang="ts">
import { RefreshOutline } from '@vicons/ionicons5'
import {
  NBadge,
  NButton,
  NCard,
  NDataTable,
  NDivider,
  NIcon,
  NInput,
  NScrollbar,
  NSwitch,
  NTag,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onBeforeUnmount, onMounted, ref } from 'vue'

import EmptyState from '../../components/EmptyState.vue'
import QueryState from '../../components/QueryState.vue'
import { usePermissionStore } from '../../store/modules/permission'
import { downloadFile, uploadFile } from '../../api/file-attachment'
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
const canUploadLogo = computed(() => permissionStore.hasPermission('file:upload'))

const currentGroup = computed(() => groups.value.find((g) => g.group_name === activeGroup.value))

const editForm = ref<Record<number, string>>({})
const uploadingLogoId = ref<number | null>(null)
const logoPreviewUrls = ref<Record<number, string>>({})

const currentGroupDirtyCount = computed(() => currentGroup.value?.items.filter((item) => isDirty(item)).length ?? 0)

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
  void loadLogoPreviews(group)
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

function isLogoConfig(config: SystemConfigItem) {
  return config.config_key === 'site.logo_url'
}

function logoValue(config: SystemConfigItem) {
  return editForm.value[config.id] || normalizeConfigValue(config)
}

function logoPreviewSrc(config: SystemConfigItem) {
  return logoPreviewUrls.value[config.id] || logoValue(config)
}

function fileIdFromAdminFileUrl(value: string) {
  const match = value.match(/\/admin-api\/files\/(\d+)(?:\/download)?(?:\?.*)?$/)
  return match ? Number(match[1]) : null
}

function setLogoPreviewUrl(configId: number, url: string) {
  const oldUrl = logoPreviewUrls.value[configId]
  if (oldUrl?.startsWith('blob:')) {
    window.URL.revokeObjectURL(oldUrl)
  }
  logoPreviewUrls.value = { ...logoPreviewUrls.value, [configId]: url }
}

async function loadLogoPreviews(group: SystemConfigGroup) {
  const logoConfigs = group.items.filter(isLogoConfig)
  await Promise.all(
    logoConfigs.map(async (config) => {
      const value = logoValue(config)
      if (!value || logoPreviewUrls.value[config.id]) return
      const fileId = fileIdFromAdminFileUrl(value)
      if (!fileId) return
      try {
        const blob = await downloadFile(fileId)
        setLogoPreviewUrl(config.id, window.URL.createObjectURL(blob))
      } catch {
        // Preview is best-effort; saving still uses the stored config value.
      }
    }),
  )
}

function clearLogoPreviewUrls() {
  for (const url of Object.values(logoPreviewUrls.value)) {
    if (url.startsWith('blob:')) {
      window.URL.revokeObjectURL(url)
    }
  }
  logoPreviewUrls.value = {}
}

function openLogoUpload(config: SystemConfigItem) {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/png,image/jpeg,image/gif,image/webp'
  input.onchange = () => {
    const file = input.files?.[0]
    if (file) void handleLogoUpload(config, file)
  }
  input.click()
}

async function handleLogoUpload(config: SystemConfigItem, file: File) {
  if (file.size <= 0) {
    message.error('Logo 文件内容不能为空')
    return
  }
  if (!file.type.startsWith('image/')) {
    message.error('Logo 仅支持图片文件')
    return
  }

  uploadingLogoId.value = config.id
  try {
    const result = await uploadFile(file)
    editForm.value[config.id] = result.url
    setLogoPreviewUrl(config.id, window.URL.createObjectURL(file))
    message.success('Logo 上传成功，请保存配置生效')
  } catch (error) {
    message.error(error instanceof Error ? error.message : 'Logo 上传失败')
  } finally {
    uploadingLogoId.value = null
  }
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

function formatUpdatedAt(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

function formatGroupSummary(group: SystemConfigGroup) {
  const secretCount = group.items.filter((item) => item.is_secret).length
  const boolCount = group.items.filter((item) => item.value_type === 'bool').length
  const parts = [`${group.items.length} 项配置`]
  if (boolCount > 0) parts.push(`${boolCount} 个开关`)
  if (secretCount > 0) parts.push(`${secretCount} 个敏感项`)
  return parts.join(' / ')
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
  {
    key: 'config_key',
    title: '配置项',
    minWidth: 260,
    render: (row) =>
      h('div', { class: 'config-key-cell' }, [
        h('div', { class: 'config-key-cell__main' }, [
          h('span', { class: 'config-key-cell__key' }, row.config_key),
          row.is_secret
            ? h(NTag, { type: row.has_value ? 'success' : 'default', size: 'small', round: true }, { default: () => (row.has_value ? '已配置密钥' : '未配置密钥') })
            : null,
          isDirty(row) ? h(NTag, { type: 'warning', size: 'small', round: true }, { default: () => '未保存' }) : null,
        ]),
        h('div', { class: 'config-key-cell__description' }, row.description || '暂无说明'),
        h('div', { class: 'config-key-cell__meta' }, `最近更新：${formatUpdatedAt(row.updated_at)}`),
      ]),
  },
  {
    key: 'value',
    title: '配置值',
    minWidth: 320,
    render: (row) => {
      if (isLogoConfig(row)) {
        const value = logoValue(row)
        const previewSrc = logoPreviewSrc(row)
        return h('div', { class: 'logo-upload-cell' }, [
          h(
            'div',
            { class: ['logo-upload-cell__preview', value ? 'logo-upload-cell__preview--filled' : ''] },
            previewSrc ? h('img', { src: previewSrc, alt: '站点 Logo' }) : h('span', null, '未设置 Logo'),
          ),
          h('div', { class: 'logo-upload-cell__actions' }, [
            h(
              NButton,
              {
                size: 'small',
                type: 'primary',
                ghost: true,
                disabled: !canUpdateConfig.value || !canUploadLogo.value,
                loading: uploadingLogoId.value === row.id,
                onClick: () => openLogoUpload(row),
              },
              { default: () => (value ? '更换 Logo' : '上传 Logo') },
            ),
            h('span', { class: 'value-cell__hint' }, '上传后点击本行保存才会写入系统配置。'),
          ]),
        ])
      }
      if (row.is_secret) {
        return h('div', { class: 'value-cell value-cell--secret' }, [
          h(NInput, {
            value: editForm.value[row.id],
            type: 'password',
            showPasswordOn: 'click',
            disabled: !canUpdateConfig.value,
            placeholder: row.has_value ? '留空则保留旧值，输入新值则覆盖' : '请输入敏感配置值',
            'onUpdate:value': (v: string) => (editForm.value[row.id] = v),
          }),
          h('span', { class: 'value-cell__hint' }, row.has_value ? '当前值已加密保存，留空提交不会覆盖。' : '尚未配置，保存非空值后生效。'),
        ])
      }
      if (isBoolConfig(row)) {
        return h('div', { class: 'value-cell value-cell--switch' }, [
          h(NSwitch, {
            value: boolValue(row),
            disabled: !canUpdateConfig.value,
            'onUpdate:value': (v: boolean) => updateBoolValue(row, v),
          }),
          h('span', { class: 'value-cell__hint' }, boolValue(row) ? '当前选择：启用' : '当前选择：停用'),
        ])
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
  {
    key: 'actions',
    title: '操作',
    width: 110,
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

onBeforeUnmount(() => {
  clearLogoPreviewUrls()
})
</script>

<template>
  <div class="system-settings-page">
    <div class="system-settings-page__header">
      <NButton :loading="loading" @click="loadConfigs">
        <template #icon>
          <NIcon><RefreshOutline /></NIcon>
        </template>
        刷新配置
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
        <div class="system-settings-workspace">
          <NCard :bordered="false" class="system-settings-nav-card">
            <template #header>配置分组</template>
            <NScrollbar style="max-height: 560px">
              <div class="system-settings-nav">
                <button
                  v-for="group in groups"
                  :key="group.group_name"
                  class="system-settings-nav__item"
                  :class="{ 'system-settings-nav__item--active': group.group_name === activeGroup }"
                  type="button"
                  @click="handleGroupChange(group.group_name)"
                >
                  <span class="system-settings-nav__title">{{ group.group_name }}</span>
                  <span class="system-settings-nav__summary">{{ formatGroupSummary(group) }}</span>
                  <NBadge
                    v-if="group.items.some((item) => isDirty(item))"
                    :value="group.items.filter((item) => isDirty(item)).length"
                    type="warning"
                    class="system-settings-nav__badge"
                  />
                </button>
              </div>
            </NScrollbar>
          </NCard>

          <NCard v-if="currentGroup" :bordered="false" class="system-settings-detail-card">
            <template #header>
              <div class="system-settings-detail-card__header">
                <div>
                  <h3>{{ currentGroup.group_name }}</h3>
                  <p>{{ formatGroupSummary(currentGroup) }}</p>
                </div>
                <NTag v-if="currentGroupDirtyCount > 0" type="warning" round>{{ currentGroupDirtyCount }} 项未保存</NTag>
              </div>
            </template>

            <NDivider class="system-settings-detail-card__divider" />

            <NDataTable
              :columns="columns"
              :data="currentGroup.items"
              :row-key="(row: SystemConfigItem) => row.id"
              striped
              :bordered="false"
              :single-line="false"
              class="system-settings-table"
            />
          </NCard>
        </div>
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
  justify-content: flex-end;
}

.system-settings-workspace {
  display: grid;
  grid-template-columns: minmax(240px, 280px) minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.system-settings-nav-card,
.system-settings-detail-card {
  border-radius: 16px;
  box-shadow: 0 14px 36px rgba(15, 23, 42, 0.05);
}

.system-settings-nav {
  display: grid;
  gap: 8px;
}

.system-settings-nav__item {
  position: relative;
  display: grid;
  width: 100%;
  gap: 4px;
  padding: 13px 14px;
  border: 1px solid transparent;
  border-radius: 12px;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    transform 0.2s ease;
}

.system-settings-nav__item:hover {
  border-color: rgba(37, 99, 235, 0.18);
  background: #f8fafc;
  transform: translateY(-1px);
}

.system-settings-nav__item--active {
  border-color: rgba(37, 99, 235, 0.3);
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.1), rgba(14, 165, 233, 0.06));
}

.system-settings-nav__title {
  color: #0f172a;
  font-size: 14px;
  font-weight: 600;
}

.system-settings-nav__summary {
  padding-right: 28px;
  color: #64748b;
  font-size: 12px;
}

.system-settings-nav__badge {
  position: absolute;
  top: 12px;
  right: 12px;
}

.system-settings-detail-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.system-settings-detail-card__header h3 {
  margin: 0;
  color: #0f172a;
  font-size: 18px;
  font-weight: 700;
}

.system-settings-detail-card__header p {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 13px;
}

.system-settings-detail-card__divider {
  margin: 0 0 12px;
}

.system-settings-table :deep(.n-data-table-th) {
  background: #f8fafc;
  color: #334155;
  font-weight: 700;
}

.config-key-cell {
  display: grid;
  gap: 7px;
  padding: 4px 0;
}

.config-key-cell__main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.config-key-cell__key {
  color: #0f172a;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
  font-weight: 700;
}

.config-key-cell__description,
.config-key-cell__meta,
.value-cell__hint {
  color: #64748b;
  font-size: 12px;
  line-height: 1.5;
}

.config-key-cell__meta {
  color: #94a3b8;
}

.value-cell {
  display: grid;
  gap: 8px;
}

.value-cell--switch {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-upload-cell {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 14px;
}

.logo-upload-cell__preview {
  display: flex;
  width: 120px;
  height: 54px;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  background: #f8fafc;
  color: #94a3b8;
  font-size: 12px;
}

.logo-upload-cell__preview--filled {
  border-style: solid;
  background: #fff;
}

.logo-upload-cell__preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.logo-upload-cell__actions {
  display: grid;
  gap: 6px;
}

@media (max-width: 1100px) {
  .system-settings-workspace {
    grid-template-columns: 1fr;
  }

  .system-settings-nav-card :deep(.n-card-header) {
    display: none;
  }

  .system-settings-nav {
    display: flex;
    gap: 10px;
    overflow-x: auto;
    padding-bottom: 4px;
  }

  .system-settings-nav__item {
    min-width: 220px;
  }
}

@media (max-width: 720px) {
  .system-settings-detail-card__header {
    flex-direction: column;
  }
}
</style>
