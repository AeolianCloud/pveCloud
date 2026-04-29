---
name: pvecloud-operations
description: Operations guardrails for pveCloud. Use when working on config, deployment, or infrastructure changes.
---

# Operations Guardrails

## 先读什么

- `docs/development/local-setup.md`
- `docs/operations/deployment.md`
- `server/config.example.yaml`

## 契约边界

- 配置项说明和默认语义以 `server/config.example.yaml` 为准。
- 本地开发流程写在 `docs/development/`。
- 部署拓扑、代理边界、依赖要求写在 `docs/operations/`。

## 守则

- 不提交真实配置和密钥。
- 新增配置项时，先更新 `server/config.example.yaml`，再改代码和文档。
- 不把生产依赖写成"可有可无"的降级选项，除非文档明确允许。
- Redis、MariaDB、API、Worker、前端开发服务的启动顺序和依赖关系要写清楚。

## 验证

- 新配置项可被示例配置表达。
- 本地开发说明与实际命令一致。
- 部署说明与代理边界、运行依赖一致。
