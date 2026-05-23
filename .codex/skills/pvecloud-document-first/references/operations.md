# Operations Guardrails

本文件定义配置、部署和运维相关实现守则。

## 先读什么

- `docs/development/local-setup.md`
- `docs/operations/deployment.md`
- `docs/security.md`
- `server/config.example.yaml`

## 契约边界

- 配置项说明和默认语义以 `server/config.example.yaml` 为准。
- 本地开发流程写在 `docs/development/`。
- 部署拓扑、代理边界、依赖要求写在 `docs/operations/`。
- 技术栈和运行时选择以 `docs/server/go-technical.md`、`docs/admin/architecture.md`、`docs/web/architecture.md` 和各自 `package.json` 为准；不要擅自替换为第二套构建器、代理或进程管理方案。

## 守则

- 未经维护者明确要求，不提交真实配置和密钥。
- 新增配置项时，先更新 `server/config.example.yaml`，再改代码和文档。
- 不把生产依赖写成“可有可无”的降级选项，除非文档明确允许。
- Redis、MariaDB、API、Worker、前端开发服务的启动顺序和依赖关系要写清楚。
- 新增 `web/` 或其他前端应用时，必须同步检查本地开发说明、部署说明和启动脚本是否覆盖它。
- 只要任务影响代理边界、启动顺序、前端入口或运行依赖，就要联动 `operations` 文档。
- 只要任务影响真实密钥、token、cookie、CORS、代理边界、日志输出、备份输出或部署暴露面，就要同步检查 `docs/security.md`。

## 真实配置与密钥

- `server/config.yaml` 是真实配置；未经维护者明确要求不得提交，维护者明确要求提交时仍不得展示或复制其中的 secret、token、password、SMTP 凭据、数据库密码、Redis 密码、对象存储密钥和代理私有信息。
- 如果 `server/config.yaml` 出现在 `git status --short`，先把它视为用户本地改动；不要为了完成任务回退、覆盖或格式化它。
- 如果任务必须理解配置形状，优先读取 `server/config.example.yaml`、配置加载代码和运维文档；不要把真实配置内容当成文档来源。
- 如果真实密钥已经进入日志、示例输出或对话内容，停止扩散，提示维护者清理并轮换相关凭据；不要把密钥粘贴进回复。维护者明确要求提交真实配置时，在提交说明中标注包含真实配置变更。

## 新增环境依赖

新增或改变环境依赖时，同步检查：

- `server/config.example.yaml` 是否能表达必要配置。
- `docs/development/local-setup.md` 是否写清本地安装、启动和验证方式。
- `docs/operations/deployment.md` 是否写清生产依赖、启动顺序、端口、代理和持久化要求。
- 现有开发脚本、构建脚本、服务启动命令和 README 索引是否仍准确。
- 如果依赖影响数据库、缓存、存储、邮件、前端构建或反向代理，联动对应 owner docs。

## 依赖验收口径

- SMTP：验证配置来源、发信调用路径、失败日志脱敏、超时或重试边界；不得在日志或测试输出中暴露邮箱密码、授权码或 token。
- Redis：验证连接、TTL、缓存/短锁/限流用途和失败处理；不得把 Redis 当成最终业务事实来源。
- MariaDB：验证迁移、连接、事务边界、字符集/排序规则和幂等初始化；不得用本地临时表结构替代迁移。
- Storage：验证本地或对象存储的配置入口、路径/桶名语义、权限边界、清理策略和失败后的数据库一致性。
- Proxy：验证 `/api/*`、`/admin-api/*`、前端静态资源、WebSocket 或长连接边界是否和部署文档一致。

## 开发脚本与生产部署边界

- 开发脚本只服务本地启动、热更新、构建和测试，不作为生产部署事实来源。
- 生产部署以 `docs/operations/`、服务二进制、构建产物、配置示例和反向代理说明为准。
- 不把 `bun run dev`、前端开发服务器、临时代理或本机路径写成生产部署方式。
- 修改端口、域名、静态资源路径、构建输出目录或反向代理规则时，必须同步部署文档和相关脚本。

## 部署检查清单

涉及部署或前端入口时，至少检查：

- 端口是否和本地开发、后端监听、前端开发服务、生产反向代理说明一致。
- 域名、站点根路径、API base URL、Cookie domain/path 和跨域策略是否有 owner docs 支撑。
- 反向代理是否区分 `/api/*`、`/admin-api/*`、管理端静态资源和用户端静态资源。
- `admin/` 与 `web/` 的构建输出目录、部署路径和缓存策略是否写清。
- 新增静态资源、上传目录或对象存储路径是否有持久化、权限和备份口径。
- CORS 是否使用明确允许来源，不能宽泛暴露管理端接口或带凭证请求。
- HSTS、CSP、`frame-ancestors` 或 `X-Frame-Options`、`X-Content-Type-Options`、`Referrer-Policy`、`Cache-Control` 是否有部署口径。
- 受保护 API、实名资料、审计详情、敏感配置和文件下载是否避免被共享缓存保存。
- 备份、日志、导出文件、对象存储桶和构建产物是否避免公开暴露真实密钥和敏感数据。
- Go/Bun 依赖、lockfile、安装脚本、构建插件和远程脚本来源是否可审查，新增依赖是否需要供应链风险说明。

## 验证

- 新配置项可被示例配置表达。
- 本地开发说明与实际命令一致。
- 部署说明与代理边界、运行依赖一致。
