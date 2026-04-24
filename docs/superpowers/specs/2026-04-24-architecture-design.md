# pveCloud 架构设计文档

**日期**：2026-04-24  
**状态**：已确认

---

## 一、项目概述

pveCloud 是一个云服务器销售平台，提供从产品浏览、下单、支付、自动开通到实例管理的完整电商闭环。

**技术栈**：
- 后端：Go 模块化单体，三入口独立进程
- 数据库：MariaDB（业务事实唯一来源）+ Redis（缓存/会话/幂等）
- 前端：Bun + Vue3 SPA × 2（官网用户端 + 管理后台）

---

## 二、核心业务范围

| 模块 | 说明 |
|------|------|
| 产品目录 | 云服务器规格展示、价格管理 |
| 下单 | 用户选配下单，事务内写订单+账单+预留 |
| 支付 | 微信支付 + 支付宝 + 余额三种方式 |
| 余额 | 用户充值账户，微信/支付宝充值入账 |
| 自动开通 | 支付成功后异步创建实例（PVE Mock） |
| 实例管理 | 用户侧开关机/重建，管理侧全量操作 |
| 工单 | 用户提交工单，管理员回复处理 |
| 公告 | 管理员发布，用户端展示 |
| 通知 | 邮件 + 短信（注册验证、找回密码、开通通知） |
| 审计 | 业务关键事件记录 |

---

## 三、后端架构

### 3.1 进程分离

```
public-api  :8080   用户侧 HTTP API
admin-api   :8081   管理侧 HTTP API
worker      :8082   异步任务处理（PVE开通、通知发送）
```

三个进程共享同一套 `internal/` 代码库，独立编译部署。安全边界：admin-api 泄露不影响用户侧。

### 3.2 模块 path

```
github.com/AeolianCloud/pveCloud/server
```

### 3.3 目录结构

```
server/
  go.mod
  go.sum
  config/
    config.yaml           # 唯一配置文件，YAML only
  cmd/
    public-api/
      main.go
    admin-api/
      main.go
    worker/
      main.go
  internal/
    bootstrap/
      public.go           # public-api 装配
      admin.go            # admin-api 装配
      worker.go           # worker 装配
    common/
      database/
        db.go             # MariaDB 连接 + WithTx helper
        redis.go          # Redis 连接
      response/
        response.go       # 统一 JSON 响应
      errors/
        errors.go         # 业务错误码
      testutil/
        testutil.go       # 集成测试辅助
    auth/
      jwt.go
      middleware.go
    user/                 # 用户注册登录
      handler.go
      service.go
      repository.go
      model.go
    adminuser/            # 管理员登录
      handler.go
      service.go
      repository.go
      model.go
    catalog/              # 商品目录
      handler.go
      service.go
      repository.go
      model.go
    order/                # 订单状态机
      handler.go
      service.go
      repository.go
      model.go
    payment/              # 支付单
      handler.go
      service.go
      repository.go
      model.go
      provider/
        wechat.go         # 微信支付 adapter
        alipay.go         # 支付宝 adapter
    billing/              # 余额账户 + 充值
      handler.go
      service.go
      repository.go
      model.go
    task/                 # 异步任务 claim/execute/retry
      handler.go
      service.go
      repository.go
      model.go
      executor.go
    instance/             # 实例生命周期
      handler.go
      service.go
      repository.go
      model.go
    resource/             # PVE provider 合约
      provider.go         # 接口定义
      mock.go             # Mock 实现（MVP）
    ticket/               # 工单
      handler.go
      service.go
      repository.go
      model.go
    notification/         # 通知
      email.go
      sms.go
    audit/                # 审计日志
      audit.go
```

### 3.4 架构约定

1. **MariaDB 是唯一业务事实源**，Redis 只做缓存/会话/幂等辅助
2. **装配顺序固定**：配置 → DB/Redis → repository → service → handler
3. **三条强事务边界**：
   - 下单事务：`orders` + `billing_records` + `payment_orders` + 预留关系
   - 支付成功事务：回调日志 + `payment_orders` 状态 + `orders` 状态 + 创建唯一任务
   - 开通成功事务：`instances` + `instance_services` + `orders` 状态 + 任务完成
4. **事务 helper**：`common/database.WithTx(ctx, db, fn)`
5. **net/http ServeMux**：不引入第三方路由框架
6. **层职责严格**：handler 只做路由/参数解析，service 只做业务逻辑，repository 只做 DB 操作
7. **配置只用 YAML**：`server/config/config.yaml`，无 `.env`，无环境变量覆盖

---

## 四、前端架构

### 4.1 web/（官网 + 用户控制台）

**定位**：营销官网与用户控制台合一的 SPA。未登录展示营销页，登录后切换为控制台布局。

**布局**：
- 未登录 → `MarketingLayout.vue`（顶部导航 + 页脚）
- 已登录 → `ConsoleLayout.vue`（左侧导航 + 顶部状态栏）

**首页风格**：SaaS 产品风，大标题 + 产品特性卡片 + 价格表 + CTA 按钮。

**路由结构**：

```
营销区（未登录可访问）
  /                      首页：Hero + 产品特性 + 价格表 + CTA
  /products              产品列表
  /products/:id          产品详情 + 购买入口
  /notices               公告列表
  /notices/:id           公告详情
  /help                  帮助文档目录
  /help/:slug            帮助文档详情
  /auth/login            登录
  /auth/register         注册（邮箱 + 手机）
  /auth/forgot           找回密码
  /auth/verify-email     邮箱验证回调

控制台区（需登录）
  /console               总览：实例数、余额、待处理工单
  /console/instances     实例列表
  /console/instances/:id 实例详情：开关机、重建
  /console/orders        订单历史
  /console/orders/:id    订单详情
  /console/billing       余额 + 消费记录
  /console/billing/topup 充值（微信/支付宝）
  /console/tickets       工单列表
  /console/tickets/new   提交工单
  /console/tickets/:id   工单详情 + 回复
  /console/profile       账号设置
```

**目录结构**：

```
web/
  index.html
  package.json
  vite.config.ts
  tsconfig.json
  bun.lock
  src/
    main.ts
    App.vue
    env.d.ts
    router/
      index.ts
    stores/
      auth.ts
      cart.ts
    api/
      auth.ts
      catalog.ts
      instance.ts
      order.ts
      billing.ts
      ticket.ts
      notice.ts
    lib/
      http.ts
      format.ts
    layouts/
      MarketingLayout.vue
      ConsoleLayout.vue
    views/
      marketing/
        HomePage.vue
        ProductListPage.vue
        ProductDetailPage.vue
        NoticePage.vue
        NoticeDetailPage.vue
        HelpPage.vue
        HelpDetailPage.vue
      auth/
        LoginPage.vue
        RegisterPage.vue
        ForgotPage.vue
        VerifyEmailPage.vue
      console/
        DashboardPage.vue
        InstanceListPage.vue
        InstanceDetailPage.vue
        OrderListPage.vue
        OrderDetailPage.vue
        BillingPage.vue
        TopupPage.vue
        TicketListPage.vue
        TicketNewPage.vue
        TicketDetailPage.vue
        ProfilePage.vue
    components/
      common/
        AppButton.vue
        AppTable.vue
        AppModal.vue
        AppPagination.vue
        StatusBadge.vue
      marketing/
        HeroSection.vue
        PricingCard.vue
        FeatureGrid.vue
      console/
        InstanceCard.vue
        OrderRow.vue
        BalanceWidget.vue
    styles/
      web.css
```

---

### 4.2 admin/（管理后台）

**定位**：内部运营后台，简洁 SaaS 风格。

**布局**：左侧固定导航 + 顶部状态栏（管理员信息 + 退出）。

**路由结构**：

```
/login                   管理员登录

后台（需登录）
  /dashboard             概览：订单数、收入、实例数、用户数
  /products              商品列表
  /products/new          新建商品
  /products/:id/edit     编辑商品
  /orders                订单列表
  /orders/:id            订单详情 + 手动操作
  /instances             实例列表
  /instances/:id         实例详情 + 操作
  /users                 用户列表
  /users/:id             用户详情 + 余额调整
  /billing               充值记录 + 流水
  /tickets               工单列表（未处理优先）
  /tickets/:id           工单详情 + 回复 + 关闭
  /tasks                 异步任务监控
  /notices               公告列表
  /notices/new           新建公告
  /notices/:id/edit      编辑公告
  /admins                管理员账号列表
  /admins/new            新建管理员
```

**目录结构**：

```
admin/
  index.html
  package.json
  vite.config.ts
  tsconfig.json
  bun.lock
  src/
    main.ts
    App.vue
    env.d.ts
    router/
      index.ts
    stores/
      auth.ts
    lib/
      http.ts
      format.ts
    styles/
      admin.css
    api/
      auth.ts
      dashboard.ts
      catalog.ts
      order.ts
      instance.ts
      user.ts
      billing.ts
      ticket.ts
      task.ts
      notice.ts
      admin.ts
    layouts/
      AdminLayout.vue
    components/
      shared/
        AppTable.vue
        AppModal.vue
        AppPagination.vue
        AppButton.vue
        StatusBadge.vue
        StatCard.vue
      shell/
        SideNav.vue
        TopBar.vue
      dashboard/
        RevenueChart.vue
        OrderTrendChart.vue
        StatSummary.vue
      catalog/
        ProductForm.vue
        ProductStatusToggle.vue
      order/
        OrderStatusFlow.vue
        OrderActionBar.vue
      instance/
        InstanceActionBar.vue
        InstanceStatusBadge.vue
      user/
        UserBalanceAdjust.vue
        UserStatusToggle.vue
      ticket/
        TicketReplyForm.vue
        TicketStatusBar.vue
      task/
        TaskRetryButton.vue
        TaskLogViewer.vue
      notice/
        NoticeForm.vue
    views/
      auth/
        LoginPage.vue
      dashboard/
        DashboardPage.vue
      catalog/
        ProductListPage.vue
        ProductFormPage.vue
      order/
        OrderListPage.vue
        OrderDetailPage.vue
      instance/
        InstanceListPage.vue
        InstanceDetailPage.vue
      user/
        UserListPage.vue
        UserDetailPage.vue
      billing/
        BillingPage.vue
      ticket/
        TicketListPage.vue
        TicketDetailPage.vue
      task/
        TaskListPage.vue
      notice/
        NoticeListPage.vue
        NoticeFormPage.vue
      admins/
        AdminListPage.vue
```

---

## 五、基础设施

```
docker-compose.yml    MariaDB + Redis 本地开发环境
start.bat             Windows 本地一键启动
```

**端口规划**：

| 服务 | 端口 |
|------|------|
| public-api | 8080 |
| admin-api | 8081 |
| worker | 8082 |
| web dev server | 5173 |
| admin dev server | 5174 |
| MariaDB | 3306 |
| Redis | 6379 |

---

## 六、认证方案

- **用户侧**：账号密码 + JWT，注册/找回密码支持邮箱验证 + 手机短信双通道
- **管理侧**：账号密码 + JWT，独立 token 不与用户侧共用

---

## 七、PVE 集成策略

MVP 阶段使用 Mock provider，`resource/provider.go` 定义接口合约：

```go
type Provider interface {
    CreateVM(ctx context.Context, spec VMSpec) (VMInfo, error)
    StartVM(ctx context.Context, vmID string) error
    StopVM(ctx context.Context, vmID string) error
    RebuildVM(ctx context.Context, vmID string, spec VMSpec) error
    DeleteVM(ctx context.Context, vmID string) error
}
```

后期替换为真实 PVE API 实现，上层 service/task 代码不变。

---

## 八、子项目实施顺序

| 阶段 | 内容 |
|------|------|
| Subproject 1 | 后端核心闭环：目录→下单→支付→开通→实例管理 |
| Subproject 2 | Admin 前端：后台管理全页面 |
| Subproject 3 | Web 前端：官网 + 用户控制台 |
| Subproject 4 | 工单系统（前后端） |
| Subproject 5 | 真实支付接入（微信/支付宝） |
| Subproject 6 | 真实 PVE API 替换 Mock |
