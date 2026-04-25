# Web 前端文档

本文件对应未来代码目录 `web/`，范围是官网和用户中心。

## 定位

`web/` 面向普通用户，包含营销官网、产品浏览、下单、支付结果、用户中心、订单、实例和工单。

## API 边界

```text
web -> /api/*
```

`web` 不直接访问 `/admin-api/*`。

## 主要页面

- 首页：品牌介绍、优势、套餐入口。
- 产品页：云服务器套餐、地域、配置说明。
- 价格页：套餐价格、计费方式。
- 登录/注册页：用户认证。
- 下单页：选择地域、镜像、套餐、周期。
- 支付结果页：支付成功/失败展示。
- 用户中心：资料、安全、余额。
- 我的订单：订单列表、订单详情、继续支付、取消。
- 我的云服务器：实例列表、开机、关机、重启、重装。
- 工单：提交问题、查看回复。

## 目录意图

```text
web/
├─ src/
│  ├─ app/
│  ├─ router/
│  ├─ layouts/
│  ├─ pages/
│  ├─ components/
│  ├─ api/
│  ├─ stores/
│  ├─ styles/
│  └─ utils/
├─ index.html
├─ package.json
├─ tsconfig.json
└─ vite.config.ts
```

## 状态管理

- `auth`：用户 token、用户资料。
- `cart`：下单配置暂存，可选。
- `order`：当前订单状态。
- `instance`：实例列表缓存，可选。

## 独立管理规则

`web` 与 `admin` 完全独立管理，不使用公共 `shared/` 前端包。

`web` 自己维护：

- `web/src/api/` 请求封装和业务接口。
- `web/src/stores/` 用户端状态。
- `web/src/utils/` 金额、时间、状态格式化工具。
- `web/src/types/` 用户端需要的 TypeScript 类型。
- `web/src/constants/` 用户端状态枚举和常量。

不要从 `admin/` 导入页面、组件、请求文件、状态或工具。
