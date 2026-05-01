import { http, type ApiEnvelope } from '../utils/request'

import type { PaginatedData } from './admin-user'

export interface FileUploaderSummary {
  id: number
  username: string
  display_name: string
}

export interface FileItem {
  id: number
  original_name: string
  mime_type: string
  extension: string
  size: number
  url: string
  uploader: FileUploaderSummary | null
  created_at: string
}

export interface FileReferenceItem {
  id: number
  file_id: number
  ref_type: string
  ref_id: string
  ref_name: string | null
  ref_path: string | null
  created_at: string
}

export interface FileReferenceResponse {
  file_id: number
  reference_count: number
  references: FileReferenceItem[]
}

export interface FileDetailResponse extends FileItem {
  storage_driver: string
  checksum: string
  reference_count: number
  references: FileReferenceItem[]
  download_url: string
  can_delete: boolean
}

export interface FileUploadResponse {
  id: number
  original_name: string
  mime_type: string
  size: number
  url: string
  created_at: string
}

export interface FileListQuery {
  page?: number
  per_page?: number
  keyword?: string
  mime_type?: string
  uploader_id?: number
  date_from?: string
  date_to?: string
}

export async function getFiles(query?: FileListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<FileItem>>>('/files', { params: query })
  return response.data.data
}

export async function getFileDetail(id: number) {
  const response = await http.get<ApiEnvelope<FileDetailResponse>>(`/files/${id}`)
  return response.data.data
}

export async function getFileReferences(id: number) {
  const response = await http.get<ApiEnvelope<FileReferenceResponse>>(`/files/${id}/references`)
  return response.data.data
}

export async function uploadFile(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  const response = await http.post<ApiEnvelope<FileUploadResponse>>('/files/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
  return response.data.data
}

export async function deleteFile(id: number) {
  await http.delete<ApiEnvelope<null>>(`/files/${id}`)
}

export function getFileDownloadUrl(id: number) {
  return `/admin-api/files/${id}/download`
}

export async function downloadFile(id: number) {
  const response = await http.get<Blob>(`/files/${id}/download`, { responseType: 'blob' })
  return response.data
}
