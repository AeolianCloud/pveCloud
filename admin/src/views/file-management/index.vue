<script setup lang="ts">
import { Delete, Download, Document, Refresh, Search, Upload } from '@element-plus/icons-vue'
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

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
  date_range: [] as string[],
})

const uploadInput = ref<HTMLInputElement | null>(null)

const canViewFiles = computed(() => permissionStore.hasPermission('page.file-management'))
const canUploadFiles = computed(() => permissionStore.hasPermission('file:upload'))
const canDeleteFiles = computed(() => permissionStore.hasPermission('file:delete'))

const hasFiles = computed(() => files.value.length > 0)
const detailReferences = computed(() => detail.value?.references || [])
const detailPreviewUrl = computed(() => {
  if (!detail.value || !isImageMime(detail.value.mime_type)) {
    return ''
  }
  return previewUrls.value[detail.value.id] || ''
})

function buildQuery(): FileListQuery {
  const [dateFrom, dateTo] = queryForm.date_range
  return {
    page: pagination.page,
    per_page: pagination.per_page,
    keyword: queryForm.keyword.trim() || undefined,
    mime_type: queryForm.mime_type.trim() || undefined,
    uploader_id: queryForm.uploader_id ? Number(queryForm.uploader_id) : undefined,
    date_from: dateFrom || undefined,
    date_to: dateTo || undefined,
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
  queryForm.date_range = []
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
    ElMessage.error('文件内容不能为空')
    if (input) input.value = ''
    return
  }

  uploading.value = true
  try {
    await uploadFile(file)
    ElMessage.success('上传成功')
    pagination.page = 1
    await loadFiles()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '上传失败')
  } finally {
    uploading.value = false
    if (input) input.value = ''
  }
}

async function handleDelete(item: FileItem) {
  try {
    await ElMessageBox.confirm(
      `确认删除文件「${item.original_name}」？删除后仅做软删除处理。`,
      '删除文件',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )
  } catch {
    return
  }

  deletingId.value = item.id
  try {
    await deleteFile(item.id)
    ElMessage.success('删除成功')
    await loadFiles({ refresh: true })
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '删除失败')
  } finally {
    deletingId.value = null
  }
}

async function openDetail(item: FileItem) {
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

function openDownload(item: FileItem) {
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
    ElMessage.error(error instanceof Error ? error.message : '下载失败')
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
  await Promise.all(previewItems.map(async (item) => {
    try {
      const blob = await downloadFile(item.id)
      const url = window.URL.createObjectURL(blob)
      previewUrls.value = { ...previewUrls.value, [item.id]: url }
    } catch {
      // 缩略图失败不影响列表显示。
    }
  }))
}

async function loadDetailThumbnail(item: FileDetailResponse) {
  if (!isImageMime(item.mime_type) || previewUrls.value[item.id]) {
    return
  }
  try {
    const blob = await downloadFile(item.id)
    const url = window.URL.createObjectURL(blob)
    previewUrls.value = { ...previewUrls.value, [item.id]: url }
  } catch {
    // 详情缩略图失败不阻断页面。
  }
}

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
        <el-button v-if="canUploadFiles" :icon="Upload" :loading="uploading" type="primary" @click="openUploadPicker">
          上传文件
        </el-button>
        <el-button :icon="Refresh" :loading="refreshing" @click="loadFiles({ refresh: true })">刷新</el-button>
      </div>
    </div>

    <QueryState :loading="loading" :error-message="errorMessage" @retry="loadFiles">
      <template v-if="!canViewFiles">
        <el-card>
          <EmptyState title="暂无权限" description="当前账号没有附件管理访问权限。" />
        </el-card>
      </template>

      <template v-else>
        <el-card shadow="never" class="file-management-page__filters-card">
          <el-form inline class="file-management-page__filters" @submit.prevent>
            <el-form-item label="关键词">
              <el-input v-model="queryForm.keyword" clearable placeholder="文件名关键词" @keyup.enter="handleSearch" />
            </el-form-item>
            <el-form-item label="类型">
              <el-input v-model="queryForm.mime_type" clearable placeholder="image/jpeg" @keyup.enter="handleSearch" />
            </el-form-item>
            <el-form-item label="上传者 ID">
              <el-input v-model="queryForm.uploader_id" clearable placeholder="例如 1" @keyup.enter="handleSearch" />
            </el-form-item>
            <el-form-item label="时间">
              <el-date-picker
                v-model="queryForm.date_range"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>
              <el-button @click="handleReset">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="never">
          <template v-if="hasFiles">
            <el-table :data="files" stripe class="file-management-page__table">
              <el-table-column label="文件" min-width="260">
                <template #default="{ row }">
                  <div class="file-management-page__file">
                    <div class="file-management-page__thumb">
                      <el-image
                        v-if="previewUrls[row.id]"
                        :src="previewUrls[row.id]"
                        :preview-src-list="[previewUrls[row.id]]"
                        fit="cover"
                        hide-on-click-modal
                        preview-teleported
                      />
                      <el-icon v-else class="file-management-page__file-icon"><Document /></el-icon>
                    </div>
                    <div>
                      <div class="file-management-page__primary">{{ row.original_name }}</div>
                      <div class="file-management-page__secondary">{{ row.extension || '-' }}</div>
                    </div>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="类型" prop="mime_type" min-width="180" show-overflow-tooltip />
              <el-table-column label="大小" min-width="110">
                <template #default="{ row }">{{ formatBytes(row.size) }}</template>
              </el-table-column>
              <el-table-column label="上传者" min-width="180">
                <template #default="{ row }">
                  <div class="file-management-page__identity">
                    <span class="file-management-page__primary">{{ uploaderLabel(row) }}</span>
                    <span class="file-management-page__secondary">{{ uploaderMeta(row) }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="创建时间" min-width="180">
                <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
              </el-table-column>
              <el-table-column label="操作" width="120" align="center">
                <template #default="{ row }">
                  <div class="file-management-page__row-actions">
                    <el-button link type="primary" @click="openDetail(row)">详情</el-button>
                    <el-button link type="success" :icon="Download" @click="openDownload(row)">下载</el-button>
                    <el-button
                      v-if="canDeleteFiles"
                      :icon="Delete"
                      :loading="deletingId === row.id"
                      link
                      type="danger"
                      @click="handleDelete(row)"
                    >
                      删除
                    </el-button>
                  </div>
                </template>
              </el-table-column>
            </el-table>

            <div class="file-management-page__pagination">
              <el-pagination
                :current-page="pagination.page"
                :page-size="pagination.per_page"
                :page-sizes="[15, 30, 50, 100]"
                :total="pagination.total"
                layout="total, sizes, prev, pager, next, jumper"
                background
                @current-change="handlePageChange"
                @size-change="handlePageSizeChange"
              />
            </div>
          </template>

          <template v-else>
            <EmptyState title="暂无文件" description="当前还没有上传任何文件。" />
          </template>
        </el-card>
      </template>
    </QueryState>

    <el-dialog v-model="detailVisible" width="640px" title="文件详情" destroy-on-close>
      <QueryState :loading="detailLoading" :error-message="detailError" @retry="() => detail && openDetail(detail)">
        <template v-if="detail">
          <div class="file-management-page__detail-preview" v-if="detailPreviewUrl">
            <el-image
              :src="detailPreviewUrl"
              :preview-src-list="[detailPreviewUrl]"
              fit="contain"
              hide-on-click-modal
              preview-teleported
            />
          </div>

          <el-descriptions :column="1" border>
            <el-descriptions-item label="文件名">{{ detail.original_name }}</el-descriptions-item>
            <el-descriptions-item label="类型">{{ detail.mime_type }}</el-descriptions-item>
            <el-descriptions-item label="大小">{{ formatBytes(detail.size) }}</el-descriptions-item>
            <el-descriptions-item label="上传者">{{ detail.uploader?.display_name || detail.uploader?.username || '-' }}</el-descriptions-item>
            <el-descriptions-item label="校验和">{{ detail.checksum }}</el-descriptions-item>
            <el-descriptions-item label="存储驱动">{{ detail.storage_driver }}</el-descriptions-item>
            <el-descriptions-item label="引用数量">{{ detail.reference_count }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatDateTime(detail.created_at) }}</el-descriptions-item>
          </el-descriptions>

          <el-divider />

          <div class="file-management-page__drawer-actions">
            <el-button type="primary" :icon="Download" @click="openDownload(detail)">下载/预览</el-button>
            <el-button v-if="canDeleteFiles && detail.can_delete" type="danger" :loading="deletingId === detail.id" @click="handleDelete(detail)">
              删除文件
            </el-button>
          </div>

          <el-divider content-position="left">引用关系</el-divider>
          <template v-if="detailReferences.length > 0">
            <el-timeline>
              <el-timeline-item v-for="item in detailReferences" :key="item.id" :timestamp="formatDateTime(item.created_at)">
                <div class="file-management-page__reference-item">
                  <div class="file-management-page__primary">{{ formatReferenceLabel(item) }}</div>
                  <div class="file-management-page__secondary">{{ item.ref_path || '无位置描述' }}</div>
                </div>
              </el-timeline-item>
            </el-timeline>
          </template>
          <template v-else>
            <EmptyState title="暂无引用" description="该文件目前没有被业务记录引用。" />
          </template>
        </template>
      </QueryState>
    </el-dialog>
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
  color: var(--el-text-color-secondary);
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

.file-management-page__table {
  margin-top: 4px;
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
  background: rgba(64, 158, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
}

.file-management-page__thumb :deep(.el-image),
.file-management-page__detail-preview :deep(.el-image) {
  width: 100%;
  height: 100%;
  display: block;
}

.file-management-page__thumb :deep(.el-image__inner),
.file-management-page__detail-preview :deep(.el-image__inner) {
  width: 100%;
  height: 100%;
}

.file-management-page__detail-preview :deep(.el-image) {
  width: 100%;
  height: 260px;
}

.file-management-page__file-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(64, 158, 255, 0.12);
  color: var(--el-color-primary);
}

.file-management-page__primary {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.file-management-page__secondary {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.file-management-page__pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.file-management-page__row-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  flex-wrap: wrap;
}

.file-management-page__drawer-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.file-management-page__detail-preview {
  width: 100%;
  max-height: 260px;
  margin-bottom: 16px;
  border-radius: 12px;
  overflow: hidden;
  background: rgba(64, 158, 255, 0.08);
}

.file-management-page__reference-item {
  display: grid;
  gap: 4px;
}
</style>
