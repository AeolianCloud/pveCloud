<script setup lang="ts">
import {
  CloudUploadOutline,
  DocumentOutline,
  DownloadOutline,
  RefreshOutline,
  SearchOutline,
  TrashOutline,
} from '@vicons/ionicons5'
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NForm,
  NFormItem,
  NIcon,
  NImage,
  NInput,
  NModal,
  NPagination,
  NSpace,
  NTimeline,
  NTimelineItem,
  type DataTableColumns,
} from 'naive-ui'
import { computed, h, onBeforeUnmount, onMounted, reactive, ref } from 'vue'

import EmptyState from '../../components/EmptyState.vue'
import QueryState from '../../components/QueryState.vue'
import {
  deleteFile,
  downloadFile,
  getFileDetail,
  getFiles,
  uploadFile,
  type FileDetailResponse,
  type FileItem,
  type FileListQuery,
} from '../../api/file-attachment'
import { usePermissionStore } from '../../store/modules/permission'
import { confirm, message } from '../../utils/feedback'

const permissionStore = usePermissionStore()
const loading = ref(false)
const refreshing = ref(false)
const uploading = ref(false)
const deletingId = ref<number | null>(null)
const errorMessage = ref('')
const files = ref<FileItem[]>([])
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref<FileDetailResponse | null>(null)
const previewUrls = ref<Record<number, string>>({})

const pagination = reactive({
  page: 1,
  per_page: 15,
  total: 0,
  last_page: 0,
})

const queryForm = reactive({
  keyword: '',
  mime_type: '',
  uploader_id: '',
  date_range: null as [number, number] | null,
})

const uploadInput = ref<HTMLInputElement | null>(null)

const canViewFiles = computed(() => permissionStore.hasPermission('page.file-management'))
const canUploadFiles = computed(() => permissionStore.hasPermission('file:upload'))
const canDeleteFiles = computed(() => permissionStore.hasPermission('file:delete'))

const hasFiles = computed(() => files.value.length > 0)
const detailReferences = computed(() => detail.value?.references || [])
const detailPreviewUrl = computed(() => {
  if (!detail.value || !isImageMime(detail.value.mime_type)) return ''
  return previewUrls.value[detail.value.id] || ''
})

function tsToDate(ts: number | null | undefined) {
  if (!ts) return undefined
  const d = new Date(ts)
  const yyyy = d.getFullYear()
  const mm = `${d.getMonth() + 1}`.padStart(2, '0')
  const dd = `${d.getDate()}`.padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

function buildQuery(): FileListQuery {
  const range = queryForm.date_range
  return {
    page: pagination.page,
    per_page: pagination.per_page,
    keyword: queryForm.keyword.trim() || undefined,
    mime_type: queryForm.mime_type.trim() || undefined,
    uploader_id: queryForm.uploader_id ? Number(queryForm.uploader_id) : undefined,
    date_from: tsToDate(range?.[0]),
    date_to: tsToDate(range?.[1]),
  }
}

async function loadFiles(options: { refresh?: boolean } = {}) {
  if (!canViewFiles.value) {
    files.value = []
    errorMessage.value = ''
    clearPreviewUrls()
    return
  }

  loading.value = !options.refresh
  refreshing.value = Boolean(options.refresh)
  errorMessage.value = ''
  try {
    clearPreviewUrls()
    const result = await getFiles(buildQuery())
    files.value = result.list
    pagination.total = result.total
    pagination.page = result.page
    pagination.per_page = result.per_page
    pagination.last_page = result.last_page
    await loadThumbnails(result.list)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载失败'
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  void loadFiles()
}

function handleReset() {
  queryForm.keyword = ''
  queryForm.mime_type = ''
  queryForm.uploader_id = ''
  queryForm.date_range = null
  pagination.page = 1
  void loadFiles()
}

function handlePageChange(page: number) {
  pagination.page = page
  void loadFiles()
}

function handlePageSizeChange(perPage: number) {
  pagination.per_page = perPage
  pagination.page = 1
  void loadFiles()
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} B`
  const units = ['KB', 'MB', 'GB', 'TB']
  let size = value / 1024
  let index = 0
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index += 1
  }
  return `${size.toFixed(size >= 10 ? 0 : 1)} ${units[index]}`
}

function formatDateTime(value: string | null | undefined) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(date)
}

function uploaderLabel(item: FileItem) {
  if (!item.uploader) return '-'
  return item.uploader.display_name || item.uploader.username
}

function uploaderMeta(item: FileItem) {
  if (!item.uploader) return '-'
  return item.uploader.username
}

function openUploadPicker() {
  uploadInput.value?.click()
}

async function handleUpload(event: Event) {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0]
  if (!file) return

  if (file.size <= 0) {
    message.error('文件内容不能为空')
    if (input) input.value = ''
    return
  }

  uploading.value = true
  try {
    await uploadFile(file)
    message.success('上传成功')
    pagination.page = 1
    await loadFiles()
  } catch (error) {
    message.error(error instanceof Error ? error.message : '上传失败')
  } finally {
    uploading.value = false
    if (input) input.value = ''
  }
}

async function handleDelete(item: FileItem | FileDetailResponse) {
  try {
    await confirm({
      title: '删除文件',
      content: `确认删除文件「${item.original_name}」？删除后仅做软删除处理。`,
      positiveText: '删除',
      negativeText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }

  deletingId.value = item.id
  try {
    await deleteFile(item.id)
    message.success('删除成功')
    await loadFiles({ refresh: true })
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除失败')
  } finally {
    deletingId.value = null
  }
}

async function openDetail(item: FileItem | FileDetailResponse) {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  detail.value = null
  try {
    detail.value = await getFileDetail(item.id)
    await loadDetailThumbnail(detail.value)
  } catch (error) {
    detailError.value = error instanceof Error ? error.message : '加载详情失败'
  } finally {
    detailLoading.value = false
  }
}

function formatReferenceLabel(item: FileDetailResponse['references'][number]) {
  const name = item.ref_name || item.ref_id
  return `${item.ref_type} · ${name}`
}

function openDownload(item: FileItem | FileDetailResponse) {
  void triggerDownload(item.id, item.original_name)
}

async function triggerDownload(id: number, filename: string) {
  try {
    const blob = await downloadFile(id)
    const url = window.URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = filename
    anchor.click()
    window.URL.revokeObjectURL(url)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '下载失败')
  }
}

function isImageMime(value: string) {
  return value.startsWith('image/')
}

function clearPreviewUrls() {
  for (const url of Object.values(previewUrls.value)) {
    window.URL.revokeObjectURL(url)
  }
  previewUrls.value = {}
}

async function loadThumbnails(items: FileItem[]) {
  const previewItems = items.filter((item) => isImageMime(item.mime_type))
  await Promise.all(
    previewItems.map(async (item) => {
      try {
        const blob = await downloadFile(item.id)
        const url = window.URL.createObjectURL(blob)
        previewUrls.value = { ...previewUrls.value, [item.id]: url }
      } catch {
        // ignore
      }
    }),
  )
}

async function loadDetailThumbnail(item: FileDetailResponse) {
  if (!isImageMime(item.mime_type) || previewUrls.value[item.id]) return
  try {
    const blob = await downloadFile(item.id)
    const url = window.URL.createObjectURL(blob)
    previewUrls.value = { ...previewUrls.value, [item.id]: url }
  } catch {
    // ignore
  }
}

const columns = computed<DataTableColumns<FileItem>>(() => [
  {
    key: 'file',
    title: '文件',
    minWidth: 260,
    render: (row) =>
      h('div', { class: 'file-management-page__file' }, [
        h(
          'div',
          { class: 'file-management-page__thumb' },
          previewUrls.value[row.id]
            ? h(NImage, {
                src: previewUrls.value[row.id],
                previewedImgProps: { style: 'max-width: 90vw' },
                objectFit: 'cover',
                width: 44,
                height: 44,
              })
            : h(
                NIcon,
                { size: 22, class: 'file-management-page__file-icon' },
                { default: () => h(DocumentOutline) },
              ),
        ),
        h('div', null, [
          h('div', { class: 'file-management-page__primary' }, row.original_name),
          h('div', { class: 'file-management-page__secondary' }, row.extension || '-'),
        ]),
      ]),
  },
  { key: 'mime_type', title: '类型', minWidth: 180, ellipsis: { tooltip: true } },
  {
    key: 'size',
    title: '大小',
    minWidth: 110,
    render: (row) => formatBytes(row.size),
  },
  {
    key: 'uploader',
    title: '上传者',
    minWidth: 180,
    render: (row) =>
      h('div', { class: 'file-management-page__identity' }, [
        h('span', { class: 'file-management-page__primary' }, uploaderLabel(row)),
        h('span', { class: 'file-management-page__secondary' }, uploaderMeta(row)),
      ]),
  },
  {
    key: 'created_at',
    title: '创建时间',
    minWidth: 180,
    render: (row) => formatDateTime(row.created_at),
  },
  {
    key: 'actions',
    title: '操作',
    width: 200,
    align: 'center',
    render: (row) => {
      const buttons = [
        h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row) }, { default: () => '详情' }),
        h(NButton, { text: true, type: 'success', onClick: () => openDownload(row) }, { default: () => '下载' }),
      ]
      if (canDeleteFiles.value) {
        buttons.push(
          h(
            NButton,
            { text: true, type: 'error', loading: deletingId.value === row.id, onClick: () => handleDelete(row) },
            { default: () => '删除' },
          ),
        )
      }
      return h(NSpace, { size: 8 }, { default: () => buttons })
    },
  },
])

onMounted(() => {
  void loadFiles()
})

onBeforeUnmount(() => {
  clearPreviewUrls()
})
</script>

<template>
  <div class="file-management-page">
    <div class="file-management-page__header">
      <div>
        <h2>附件管理</h2>
        <p>统一管理图片和附件上传记录，支持列表筛选、审计追踪和软删除。</p>
      </div>
      <div class="file-management-page__actions">
        <input ref="uploadInput" class="file-management-page__file-input" type="file" @change="handleUpload" />
        <NButton v-if="canUploadFiles" type="primary" :loading="uploading" @click="openUploadPicker">
          <template #icon>
            <NIcon><CloudUploadOutline /></NIcon>
          </template>
          上传文件
        </NButton>
        <NButton :loading="refreshing" @click="loadFiles({ refresh: true })">
          <template #icon>
            <NIcon><RefreshOutline /></NIcon>
          </template>
          刷新
        </NButton>
      </div>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadFiles">
      <template v-if="!canViewFiles">
        <NCard>
          <EmptyState title="暂无权限" description="当前账号没有附件管理访问权限。" />
        </NCard>
      </template>

      <template v-else>
        <NCard :bordered="false" class="file-management-page__filters-card">
          <NForm inline label-placement="left" class="file-management-page__filters" @submit.prevent>
            <NFormItem label="关键词">
              <NInput v-model:value="queryForm.keyword" clearable placeholder="文件名关键词" @keyup.enter="handleSearch" />
            </NFormItem>
            <NFormItem label="类型">
              <NInput v-model:value="queryForm.mime_type" clearable placeholder="image/jpeg" @keyup.enter="handleSearch" />
            </NFormItem>
            <NFormItem label="上传者 ID">
              <NInput v-model:value="queryForm.uploader_id" clearable placeholder="例如 1" @keyup.enter="handleSearch" />
            </NFormItem>
            <NFormItem label="时间">
              <NDatePicker v-model:value="queryForm.date_range" type="daterange" clearable />
            </NFormItem>
            <NFormItem :show-label="false">
              <NSpace>
                <NButton type="primary" @click="handleSearch">
                  <template #icon>
                    <NIcon><SearchOutline /></NIcon>
                  </template>
                  查询
                </NButton>
                <NButton @click="handleReset">重置</NButton>
              </NSpace>
            </NFormItem>
          </NForm>
        </NCard>

        <NCard :bordered="false">
          <template v-if="hasFiles">
            <NDataTable
              :columns="columns"
              :data="files"
              :row-key="(row: FileItem) => row.id"
              striped
              :bordered="false"
            />
            <div class="file-management-page__pagination">
              <NPagination
                :page="pagination.page"
                :page-size="pagination.per_page"
                :item-count="pagination.total"
                :page-sizes="[15, 30, 50, 100]"
                show-size-picker
                show-quick-jumper
                @update:page="handlePageChange"
                @update:page-size="handlePageSizeChange"
              />
            </div>
          </template>
          <template v-else>
            <EmptyState title="暂无文件" description="当前还没有上传任何文件。" />
          </template>
        </NCard>
      </template>
    </QueryState>

    <NModal
      v-model:show="detailVisible"
      preset="card"
      title="文件详情"
      style="width: 640px"
    >
      <QueryState :loading="detailLoading" :error-message="detailError" @retry="() => detail && openDetail(detail)">
        <template v-if="detail">
          <div v-if="detailPreviewUrl" class="file-management-page__detail-preview">
            <NImage :src="detailPreviewUrl" object-fit="contain" />
          </div>

          <NDescriptions :column="1" bordered label-placement="left" size="small">
            <NDescriptionsItem label="文件名">{{ detail.original_name }}</NDescriptionsItem>
            <NDescriptionsItem label="类型">{{ detail.mime_type }}</NDescriptionsItem>
            <NDescriptionsItem label="大小">{{ formatBytes(detail.size) }}</NDescriptionsItem>
            <NDescriptionsItem label="上传者">
              {{ detail.uploader?.display_name || detail.uploader?.username || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem label="校验和">{{ detail.checksum }}</NDescriptionsItem>
            <NDescriptionsItem label="存储驱动">{{ detail.storage_driver }}</NDescriptionsItem>
            <NDescriptionsItem label="引用数量">{{ detail.reference_count }}</NDescriptionsItem>
            <NDescriptionsItem label="创建时间">{{ formatDateTime(detail.created_at) }}</NDescriptionsItem>
          </NDescriptions>

          <NDivider />

          <div class="file-management-page__drawer-actions">
            <NButton type="primary" @click="openDownload(detail)">
              <template #icon>
                <NIcon><DownloadOutline /></NIcon>
              </template>
              下载/预览
            </NButton>
            <NButton
              v-if="canDeleteFiles && detail.can_delete"
              type="error"
              :loading="deletingId === detail.id"
              @click="handleDelete(detail)"
            >
              <template #icon>
                <NIcon><TrashOutline /></NIcon>
              </template>
              删除文件
            </NButton>
          </div>

          <NDivider title-placement="left">引用关系</NDivider>
          <template v-if="detailReferences.length > 0">
            <NTimeline>
              <NTimelineItem
                v-for="item in detailReferences"
                :key="item.id"
                :time="formatDateTime(item.created_at)"
              >
                <div class="file-management-page__reference-item">
                  <div class="file-management-page__primary">{{ formatReferenceLabel(item) }}</div>
                  <div class="file-management-page__secondary">{{ item.ref_path || '无位置描述' }}</div>
                </div>
              </NTimelineItem>
            </NTimeline>
          </template>
          <template v-else>
            <EmptyState title="暂无引用" description="该文件目前没有被业务记录引用。" />
          </template>
        </template>
      </QueryState>
    </NModal>
  </div>
</template>

<style scoped>
.file-management-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.file-management-page__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.file-management-page__header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.file-management-page__header p {
  margin: 6px 0 0;
  color: rgba(15, 23, 42, 0.55);
}

.file-management-page__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-management-page__file-input {
  display: none;
}

.file-management-page__filters-card {
  margin-bottom: 0;
}

.file-management-page__filters {
  margin-bottom: -8px;
}

.file-management-page__file,
.file-management-page__identity {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-management-page__file {
  flex-direction: row;
  align-items: center;
  gap: 12px;
}

.file-management-page__thumb {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  overflow: hidden;
  flex-shrink: 0;
  background: rgba(37, 99, 235, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
}

.file-management-page__file-icon {
  color: #2563eb;
}

.file-management-page__primary {
  font-weight: 600;
  color: rgba(15, 23, 42, 0.92);
}

.file-management-page__secondary {
  color: rgba(15, 23, 42, 0.55);
  font-size: 12px;
}

.file-management-page__pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.file-management-page__drawer-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.file-management-page__detail-preview {
  width: 100%;
  margin-bottom: 16px;
  border-radius: 12px;
  overflow: hidden;
  background: rgba(37, 99, 235, 0.06);
  display: flex;
  justify-content: center;
}

.file-management-page__reference-item {
  display: grid;
  gap: 4px;
}
</style>
