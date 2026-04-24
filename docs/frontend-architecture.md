# 前端架构设计

## 项目定位

本项目包含两个独立前端：

- `web/`：官网 + 用户中心，面向购买云服务器的客户。
- `admin/`：管理后台，面向平台运营、客服、管理员。

两个前端都采用 Bun + Vue 3 + Vite + TypeScript。两个应用独立开发、独立构建、独立部署，但共享部分类型、请求封装和工具函数。

## 总体目录

```text
pveCloud/
├─ server/                  # Go 后端
├─ web/                     # 官网 + 用户中心
├─ admin/                   # 管理后台
├─ shared/                  # 前端共享代码
├─ docs/
├─ deploy/
├─ package.json             # Bun workspace 根配置
├─ bun.lockb
└─ README.md
```

## 前端技术栈

- 运行/包管理：Bun。
- 框架：Vue 3。
- 构建：Vite。
- 语言：TypeScript。
- 路由：Vue Router。
- 状态：Pinia。
- 请求：Axios 或 Fetch 封装，建议统一封装。
- 样式：Tailwind CSS 或 UnoCSS。
- 后台 UI：Element Plus 或 Naive UI。
- 官网 UI：自定义组件 + Tailwind/UnoCSS，更适合营销展示。

## API 边界

```text
web    -> /api/*
admin  -> /admin-api/*
```

两个前端不直接混用 API。

- 官网/用户中心只访问 `/api/*`。
- 管理后台只访问 `/admin-api/*`。
- 共享请求基础能力放在 `shared/api/`。
- 具体业务 API 按应用分别放在 `web/src/api/` 和 `admin/src/api/`。

## 官网 web 目录

```text
web/
├─ public/
│  ├─ favicon.ico
│  └─ images/
│
├─ src/
│  ├─ app/
│  │  ├─ App.vue
│  │  ├─ bootstrap.ts              # 应用初始化
│  │  └─ providers.ts              # Pinia、Router 等插件注册
│  │
│  ├─ router/
│  │  ├─ index.ts
│  │  ├─ guards.ts                 # 用户登录路由守卫
│  │  └─ routes.ts
│  │
│  ├─ layouts/
│  │  ├─ MarketingLayout.vue       # 官网首页/产品页布局
│  │  ├─ AccountLayout.vue         # 用户中心布局
│  │  └─ BlankLayout.vue           # 登录注册页布局
│  │
│  ├─ pages/
│  │  ├─ home/
│  │  │  └─ HomePage.vue
│  │  ├─ products/
│  │  │  ├─ ProductListPage.vue
│  │  │  └─ ProductDetailPage.vue
│  │  ├─ pricing/
│  │  │  └─ PricingPage.vue
│  │  ├─ auth/
│  │  │  ├─ LoginPage.vue
│  │  │  └─ RegisterPage.vue
│  │  ├─ account/
│  │  │  ├─ ProfilePage.vue
│  │  │  ├─ WalletPage.vue
│  │  │  └─ SecurityPage.vue
│  │  ├─ orders/
│  │  │  ├─ OrderListPage.vue
│  │  │  ├─ OrderCreatePage.vue
│  │  │  ├─ OrderDetailPage.vue
│  │  │  └─ PaymentResultPage.vue
│  │  ├─ instances/
│  │  │  ├─ InstanceListPage.vue
│  │  │  └─ InstanceDetailPage.vue
│  │  ├─ tickets/
│  │  │  ├─ TicketListPage.vue
│  │  │  ├─ TicketCreatePage.vue
│  │  │  └─ TicketDetailPage.vue
│  │  └─ error/
│  │     ├─ NotFoundPage.vue
│  │     └─ ForbiddenPage.vue
│  │
│  ├─ components/
│  │  ├─ common/                   # 通用组件
│  │  ├─ marketing/                # 官网展示组件
│  │  ├─ product/                  # 产品卡片、价格卡片
│  │  ├─ order/                    # 下单相关组件
│  │  └─ instance/                 # 实例状态、操作按钮
│  │
│  ├─ api/
│  │  ├─ auth.ts
│  │  ├─ product.ts
│  │  ├─ order.ts
│  │  ├─ payment.ts
│  │  ├─ instance.ts
│  │  ├─ ticket.ts
│  │  └─ profile.ts
│  │
│  ├─ stores/
│  │  ├─ auth.ts
│  │  ├─ cart.ts                   # 如需要购物车/配置暂存
│  │  ├─ order.ts
│  │  └─ instance.ts
│  │
│  ├─ styles/
│  │  ├─ index.css
│  │  ├─ theme.css
│  │  └─ variables.css
│  │
│  ├─ utils/
│  │  ├─ format.ts
│  │  └─ route.ts
│  │
│  └─ main.ts
│
├─ index.html
├─ package.json
├─ tsconfig.json
└─ vite.config.ts
```

## 官网页面规划

官网不只是展示页，还包含用户购买和控制台入口。

核心页面：

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

## 管理后台 admin 目录

```text
admin/
├─ public/
│  ├─ favicon.ico
│  └─ images/
│
├─ src/
│  ├─ app/
│  │  ├─ App.vue
│  │  ├─ bootstrap.ts
│  │  └─ providers.ts
│  │
│  ├─ router/
│  │  ├─ index.ts
│  │  ├─ guards.ts                 # 管理员登录/权限守卫
│  │  └─ routes.ts
│  │
│  ├─ layouts/
│  │  ├─ AdminLayout.vue           # 侧边栏 + 顶栏
│  │  ├─ AuthLayout.vue            # 管理员登录布局
│  │  └─ BlankLayout.vue
│  │
│  ├─ pages/
│  │  ├─ auth/
│  │  │  └─ LoginPage.vue
│  │  ├─ dashboard/
│  │  │  └─ DashboardPage.vue
│  │  ├─ users/
│  │  │  ├─ UserListPage.vue
│  │  │  └─ UserDetailPage.vue
│  │  ├─ products/
│  │  │  ├─ ProductListPage.vue
│  │  │  ├─ ProductEditPage.vue
│  │  │  ├─ PlanListPage.vue
│  │  │  ├─ RegionListPage.vue
│  │  │  └─ ImageListPage.vue
│  │  ├─ orders/
│  │  │  ├─ OrderListPage.vue
│  │  │  └─ OrderDetailPage.vue
│  │  ├─ payments/
│  │  │  ├─ PaymentListPage.vue
│  │  │  └─ WalletTransactionPage.vue
│  │  ├─ instances/
│  │  │  ├─ InstanceListPage.vue
│  │  │  └─ InstanceDetailPage.vue
│  │  ├─ tickets/
│  │  │  ├─ TicketListPage.vue
│  │  │  └─ TicketDetailPage.vue
│  │  ├─ admins/
│  │  │  ├─ AdminListPage.vue
│  │  │  ├─ RoleListPage.vue
│  │  │  └─ PermissionListPage.vue
│  │  ├─ system/
│  │  │  ├─ SystemConfigPage.vue
│  │  │  ├─ AuditLogPage.vue
│  │  │  └─ NodeConfigPage.vue
│  │  └─ error/
│  │     ├─ NotFoundPage.vue
│  │     └─ ForbiddenPage.vue
│  │
│  ├─ components/
│  │  ├─ common/                   # 通用后台组件
│  │  ├─ table/                    # 表格封装
│  │  ├─ form/                     # 表单封装
│  │  ├─ search/                   # 搜索筛选
│  │  ├─ dialog/                   # 弹窗
│  │  └─ charts/                   # 统计图表
│  │
│  ├─ api/
│  │  ├─ auth.ts
│  │  ├─ dashboard.ts
│  │  ├─ user.ts
│  │  ├─ product.ts
│  │  ├─ order.ts
│  │  ├─ payment.ts
│  │  ├─ instance.ts
│  │  ├─ ticket.ts
│  │  ├─ admin.ts
│  │  └─ system.ts
│  │
│  ├─ stores/
│  │  ├─ auth.ts
│  │  ├─ permission.ts
│  │  ├─ layout.ts
│  │  └─ tabs.ts                   # 如需要多标签页
│  │
│  ├─ styles/
│  │  ├─ index.css
│  │  ├─ theme.css
│  │  └─ variables.css
│  │
│  ├─ utils/
│  │  ├─ format.ts
│  │  ├─ permission.ts
│  │  └─ route.ts
│  │
│  └─ main.ts
│
├─ index.html
├─ package.json
├─ tsconfig.json
└─ vite.config.ts
```

## 管理后台页面规划

核心页面：

- 登录页：管理员登录。
- 仪表盘：用户数、订单数、收入、实例数、待处理工单。
- 用户管理：用户列表、详情、余额、订单、实例。
- 产品管理：套餐、地域、节点、镜像、价格。
- 订单管理：订单列表、详情、人工取消、备注。
- 支付管理：支付记录、余额流水、人工入账。
- 实例管理：实例列表、详情、开关机、重启、同步状态。
- 工单管理：工单列表、回复、关闭。
- 管理员管理：管理员、角色、权限。
- 系统设置：站点配置、PVE 节点配置、支付配置、审计日志。

## shared 共享目录

```text
shared/
├─ api/
│  ├─ http.ts                    # 基础请求实例
│  ├─ error.ts                   # 请求错误处理
│  └─ token.ts                   # token 读写工具
│
├─ types/
│  ├─ common.ts                  # 分页、统一响应
│  ├─ user.ts
│  ├─ product.ts
│  ├─ order.ts
│  ├─ payment.ts
│  ├─ instance.ts
│  └─ ticket.ts
│
├─ constants/
│  ├─ status.ts                  # 订单/支付/实例状态
│  └─ route.ts
│
├─ utils/
│  ├─ format.ts                  # 时间、金额、状态格式化
│  ├─ storage.ts
│  └─ validators.ts
│
└─ package.json
```

## shared 使用规则

可以共享：

- API 请求基础封装。
- 通用 TS 类型。
- 状态枚举。
- 金额、时间、状态格式化工具。
- 表单校验工具。

不要共享：

- 具体页面组件。
- 官网营销组件。
- 后台业务表格组件。
- 前端路由配置。
- 应用级状态。

避免 `shared/` 变成杂物目录。只有两个前端都确实需要的东西才放进去。

## 状态管理规则

`web`：

- `auth`：用户 token、用户资料。
- `cart`：下单配置暂存，可选。
- `order`：当前订单状态。
- `instance`：实例列表缓存，可选。

`admin`：

- `auth`：管理员 token、管理员信息。
- `permission`：权限码、菜单权限。
- `layout`：侧边栏、主题、折叠状态。
- `tabs`：后台多标签页，可选。

## 请求封装规则

`shared/api/http.ts` 只负责基础能力：

- baseURL。
- 请求超时。
- token 注入。
- 统一错误处理。
- 统一响应解包。

具体业务请求放应用内：

```text
web/src/api/order.ts
admin/src/api/order.ts
```

这样能保证两个前端 API 边界清楚。

## 构建和开发命令建议

根目录 `package.json`：

```json
{
  "scripts": {
    "dev:web": "bun --cwd web dev",
    "dev:admin": "bun --cwd admin dev",
    "build:web": "bun --cwd web build",
    "build:admin": "bun --cwd admin build",
    "lint:web": "bun --cwd web lint",
    "lint:admin": "bun --cwd admin lint"
  }
}
```

## 部署建议

```text
www.example.com        -> web/dist
admin.example.com      -> admin/dist
api.example.com/api    -> Go /api/*
api.example.com/admin-api -> Go /admin-api/*
```

也可以同域部署：

```text
/              -> web/dist
/admin/        -> admin/dist
/api/          -> Go 用户端 API
/admin-api/    -> Go 管理端 API
```

## 当前不建议

- 不建议把 web 和 admin 做成一个 Vue 应用。
- 不建议两个前端共用同一套路由和页面组件。
- 不建议一开始引入复杂微前端。
- 不建议 `shared/` 放太多业务 UI。
- 不建议 admin 直接复用 web 的请求文件。
