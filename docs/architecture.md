# YUNO Nexus 技术架构设计文档

> 版本: 0.1.0  
> 日期: 2026-07-03  
> 状态: 设计阶段  
> by Yu_xiaohei

---

## 1. 系统架构总览

### 1.1 架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          互联网 / 公网                                    │
└───────────────────────────┬─────────────────────────────────────────────┘
                            │
              ┌─────────────▼─────────────┐
              │     Nginx 反向代理         │
              │  (TLS 终止 / WebSocket)    │
              │  :443 (HTTPS+WSS)         │
              │  :80  (HTTP 重定向)        │
              └─────────┬───────┬─────────┘
                        │       │
          ┌─────────────▼──┐  ┌─▼──────────────┐
          │  Web 管理界面   │  │  Go API 服务    │
          │  (React SPA)   │  │  (Echo)         │
          │  静态文件托管    │  │  :8080          │
          └────────────────┘  └──────┬──────────┘
                                     │
                    ┌────────────────▼────────────────┐
                    │         PostgreSQL               │
                    │  (用户/隧道/日志/配置)            │
                    │  :5432                           │
                    └─────────────────────────────────┘
                                     │
              ┌──────────────────────▼──────────────────────┐
              │           Go 核心代理引擎                     │
              │  ┌──────────┐ ┌──────────┐ ┌─────────────┐  │
              │  │ TCP 代理  │ │ UDP 代理  │ │ HTTP 代理    │  │
              │  └──────────┘ └──────────┘ └─────────────┘  │
              │  ┌──────────┐ ┌──────────┐                  │
              │  │ WS 代理   │ │ 数据压缩  │                  │
              │  └──────────┘ └──────────┘                  │
              └──────────────┬─────────────────────────────┘
                             │ 加密隧道 (AES-256-GCM + zlib)
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────┐  ┌─────▼──────┐  ┌───▼──────────┐
     │  Windows    │  │   Linux    │  │   Android    │
     │  客户端     │  │   客户端    │  │   客户端     │
     │  (Tauri+Go) │  │ (Tauri+Go) │  │(Flutter+Go)  │
     └────────────┘  └────────────┘  └──────────────┘
```

### 1.2 通信流概览

| 流向 | 协议 | 加密 | 说明 |
|------|------|------|------|
| 管理员 → Web UI → API | HTTPS/WSS | TLS 1.3 | 管理界面与后端通信 |
| 客户端 → 服务端 | WSS | TLS 1.3 + AES-256-GCM | 控制通道（心跳/配置/状态） |
| 服务端 → 客户端 | WSS | TLS 1.3 + AES-256-GCM | 隧道数据转发 |
| 内部服务 → 客户端 | TCP/UDP | 明文（内网） | 客户端本地转发 |
| 服务端 → 公网用户 | TCP/UDP/HTTP | TLS 或明文 | 外部访问入口 |

### 1.3 核心技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 服务端 | Go + Echo | Go 1.22+ |
| 数据库 | PostgreSQL | 16+ |
| Web UI | React + TypeScript + Vite | React 18+ |
| UI 框架 | Ant Design 5 | - |
| Windows/Linux 客户端 | Tauri 2.x + Go (CGO) | - |
| Android 客户端 | Flutter 3.x + Go (gomobile) | - |
| 加密 | AES-256-GCM + zlib | - |
| 传输层 | WebSocket (gorilla/websocket) | - |
| 容器化 | Docker + Docker Compose | - |
| 反向代理 | Nginx | 1.25+ |

---

## 2. 端口规划

| 端口用途 | 端口号 | 对外暴露 | 说明 |
|----------|--------|----------|------|
| HTTPS伪装 | **443** | 是 | 标准HTTPS，伪装成正常网站 |
| 客户端WebSocket | **9443** | 是 | 避开常见8443/8080 |
| 隧道端口范围 | **30000-30100** | 是 | 避开FRP/NPS常用端口 |
| API服务 | 8080 | 否 | 内部通信，不对外 |
| PostgreSQL | 5432 | 否 | 内部数据库 |
| Redis | 6379 | 否 | 内部缓存 |

**对外暴露端口总结**：443、9443、30000-30100

---

## 3. 数据库表结构

### 3.1 ER 关系图

```
users 1──N devices
users 1──N tunnels
tunnels 1──N port_bindings
tunnels 1──N traffic_logs
devices 1──N sessions
users 1──N audit_logs
```

### 3.2 完整表结构

详见 `server/internal/database/migrations/001_init.up.sql`

主要表：
- **users** - 用户表
- **devices** - 设备表
- **sessions** - 会话表
- **tunnels** - 隧道表
- **port_bindings** - 端口绑定表
- **traffic_logs** - 流量日志表（按月分区）
- **port_allocations** - 端口分配表
- **audit_logs** - 审计日志表
- **system_configs** - 系统配置表
- **encryption_keys** - 加密密钥表

---

## 4. API 接口设计

### 4.1 API 规范

- 基础路径: `/api/v1`
- 认证方式: Bearer Token (JWT)
- 内容类型: `application/json`

### 4.2 主要接口

#### 认证模块
- `POST /api/v1/auth/register` - 注册
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/refresh` - 刷新Token
- `POST /api/v1/auth/logout` - 登出

#### 用户模块
- `GET /api/v1/users/me` - 获取当前用户信息
- `PUT /api/v1/users/me` - 更新用户信息

#### 设备管理
- `GET /api/v1/devices` - 列出设备
- `GET /api/v1/devices/:id` - 获取设备详情
- `PUT /api/v1/devices/:id` - 更新设备
- `DELETE /api/v1/devices/:id` - 吊销设备

#### 隧道管理
- `GET /api/v1/tunnels` - 列出隧道
- `POST /api/v1/tunnels` - 创建隧道
- `GET /api/v1/tunnels/:id` - 获取隧道详情
- `PUT /api/v1/tunnels/:id` - 更新隧道
- `DELETE /api/v1/tunnels/:id` - 删除隧道
- `POST /api/v1/tunnels/:id/start` - 启动隧道
- `POST /api/v1/tunnels/:id/stop` - 停止隧道
- `GET /api/v1/tunnels/:id/stats` - 获取隧道统计

#### 流量统计
- `GET /api/v1/traffic/overview` - 流量概览
- `GET /api/v1/traffic/tunnel/:id` - 按隧道查询
- `GET /api/v1/traffic/user/:id` - 按用户查询

#### 配置导入导出
- `GET /api/v1/config/export` - 导出配置
- `POST /api/v1/config/import` - 导入配置

#### 管理员接口
- `GET /api/v1/admin/users` - 管理员查看用户
- `PUT /api/v1/admin/users/:id` - 管理员更新用户
- `GET /api/v1/admin/system/stats` - 系统统计
- `GET /api/v1/admin/audit-logs` - 审计日志

---

## 5. 客户端-服务端通信协议

### 5.1 传输层

所有客户端-服务端通信使用 **WebSocket over TLS (WSS)**。

连接地址: `wss://{server_host}:{port}/ws`

### 5.2 消息格式

所有消息使用 JSON 编码，外层包裹 AES-256-GCM 加密。

### 5.3 消息类型

- 认证相关：MsgTypeAuth、MsgTypeAuthResponse
- 心跳：MsgTypeHeartbeat、MsgTypeHeartbeatAck
- 隧道控制：MsgTypeTunnelCreate、MsgTypeTunnelDelete、MsgTypeTunnelStart、MsgTypeTunnelStop
- 数据转发：MsgTypeDataOpen、MsgTypeDataClose、MsgTypeDataTransfer
- 配置：MsgTypeConfigPush、MsgTypeConfigPull
- 统计：MsgTypeStatsReport

---

## 6. 安全机制

### 6.1 认证安全

- 密码存储：Argon2id
- 会话管理：JWT (HS256, 2h) + Refresh Token (7d)
- 两步验证：TOTP (RFC 6238)
- 登录保护：同IP 5次失败后锁定15分钟

### 6.2 传输安全

- TLS 1.3（仅强密码套件）
- AES-256-GCM（每消息独立nonce）
- X25519密钥协商
- 重放攻击防护

### 6.3 数据安全

- 客户端配置本地加密存储
- 设备指纹派生密钥
- 配置导入导出加密打包

---

## 7. Docker 部署架构

详见 `docker/docker-compose.yml`

组件：
- Nginx (反向代理/TLS终止)
- API服务 (Go)
- 代理服务 (Go)
- PostgreSQL (数据库)
- Redis (缓存/限速)

---

## 8. 客户端模块划分

### 8.1 Tauri 客户端 (Windows/Linux)

- **Rust层**：系统交互、通知、自动更新、配置存储
- **Go层**：隧道引擎、协议实现、加密解密
- **React层**：用户界面

### 8.2 Flutter 客户端 (Android)

- **Dart层**：用户界面、状态管理
- **Go层**：隧道引擎（gomobile桥接）
- **Kotlin层**：Android原生功能（通知、后台服务）

---

## 9. 开发计划

1. **第一阶段**：服务端核心（1-2周）
2. **第二阶段**：Web管理界面（1-2周）
3. **第三阶段**：客户端开发（2-3周）
4. **第四阶段**：安全与伪装（1周）
5. **第五阶段**：高级功能（1周）
6. **第六阶段**：测试与文档（1周）

---

## 附录

### A. 环境变量配置

详见 `.env.example`

### B. 设备指纹生成算法

- Windows: Machine GUID + MAC + CPU ID + 主板序列号 + 磁盘序列号
- Linux: /etc/machine-id + MAC + CPU info + DMI 信息
- Android: Android ID + Build信息 + IMEI/MAC

### C. 流量限制算法

- 算法：令牌桶 (Token Bucket)
- 每个隧道独立令牌桶
- 令牌不足时拒绝或延迟
