# Web 基础前台实施计划

本计划描述 Web 基础前台阶段的目标、边界、路线和验收重点。

## 阶段目标

创建最小可运行用户端前台，为后续产品、账号、订单、支付、实例和工单业务接入提供前端壳、路由、样式基础和请求边界准备。

## 阶段边界

### 当前阶段纳入

- 创建 `web/` 前端应用
- 用户端基础布局和导航
- Home
- Products 占位页
- Pricing 占位页
- Login / Register 占位页
- User Center 占位页
- 404
- `/api/*` 请求封装骨架
- 本地开发脚本识别 `web/package.json` 后启动 `web`
- 用户端构建验证

### 当前阶段不纳入

- 用户端 `/api/*` 业务接口
- 用户注册、登录、会话恢复、刷新和退出
- 产品套餐真实数据
- 订单创建、订单列表和订单详情
- 支付发起、回调和支付结果确认
- 实例列表、实例开通和实例操作
- 工单创建、回复和列表
- 用户中心真实账号资料
- 数据库业务域迁移
- 管理端新增业务页面

## 实施路线

### Phase W0：文档与契约确认

确认 `docs/web/architecture.md`、本计划、本地开发说明、部署说明和进度总览已经说明本阶段范围。

### Phase W1：Web 应用骨架

创建 `web/`，沿用 Bun、Vite、Vue 3、TypeScript、Vue Router、Pinia 和 Axios。

### Phase W2：页面与路由

落地公开路由、基础布局、静态页面、占位页和 404。

### Phase W3：请求边界与验证

预留 `/api/*` 请求封装骨架，但不接真实业务接口。验证 `web` 构建通过，并确认 `admin` 与 `web` 无运行时代码共享。

## 验收口径

- `web/` 能独立安装依赖和构建
- `web/` 不导入 `admin/` 运行时代码
- `web/` 不调用 `/admin-api/*`
- 本阶段不新增 `/api/*` 业务接口和数据库迁移
- 目标路由可访问，未匹配路由进入 404
- 页面在桌面和移动端可用
- 本地开发说明和部署说明与 `web` 存在后的边界一致

## 后续阶段入口

后续若要开放产品、账号、订单、支付、实例或工单能力，必须先更新：

- `docs/web/architecture.md`
- `docs/server/api/`
- `docs/server/database/design.md`
- `server/migrations/`
- 必要时更新 `docs/admin/`、`docs/development/` 和 `docs/operations/`
