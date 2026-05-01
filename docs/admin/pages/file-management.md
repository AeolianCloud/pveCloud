# File Management 页面契约

## 页面定位

`File Management` 是管理端的附件与文件资源管理页面。

当前承载：

- 图片/附件上传
- 文件列表查询
- 文件详情查看
- 文件下载与预览
- 文件软删除
- 上传审计入口

当前不承载：

- 公网直接文件下载
- 用户端文件资源管理
- 文件类型配置编辑
- 存储驱动配置编辑

## 路由结构

- 路径：`/files`
- 标题：附件管理
- 作为受保护页面挂在后台壳下
- 侧栏菜单使用独立一级入口，不挂入 `System Settings`

## 页面职责

- 以表格展示文件记录。
- 提供上传按钮和删除操作。
- 提供关键词、类型、上传者、时间范围等筛选。
- 展示文件名、类型、大小、上传者、创建时间和状态。
- 详情抽屉展示引用关系、校验和、存储驱动等信息，但不暴露物理路径。
- 上传后刷新列表并提示结果。
- 删除前必须二次确认。
- 列表和详情不暴露物理存储路径。

## 权限建议

- 页面入口：`page.file-management`
- 上传：`file:upload` 或 `file:*`
- 删除：`file:delete` 或 `file:*`
- 列表查看：`page.file-management`
- 下载/预览：`page.file-management`

## 关联接口

- `POST /admin-api/files/upload`
- `GET /admin-api/files`
- `GET /admin-api/files/{id}`
- `GET /admin-api/files/{id}/download`
- `GET /admin-api/files/{id}/references`
- `DELETE /admin-api/files/{id}`

## 验收重点

- 文件管理必须只调用 `/admin-api/*`。
- 路由、侧栏、按钮权限都通过统一权限能力判断。
- 页面不显示本地磁盘路径和敏感存储细节。
- 文件详情抽屉必须展示引用数量和引用摘要，删除前需提示是否仍被引用。
- 上传时要展示校验失败的明确原因。
- 删除时要保留二次确认。
- 列表分页和筛选应与后端接口参数一致。
