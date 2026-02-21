# AGENTS.md

# AI 助手工作规则

- 每次修改代码前，必须先读取目标文件，了解当前技术栈版本后再动手
- 生成代码后，推荐同类可改进的功能供参考，由人类决定是否需要
- 不需要自动编译，出错由人类反馈后再修
- 不需要自动编写测试，说明测试方法即可
- 写代码要做注释，注释要详细

# 前端设计规则

- 走扁平化风格：简约、不自定义颜色，使用框架默认色板
- 设计关键词：现代化、简洁、专业、科技感、高效、清晰、一致、优雅、流畅、直观
- 交互关键词：响应式、即时反馈、平滑过渡、直观操作、智能提示、友好错误处理
- 体验关键词：易用性、可访问性、可维护性、性能优化、操作便捷
- 技术关键词：组件化、模块化、类型安全、状态管理、数据缓存、实时更新、错误边界

# 八荣八耻

- 以瞎猜接口为耻，以认真查询为荣
- 以模糊执行为耻，以寻求确认为荣
- 以臆想业务为耻，以人类确认为荣
- 以创造接口为耻，以复用现有为荣
- 以一遍通过为耻，以回头校验为荣
- 以破坏架构为耻，以遵循规范为荣
- 以假装理解为耻，以诚实无知为荣
- 以盲目修改为耻，以谨慎重构为荣

# 奥卡姆剃刀原则

**如无必要，勿增实体。**

- 优先选择简单成熟的方案，不引入不必要的复杂性
- 从最简单可行的架构开始，根据实际需求演进
- 能用标准库解决的，不引入第三方库
- 每增加一个依赖/抽象层，先问"这真的必要吗？"

---

# pveCloud 项目开发规范

> 完整规范可通过 `/pvecloud-standards` skill 查询。

## 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| 后端语言 | Go | 1.23+ |
| Web 框架 | Gin | v1.11.0 |
| ORM | GORM + MySQL Driver | v1.25.12 |
| 配置 | Viper | v1.20.1 |
| 日志 | Zap | v1.27.1 |
| 认证 | golang-jwt/jwt/v5 | v5.2.1 |
| 热更新 | air | latest |

## 目录职责

```
backend/internal/
├── handler/   # 参数绑定 + 调用 service + 返回响应（禁止写业务逻辑）
├── service/   # 业务逻辑（禁止直接操作 gin.Context）
├── model/     # struct 定义 + GORM scope（禁止写业务逻辑）
├── middleware/ # 中间件（顺序：Recovery → CORS → Logger → 路由）
└── router/    # 路由注册
```

## 统一响应格式

```json
{ "code": 0, "message": "成功", "data": {} }
```

- HTTP 状态码只用 200 / 401 / 403 / 404 / 500
- 业务失败统一 HTTP 200，用 `code` 字段区分（0=成功，非0=失败）
- **禁止在代码里硬编码错误码数字**，必须使用 `errcode.XXX` 枚举

```go
response.Success(c, data)                               // 成功
response.Fail(c, errcode.UserNotFound)                  // 业务失败
response.FailMsg(c, errcode.InvalidParams, "自定义消息")  // 带自定义消息
```

## 错误码分段

```
0     成功
1xxxx 通用（10000服务器错误 10001参数错误 10002不存在 10003限流）
2xxxx 认证（20001未授权 20002过期 20003无效 20004无权限）
3xxxx 用户模块
4xxxx 订单模块
```

新增错误码在 `pkg/response/errcode/errcode.go` 追加。

## 分页规范

请求：`?page_num=1&page_size=20`

```go
var p pagination.Page
c.ShouldBindQuery(&p)
p.Normalize()
response.Success(c, pagination.NewResult(&p, total, list))
```

响应 `data` 字段：`page_num` / `page_size` / `total` / `list`

## Go 代码规范

- 包名小写单词，不用下划线
- 导出符号 PascalCase，私有符号 camelCase
- 导出函数必须有注释：`// FuncName 动词描述`
- 错误立即处理，禁止 `_, _ =` 忽略错误
- 日志 key 使用中文，与项目统一风格
- Model 必须嵌入 `model.Model`（含 ID / 时间 / 软删除）
- 密码等敏感字段 json tag 标注 `json:"-"`

## 配置规范

- `config.yaml` 的 `server.mode` 决定加载环境：`debug`→`config.dev.yaml`，`release`→`config.prod.yaml`
- `config.prod.yaml` 已 gitignore，不提交敏感信息
- 禁止代码里硬编码 IP、端口、密钥

## API 规范

- 遵循 REST 风格，资源名用复数名词
- 路由：`/api/v1/auth/*`（公开）、`/api/v1/*`（JWT 保护）
- 404 / 405 均返回 JSON，不返回 HTML
