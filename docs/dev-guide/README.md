# YUNO Nexus 开发指南

> 版本: 0.1.0  
> by Yu_xiaohei

## 项目概述

YUNO Nexus 是一个轻量级、高安全性的内网穿透工具，支持 TCP/UDP/HTTP/HTTPS/WebSocket 协议转发。

**服务端**: YUNO Nexus Server  
**客户端**: YUNO Nexus

## 当前进度

### 已完成

1. **项目结构搭建**
   - 创建完整的目录结构
   - 配置 .gitignore
   - 初始化 Git 仓库

2. **服务端基础代码**
   - 配置加载模块 (`server/internal/config/`)
   - 数据库连接模块 (`server/internal/database/`)
   - JWT认证中间件 (`server/internal/api/middleware/auth/`)
   - API服务入口 (`server/cmd/server/main.go`)
   - 代理服务入口 (`server/cmd/proxy/main.go`)

3. **数据库设计**
   - 完整的数据库迁移文件
   - 包含10个核心表

4. **Docker配置**
   - docker-compose.yml
   - Nginx配置
   - Dockerfile（API服务和代理服务）

5. **文档**
   - 架构设计文档
   - README.md
   - 环境变量示例

### 进行中

- [ ] API处理器实现
- [ ] 用户认证逻辑
- [ ] 设备管理逻辑
- [ ] 隧道管理逻辑

### 待开发

- [ ] Web管理界面
- [ ] Tauri桌面客户端
- [ ] Flutter Android客户端
- [ ] 流量统计
- [ ] 配置导入导出
- [ ] 自动更新
- [ ] 系统通知

## 快速开始

### 1. 环境要求

- Go 1.22+
- PostgreSQL 16+
- Redis 7+
- Node.js 18+ (用于Web管理界面)
- Docker (可选，用于部署)

### 2. 本地开发

```bash
# 克隆项目
git clone <repository-url>
cd YunoNexus

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件

# 启动数据库（使用Docker）
cd docker
docker-compose up -d postgres redis

# 运行数据库迁移
cd ../server
go run cmd/server/main.go
```

### 3. Docker部署

```bash
cd docker
docker-compose up -d
```

## 项目结构

```
YunoNexus/
├── docs/                    # 文档目录
│   ├── architecture.md      # 架构设计文档
│   ├── api/                 # API文档
│   ├── user-guide/          # 用户手册
│   └── dev-guide/           # 开发指南
├── server/                  # Go服务端
│   ├── cmd/                 # 入口文件
│   ├── internal/            # 内部模块
│   ├── pkg/                 # 公共包
│   └── migrations/          # 数据库迁移
├── web-admin/               # React管理界面
├── client-desktop/          # Tauri桌面客户端
├── client-android/          # Flutter Android客户端
├── docker/                  # Docker配置
├── scripts/                 # 脚本工具
├── .gitignore
├── .env.example
└── README.md
```

## 开发规范

### 代码规范

- Go：遵循 `gofmt` + `golangci-lint`
- TypeScript：遵循 `eslint` + `prettier`
- Dart：遵循 `dart format`

### 提交规范

- 提交信息使用中文
- 格式：`<类型>: <描述>`
- 类型：feat/fix/docs/style/refactor/test/chore

### 测试规范

- 测试文件使用 `test_` 前缀
- 完成测试后删除或归档到隔离文件夹
- 覆盖核心业务逻辑

### 文档规范

- 所有文档放在 `docs/` 目录
- 使用 Markdown 格式
- 按内容分子目录

### 安全规范

- `.env` 文件不提交 Git
- `.agents` 文件夹不提交 Git
- 敏感信息加密存储
- API 接口需要认证

## 下一步

1. 实现 API 处理器
2. 完善用户认证逻辑
3. 实现设备管理功能
4. 实现隧道管理功能
5. 开始 Web 管理界面开发

## 问题反馈

如有问题，请提交 Issue 或联系开发团队。
