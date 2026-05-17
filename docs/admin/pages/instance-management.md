# Instance Management 页面契约

实例管理用于维护交付映射、查看云主机实例、触发实例操作和同步 MCP PVE client API 状态，页面只调用 `/admin-api/*`。

## 页面范围

- 实例列表
- 实例详情
- 交付映射列表和维护
- MCP 节点、节点详情、节点 VM 列表和存储只读查看
- 从订单触发交付后的实例状态排障
- 开机、关机、释放和同步

本页面不开放通用 PVE 运维管理，不提供重启、重装、重置密码、控制台、快照、备份、迁移、监控、网络防火墙或资源池管理。

## 路由与权限

- 路由：`/instances`
- 菜单权限：`page.instances`
- 查看：`instance:view` 或 `instance:*`
- 触发交付和维护交付映射：`instance:provision` 或 `instance:*`
- 开机、关机：`instance:operate` 或 `instance:*`
- 释放：`instance:release` 或 `instance:*`
- 同步：`instance:sync` 或 `instance:*`

## 页面结构

实例管理是复杂管理页，应使用页面容器结构：

```text
admin/src/views/instances/
  index.vue
  types.ts
  components/
    InstancesTab.vue
    ProvisionMappingsTab.vue
    McpResourcesTab.vue
```

## 行为约束

- 列表支持按实例状态、实例编号、订单编号、用户关键字和创建时间范围筛选。
- 实例详情可展示管理端可见的 MCP `node`、`vmid`、最近 operation 和失败原因。
- 用户端不可见的 `node`、`storage`、`disk_source`、`snippets_storage`、`vmid` 和上游 operation ID 不得出现在用户端接口或用户端页面。
- 交付映射维护 MCP 创建 VM 所需参数和 VMID 分配范围；`ci_password` 当前不作为配置项维护。
- MCP 节点、存储和节点 VM 列表仅用于配置映射和排障，不作为资源池管理页面。
- 开机、关机、释放和同步必须以服务端返回状态为准，前端只做按钮可见性和二次确认。

## 关联接口

接口字段和响应结构以 `docs/server/api/endpoints.md` 为准。

- `GET /admin-api/instance-provision-mappings`
- `POST /admin-api/instance-provision-mappings`
- `PATCH /admin-api/instance-provision-mappings/{id}`
- `GET /admin-api/mcp-pve/nodes`
- `GET /admin-api/mcp-pve/nodes/{node}`
- `GET /admin-api/mcp-pve/nodes/{node}/vms`
- `GET /admin-api/mcp-pve/storage`
- `GET /admin-api/instances`
- `GET /admin-api/instances/{instance_no}`
- `POST /admin-api/instances/{instance_no}/start`
- `POST /admin-api/instances/{instance_no}/stop`
- `POST /admin-api/instances/{instance_no}/release`
- `POST /admin-api/instances/{instance_no}/sync`
- `POST /admin-api/orders/{order_no}/provision`

## 验收重点

- 侧栏菜单来自后端 `menus`，本地路由只作为页面组件白名单。
- 无权限访问 `/instances` 时展示管理端 403 反馈。
- 低权限管理员看不到或无法触发无权限操作按钮。
- 交付映射、实例列表、实例详情和 MCP 只读资源查询正常。
- 页面不出现 MCP 未提供能力的入口。
