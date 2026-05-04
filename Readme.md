# 个人网站后端

> 该项目仍处于早期开发阶段，当前功能、接口与目录结构仅代表现阶段实现，不代表最终版本。

一个使用 Go、Gin 和 PostgreSQL 构建的个人网站后端服务。

## 项目状态

- 早期开发中，接口和数据结构仍可能调整。
- 暂未提供稳定版本承诺。

## 功能特性

- 基于 Gin 的 HTTP API 服务
- PostgreSQL 连接池管理
- 统一 JSON 响应格式
- JWT 认证能力
- 管理员权限中间件

## 技术栈

| 类型 | 技术 |
| --- | --- |
| 语言 | Go 1.26.2 |
| Web 框架 | Gin |
| 数据库 | PostgreSQL |
| 数据库驱动 | pgx / pgxpool |
| 认证 | JWT |
| 配置加载 | godotenv |

## 项目结构

```text
.
├── main.go                 # 服务入口与路由注册
├── go.mod                  # Go 模块定义
├── go.sum                  # 依赖版本锁定
├── internal
│   ├── auth
│   │   └── jwt.go          # JWT 生成、解析与校验
│   ├── database
│   │   └── db.go           # PostgreSQL 连接池初始化与关闭
│   ├── middleware
│   │   └── auth.go         # JWT 认证与管理员权限校验
│   ├── model
│   │   ├── hitokoto.go     # 一言数据模型
│   │   ├── response.go     # 通用响应结构
│   │   └── user.go         # 用户数据模型
│   └── service
│       └── hitokoto
│           └── hitokoto.go # 一言相关接口处理逻辑
└── Readme.md               # 项目说明文档
```

## 快速开始

安装依赖：

```bash
go mod tidy
```

启动服务：

```bash
go run .
```

服务默认监听 Gin 的 `:8080` 端口。

## 接口说明

### 公开接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/hitokoto/` | 随机获取一条一言 |

### 管理接口

管理接口需要在请求头中携带管理员 JWT：

```http
Authorization: Bearer <token>
```

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/hitokoto/list` | 获取一言列表 |
| GET | `/api/hitokoto/:id` | 根据 ID 获取一言 |
| POST | `/api/hitokoto/` | 新增一言 |
| DELETE | `/api/hitokoto/:id` | 根据 ID 删除一言 |

## 响应格式

接口统一返回 JSON：

```json
{
  "code": 200,
  "message": "请求成功",
  "data": {}
}
```


## 功能特性

- [x] 基于 Gin 的 HTTP API 服务
- [x] PostgreSQL 连接池支持
- [x] 统一 JSON 响应格式
- [x] JWT 认证能力
- [x] 管理员权限中间件
- [x] 一言内容随机获取
- [x] 一言内容列表查询
- [x] 一言内容新增与删除
- [ ] 用户注册、登录与权限管理完善
- [ ] 接口测试与单元测试
- [ ] 接口文档、部署说明
- [ ] 更多业务模块扩展

## 致谢

本项目受益于开源社区长期积累的工具、框架与文档。感谢这些项目让开发变得更可靠、更高效。

### 核心依赖

- [Go](https://go.dev/)：简洁、高效的后端开发语言。
- [Gin](https://gin-gonic.com/)：轻量快速的 Go Web 框架。
- [PostgreSQL](https://www.postgresql.org/)：稳定可靠的关系型数据库。
- [pgx](https://github.com/jackc/pgx)：功能完整的 PostgreSQL 驱动与连接池。
- [jwt](https://github.com/golang-jwt/jwt)：JWT 生成与校验支持。
- [godotenv](https://github.com/joho/godotenv)：本地环境变量加载支持。

### 开发工具
- [Visual Studio Code](https://code.visualstudio.com/)：优秀的代码编辑器与开发环境。
<p align="center">
  <a href="https://code.visualstudio.com/">
    <img src="https://upload.wikimedia.org/wikipedia/commons/9/9a/Visual_Studio_Code_1.35_icon.svg" alt="Visual Studio Code" width="80">
  </a>
</p>
