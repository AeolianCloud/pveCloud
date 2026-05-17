import type { InstanceMappingPayload, InstanceStatus, MappingStatus } from '../../api/instance'

export type InstanceTabKey = 'instances' | 'mappings' | 'mcp'
export type MappingDialogMode = 'create' | 'edit'

export const instanceStatusText: Record<InstanceStatus, string> = {
  creating: '创建中',
  running: '运行中',
  stopped: '已停止',
  error: '异常',
  releasing: '释放中',
  released: '已释放',
}

export const operationActionText: Record<string, string> = {
  provision: '交付',
  start: '开机',
  stop: '关机',
  release: '释放',
  sync: '同步',
}

export const operationStatusText: Record<string, string> = {
  running: '执行中',
  succeeded: '成功',
  failed: '失败',
}

export const mappingStatusText: Record<MappingStatus, string> = {
  active: '启用',
  inactive: '停用',
}

export function makeEmptyMappingForm(): InstanceMappingPayload {
  return {
    mapping_no: '',
    product_no: null,
    plan_no: '',
    region_no: '',
    template_no: '',
    network_type_no: '',
    node: '',
    storage: '',
    disk_source: '',
    disk_format: null,
    disk_interface: null,
    snippets_storage: null,
    ci_user: null,
    ssh_keys: null,
    ip_config0: null,
    nameserver: null,
    search_domain: null,
    ci_packages: null,
    apt_mirror: null,
    vmid_start: 100,
    vmid_end: 999,
    next_vmid: 100,
    status: 'active',
    remark: null,
  }
}
