# 开发文档（DEVELOPMENT）

## 1. 项目概述

本项目为基础的用户管理系统，基于 Go 原生 `net/http` 开发，采用模板渲染方式（非前后端分离），并使用 MySQL 持久化用户数据。系统支持注册、登录、用户列表展示，并对管理员与普通用户做权限区分：仅管理员可编辑/删除用户，普通用户仅可查看。

技术栈：

- Go：`net/http`、`html/template`、`database/sql`
- MySQL：`github.com/go-sql-driver/mysql`
- Session：内存 SessionStore + Cookie（HttpOnly）
- 密码：bcrypt（`golang.org/x/crypto/bcrypt`）
- 页面：复用 `static/` 下现有模板文件（页面资源来自模板内 CDN）

## 2. 目录结构

```
cmd/
  main.go                  # 程序入口
internal/
  app/                     # 启动与依赖注入
  config/                  # 配置与 DB 连接
  controller/              # 控制器（HTTP Handler）
  middleware/              # Session 与鉴权中间件
  model/                   # 领域模型
  repository/              # 仓储接口与实现
    mysql/                 # MySQL 仓储实现
  router/                  # 路由注册
  service/                 # 业务逻辑
  view/                    # 模板渲染
scripts/
  schema.sql               # 数据库初始化脚本
static/
  *.html                   # 模板文件（Go template）
docs/
  DEVELOPMENT.md           # 本文档
```

## 3. 环境要求

- Go：项目 `go.mod` 为 `go 1.23.9`
- MySQL：推荐 5.7+/8.x

## 4. 配置说明

项目默认通过环境变量读取配置：

- `APP_ADDR`：HTTP 监听地址，默认 `:8090`
- `SESSION_COOKIE_NAME`：Session Cookie 名称，默认 `sid`
- `MYSQL_DSN`：MySQL 连接串（driver DSN），默认：

  ```
  root:123456@tcp(127.0.0.1:3306)/user_management?parseTime=true&loc=Local
  ```

相关代码位置：

- 启动与注入：[bootstrap.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/app/bootstrap.go)
- 读取监听地址与 Cookie 名：[config.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/config/config.go)
- 打开 DB 连接与默认 DSN：[db.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/config/db.go)

## 5. 数据库初始化

1) 确保本地 MySQL 可连接。

2) 执行初始化脚本（创建数据库与 users 表）：

- 脚本路径：[schema.sql](file:///d:/23026/Documents/Go%20code/UserManagement/scripts/schema.sql)
- 常用方式（示例）：

```bash
mysql -u root -p < scripts/schema.sql
```

表结构要点：

- 表：`users`
- 密码字段：`password_hash`（存储 bcrypt hash）
- 角色字段：`role`（`admin` / `user`）
- 状态字段：`status`（1 启用，0 禁用）
- 登录时间：`last_time`（登录成功后更新）

## 6. 启动与运行

在项目根目录执行：

```bash
go run ./cmd
```

默认访问地址：

- http://localhost:8090/

根路径 `/` 会根据是否已登录自动跳转：

- 已登录 -> `/users`
- 未登录 -> `/login`

## 7. 路由与权限

路由统一注册在：[router.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/router/router.go)

| 路径 | 方法 | 说明 | 权限 |
|---|---|---|---|
| `/login` | GET/POST | 登录页 / 登录提交 | 公开 |
| `/register` | GET/POST | 注册页 / 注册提交 | 公开 |
| `/logout` | POST | 退出登录 | 公开（未登录也可调用） |
| `/users` | GET | 用户列表（查看所有用户） | 已登录 |
| `/users/edit?id=...` | GET | 编辑页 | 管理员 |
| `/users/edit` | POST | 保存编辑 | 管理员 |
| `/users/delete` | POST | 删除用户（表单提交） | 管理员 |

权限控制实现：

- Session 解析与注入：`WithAuth`（将用户身份写入 request context）
- 必须登录：`RequireLogin`
- 必须管理员：`RequireAdmin`

代码位置：[auth.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/middleware/auth.go)

## 8. 业务规则

业务逻辑集中在 `service` 层：[user_service.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/service/user_service.go)

- 注册
  - 用户名/密码不能为空
  - 用户名唯一（已存在返回 `ErrUsernameTaken`）
  - 第一个注册用户自动成为管理员（`repo.Count(ctx) == 0`）
  - 密码使用 bcrypt 生成 hash，写入 `password_hash`
- 登录
  - 用户不存在 / 密码不匹配 / 状态为禁用（`status == 0`）均视为登录失败
  - 登录成功后更新 `last_time`

## 9. 分层说明（MVC + Repository）

### Controller（HTTP 层）

- 登录/注册/退出：[auth_controller.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/controller/auth_controller.go)
- 用户列表/编辑/删除：[user_controller.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/controller/user_controller.go)

### Service（业务层）

- 用户注册、鉴权、列表、更新、删除：[user_service.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/service/user_service.go)

### Repository（数据访问层）

- 接口定义（供 service 依赖）：[user_repository.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/repository/user_repository.go)
- MySQL 实现：[user_repository_mysql.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/repository/mysql/user_repository_mysql.go)

## 10. Session 设计

Session 为内存存储（进程内 map），重启服务会丢失所有登录态。

- SessionStore：[session.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/middleware/session.go)
- Cookie：
  - `HttpOnly: true`
  - `SameSite: Lax`
  - `Path: /`

## 11. 模板渲染与页面

模板目录：`static/`，渲染器每次渲染会按文件名加载对应模板文件。

- 渲染器实现：[renderer.go](file:///d:/23026/Documents/Go%20code/UserManagement/internal/view/renderer.go)
- 使用的模板函数：
  - `eq a b`：判断相等（用于 role/status 的展示或选中状态）
  - `formatTime t`：将时间格式化为 `YYYY-MM-DD HH:mm`（空时间返回空字符串）

主要模板：

- 登录页：`static/login.html`
- 注册页：`static/register.html`
- 用户列表：`static/userList.html`
- 用户编辑：`static/userEdit.html`

## 12. 常见问题排查

### 1) 启动时报数据库不存在

报错类似：`Unknown database 'user_management'`

- 先执行初始化脚本：[schema.sql](file:///d:/23026/Documents/Go%20code/UserManagement/scripts/schema.sql)
- 或通过 `MYSQL_DSN` 指向已存在的数据库

### 2) 普通用户访问编辑/删除

- 访问 `/users/edit` 或提交 `/users/delete` 会返回 `403 forbidden`（权限校验：`RequireAdmin`）

### 3) 登录后刷新丢失登录态

- 本项目 Session 在内存中；如果服务重启，Session 会全部失效，需要重新登录

## 13. 本地自检建议

```bash
go test ./...
go vet ./...
```
