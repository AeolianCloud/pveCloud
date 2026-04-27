# pveCloud — Claude Code 协作入口

本文件定义 Claude Code 在本仓库中的工作方法。不承载接口字段、表结构、页面契约或产品说明。

## 双轨协作原则

- **AI 轨道**：`CLAUDE.md` 和 `.claude/` 和 `AGENTS.md` — 约束 AI 的工作流、守则和验收方式
- **项目轨道**：`docs/`、`server/migrations/`、`server/config.example.yaml` — 承载事实、契约、架构、计划和进度

不要把两条轨道混写。

## 权威来源

- 接口契约：`docs/server/api/`
- 后端架构与业务规则：`docs/server/`
- 管理端前端契约：`docs/admin/`
- 用户端前端契约与规划：`docs/web/`
- 数据库可执行契约：`server/migrations/`
- 配置示例契约：`server/config.example.yaml`

如果本文件和项目文档冲突，以项目文档、迁移 SQL 和配置示例为准。

## 必读顺序

1. `CLAUDE.md`（本文件）
2. `docs/progress/MASTER.md`（当前进度）
3. 对应领域的项目文档（见下方）

按任务类型继续读取：

- 后端与 API：`docs/server/`、`docs/server/api/`
- 管理端前端：`docs/admin/`
- 用户端前端：`docs/web/`
- 数据库：`docs/server/database/`、`server/migrations/`
- 配置与部署：`docs/development/`、`docs/operations/`、`server/config.example.yaml`
- 计划与进度：`docs/analysis/`、`docs/plan/`、`docs/progress/`

## 文档先行门禁

契约或行为变更必须先更新文档或机器契约，再暂停等待确认：

- API、错误码、鉴权、权限码、菜单来源、路由语义
- 数据库结构、迁移、事务边界、幂等规则
- 页面行为、页面范围、状态语义、请求包装、权限判断
- 配置项、部署方式、运行依赖、运维流程
- 业务流程、阶段边界、产品开放范围

纯 UI/UX 视觉优化可直接实现，不触发确认门禁：

- 布局、间距、颜色、字体、图标、响应式样式
- 不影响流程和契约的静态文案调整

## 标准工作流

1. `git status --short`，确认工作区是否有未提交改动
2. 读取本文件和对应项目文档
3. 判断是契约/行为变更还是纯视觉优化
4. 契约/行为变更 → 先更新文档或机器契约 → 暂停等待确认
5. 纯视觉优化 → 直接实现
6. 只在与任务直接相关的目录内改动，避免顺手扩散
7. 用最小但有意义的验证命令收尾
8. 报告改动内容、验证结果和残留风险

## 仓库级硬规则

- 不覆盖用户未提交改动
- `admin/` 和 `web/` 独立管理，不创建跨前端共享运行时代码包
- 管理端只调用 `/admin-api/*`
- 用户端只调用 `/api/*`
- 如果仓库里尚不存在 `web/` 前端目录，`docs/web/` 视为规划与契约草案，不能假定用户端实现已经存在
- 不重建已移除的页面、路由或菜单，除非文档先更新且维护者确认

## 领域守则

### 后端

- API 契约来自 `docs/server/api/`，不要只改 handler 或 DTO
- 表结构契约最终来自 `server/migrations/`
- 配置项契约最终来自 `server/config.example.yaml`
- 服务负责业务规则，handler 负责请求解析、权限声明和响应
- 外部系统协议适配放 `integrations/`，业务裁决放 `services/`
- 不把 RBAC 最终授权逻辑下放到前端
- 不把长耗时外部调用放进长事务
- 幂等必须依赖业务唯一键、状态检查或任务键，不能只依赖前端防重复点击
- 使用明确业务域命名，避免 `common`、`helper`、`manager`、`base` 等泛名
- 审计域统一使用 `audit` 命名
- 验证：`cd server && gofmt -w . && go test ./...`

### 数据库

- 最终表结构以 `server/migrations/` 为准
- 新增表、字段、索引、约束时，迁移和设计文档一起更新
- 迁移必须可重复执行，并考虑已有数据
- 不把高频查询条件塞进 JSON 字段
- 状态值由应用层常量维护，不用数据库 enum
- 事务只包本地必须原子完成的步骤
- 外部系统调用放到事务外，通过锚点、补偿或重试恢复
- MariaDB 为关键业务事实唯一源，Redis 只做短 TTL 状态、限流、缓存、短锁和辅助幂等

### 前端

- `admin/` 和 `web/` 完全独立，不跨应用导入任何代码
- 管理端优先用 Element Plus，不建设替代 UI 框架的本地工具类体系
- 请求包装按业务域拆分
- 路由元信息承担标题、图标、权限、菜单可见性等页面契约
- 全局状态只承载跨页面 concern，不为单页过度建模
- 验证：`cd admin && bun run build`

### 运维

- 配置项说明和默认语义以 `server/config.example.yaml` 为准
- 新增配置项时，先更新示例配置，再改代码和文档
- 不提交真实配置和密钥

## Basic Admin 阶段范围

- 后端仍覆盖：admin auth、dashboard、RBAC、admin sessions、system configs、audit logs、risk logs
- 管理端前端当前范围缩小为：Login、Dashboard、403
- 后端能力存在 ≠ 前端页面应该存在，不要混淆
- 后端 RBAC 是最终授权权威，前端权限逻辑仅用于可用性

## 确认消息模板

契约/行为变更的文档更新后，用用户语言发送：

```
文档/契约已先更新。确认点：
- ...

请确认这些设计、契约或验收口径是否通过。你确认后我再进入实现。
```
