# YunoNexus 技术架构设计文档 (完整版)

> 版本: 1.0.0  
> 日期: 2026-07-03  
> 状态: 设计阶段

---

## 目录

1. [系统架构总览](#1-系统架构总览)
2. [数据库表结构](#2-数据库表结构)
3. [RESTful API 接口设计](#3-restful-api-接口设计)
4. [客户端-服务端通信协议](#4-客户端-服务端通信协议)
5. [安全机制](#5-安全机制)
6. [Docker 部署架构](#6-docker-部署架构)
7. [Go 服务端模块划分](#7-go-服务端模块划分)
8. [Tauri 客户端模块划分](#8-tauri-客户端模块划分)
9. [Flutter Android 客户端模块划分](#9-flutter-android-客户端模块划分)
10. [React Web 管理界面模块划分](#10-react-web-管理界面模块划分)

---

## 1. 系统架构总览

### 1.1 架构图（文本描述）

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
          │  (React SPA)   │  │  (Echo/Gin)     │
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
     │             │  │            │  │              │
     │ ┌─────────┐│  │ ┌────────┐ │  │ ┌──────────┐ │
     │ │Go隧道   ││  │ │Go隧道   │ │  │ │Go隧道    │ │
     │ │引擎     ││  │ │引擎    │ │  │ │引擎      │ │
     │ └─────────┘│  │ └────────┘ │  │ └──────────┘ │
     │ ┌─────────┐│  │ ┌────────┐ │  │ ┌──────────┐ │
     │ │Tauri UI ││  │ │Tauri UI│ │  │ │Flutter UI│ │
     │ └─────────┘│  │ └────────┘ │  │ └──────────┘ │
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
| ORM | sqlx / pgx | - |
| Web UI | React + TypeScript + Vite | React 18+ |
| UI 框架 | Ant Design 5 | - |
| Windows/Linux 客户端 | Tauri 2.x + Go (CGO) | - |
| Android 客户端 | Flutter 3.x + Go (gomobile) | - |
| 加密 | AES-256-GCM + zlib | - |
| 传输层 | WebSocket (gorilla/websocket) | - |
| 容器化 | Docker + Docker Compose | - |
| 反向代理 | Nginx | 1.25+ |

---

## 2. 数据库表结构

### 2.1 ER 关系图（文本）

```
users 1──N devices
users 1──N tunnels
tunnels 1──N port_bindings
tunnels 1──N traffic_logs
devices 1──N sessions
users 1──N audit_logs
```

### 2.2 完整表结构

#### 2.2.1 users（用户表）

```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(64) NOT NULL UNIQUE,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,        -- bcrypt / argon2id
    salt            VARCHAR(64) NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'user',  -- admin / user
    max_tunnels     INT NOT NULL DEFAULT 3,
    max_bandwidth   BIGINT NOT NULL DEFAULT 10485760,     -- bytes/sec, 默认 10MB/s
    status          VARCHAR(20) NOT NULL DEFAULT 'active', -- active / suspended / banned
    expires_at      TIMESTAMPTZ,                           -- 账号到期时间, NULL=永不过期
    two_factor_key  VARCHAR(255),                          -- TOTP 密钥 (加密存储)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
```

#### 2.2.2 devices（设备表）

```sql
CREATE TABLE devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name     VARCHAR(128) NOT NULL,
    device_type     VARCHAR(20) NOT NULL,          -- windows / linux / android
    fingerprint     VARCHAR(64) NOT NULL,           -- 设备指纹 (硬件特征哈希)
    os_version      VARCHAR(64),
    app_version     VARCHAR(20),
    public_key      TEXT,                           -- 设备公钥 (X25519)
    status          VARCHAR(20) NOT NULL DEFAULT 'active',  -- active / revoked
    ip_whitelist    INET[] DEFAULT '{}',            -- IP 白名单
    last_seen_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, fingerprint)
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_fingerprint ON devices(fingerprint);
```

#### 2.2.3 sessions（会话表）

```sql
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id       UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token           VARCHAR(64) NOT NULL UNIQUE,   -- JWT session token
    refresh_token   VARCHAR(64) NOT NULL UNIQUE,
    ip_address      INET NOT NULL,
    user_agent      TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_device_id ON sessions(device_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

#### 2.2.4 tunnels（隧道表）

```sql
CREATE TABLE tunnels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id       UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name            VARCHAR(128) NOT NULL,
    protocol        VARCHAR(10) NOT NULL,          -- tcp / udp / http / https / websocket
    local_host      VARCHAR(255) NOT NULL,          -- 客户端本地地址
    local_port      INT NOT NULL,
    remote_host     VARCHAR(255),                   -- 服务端监听地址 (NULL=自动分配)
    remote_port     INT NOT NULL,                   -- 服务端对外端口
    domain          VARCHAR(255),                   -- HTTP/HTTPS 绑定域名
    status          VARCHAR(20) NOT NULL DEFAULT 'inactive', -- inactive / active / error
    error_message   TEXT,
    config          JSONB DEFAULT '{}',             -- 协议特定配置 (加密存储)
    bandwidth_limit BIGINT,                         -- 该隧道带宽限制 (bytes/sec)
    traffic_used    BIGINT DEFAULT 0,               -- 已用流量 (bytes)
    traffic_limit   BIGINT,                         -- 流量限额 (bytes), NULL=无限
    expires_at      TIMESTAMPTZ,                    -- 隧道到期时间
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, remote_port),
    UNIQUE(user_id, domain),
    CONSTRAINT chk_protocol CHECK (protocol IN ('tcp','udp','http','https','websocket')),
    CONSTRAINT chk_ports CHECK (remote_port > 0 AND remote_port < 65536)
);

CREATE INDEX idx_tunnels_user_id ON tunnels(user_id);
CREATE INDEX idx_tunnels_device_id ON tunnels(device_id);
CREATE INDEX idx_tunnels_status ON tunnels(status);
```

#### 2.2.5 port_bindings（端口绑定表）

```sql
CREATE TABLE port_bindings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tunnel_id       UUID NOT NULL REFERENCES tunnels(id) ON DELETE CASCADE,
    port            INT NOT NULL UNIQUE,
    protocol        VARCHAR(10) NOT NULL,
    description     VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_port_bindings_port ON port_bindings(port);
```

#### 2.2.6 traffic_logs（流量日志表）

```sql
CREATE TABLE traffic_logs (
    id              BIGSERIAL PRIMARY KEY,
    tunnel_id       UUID NOT NULL REFERENCES tunnels(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id),
    bytes_sent      BIGINT NOT NULL DEFAULT 0,
    bytes_recv      BIGINT NOT NULL DEFAULT 0,
    connections     INT NOT NULL DEFAULT 0,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 按时间分区
CREATE INDEX idx_traffic_logs_tunnel_time ON traffic_logs(tunnel_id, recorded_at);
CREATE INDEX idx_traffic_logs_user_time ON traffic_logs(user_id, recorded_at);
```

#### 2.2.7 port_allocations（端口分配表）

```sql
CREATE TABLE port_allocations (
    port            INT PRIMARY KEY,
    allocated_to    UUID REFERENCES tunnels(id) ON DELETE SET NULL,
    allocated_by    UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### 2.2.8 audit_logs（审计日志表）

```sql
CREATE TABLE audit_logs (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    action          VARCHAR(64) NOT NULL,          -- login / tunnel_create / config_change ...
    resource_type   VARCHAR(32),                   -- user / tunnel / device
    resource_id     UUID,
    ip_address      INET,
    user_agent      TEXT,
    details         JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

#### 2.2.9 system_configs（系统配置表）

```sql
CREATE TABLE system_configs (
    key             VARCHAR(64) PRIMARY KEY,
    value           JSONB NOT NULL,
    description     TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by      UUID REFERENCES users(id)
);
```

#### 2.2.10 encryption_keys（加密密钥表）

```sql
CREATE TABLE encryption_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_type        VARCHAR(32) NOT NULL,          -- server_tls / api_jwt / tunnel_aes
    key_hash        VARCHAR(64) NOT NULL,          -- 密钥指纹
    public_key      TEXT,
    encrypted_key   TEXT,                          -- 加密存储的私钥
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotated_at      TIMESTAMPTZ
);

CREATE INDEX idx_encryption_keys_type ON encryption_keys(key_type);
```

---

## 3. RESTful API 接口设计

### 3.1 API 规范

- 基础路径: `/api/v1`
- 认证方式: Bearer Token (JWT)
- 内容类型: `application/json`
- 响应格式:

```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "timestamp": "2026-07-03T12:00:00Z"
}
```

- 分页格式:

```json
{
  "items": [...],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "has_more": true
}
```

- 错误码: `0` = 成功, `1xxx` = 认证错误, `2xxx` = 业务错误, `9xxx` = 系统错误

### 3.2 认证模块

#### POST /api/v1/auth/register
注册新用户
```
Request:
{
  "username": "string (3-64, alphanumeric_)",
  "email": "string",
  "password": "string (8-128, 复杂度要求)",
  "device_name": "string",
  "device_type": "windows|linux|android",
  "device_fingerprint": "string (sha256)"
}

Response 201:
{
  "code": 0,
  "data": {
    "user": { "id", "username", "email", "role", "created_at" },
    "session": { "token", "refresh_token", "expires_at" },
    "device": { "id", "device_name", "fingerprint" },
    "config_export": "加密的配置包 (base64)"
  }
}
```

#### POST /api/v1/auth/login
用户登录（支持多设备）
```
Request:
{
  "email": "string",
  "password": "string",
  "device_name": "string",
  "device_type": "windows|linux|android",
  "device_fingerprint": "string",
  "totp_code": "string (可选, 6位)"
}

Response 200:
{
  "code": 0,
  "data": {
    "session": { "token", "refresh_token", "expires_at", "refresh_expires_at" },
    "user": { "id", "username", "email", "role", "max_tunnels", "expires_at" },
    "device": { "id", "status" },
    "config_export": "string (已有配置则返回)"
  }
}
```

#### POST /api/v1/auth/refresh
刷新 Token
```
Request:
{
  "refresh_token": "string"
}

Response 200:
{
  "data": {
    "session": { "token", "refresh_token", "expires_at" }
  }
}
```

#### POST /api/v1/auth/logout
登出
```
Headers: Authorization: Bearer <token>
Response 200: { "code": 0 }
```

### 3.3 用户模块

#### GET /api/v1/users/me
获取当前用户信息
```
Response 200:
{
  "data": {
    "id", "username", "email", "role", "max_tunnels",
    "max_bandwidth", "status", "expires_at",
    "created_at", "updated_at"
  }
}
```

#### PUT /api/v1/users/me
更新当前用户信息
```
Request:
{
  "username": "string (可选)",
  "email": "string (可选)",
  "password": "string (可选, 需 old_password)"
}

Response 200: { "code": 0 }
```

#### POST /api/v1/users/me/2fa/enable
启用两步验证
```
Response 200:
{
  "data": {
    "secret": "TOTP密钥",
    "qr_code": "otpauth:// URI",
    "backup_codes": ["xxxx-xxxx", ...]
  }
}
```

#### POST /api/v1/users/me/2fa/verify
验证两步验证
```
Request: { "code": "6位数字" }
Response 200: { "code": 0 }
```

#### POST /api/v1/users/me/2fa/disable
禁用两步验证
```
Request: { "password": "string", "code": "6位数字" }
Response 200: { "code": 0 }
```

### 3.4 设备管理模块

#### GET /api/v1/devices
列出所有设备
```
Query: ?page=1&page_size=20&status=active

Response 200:
{
  "data": {
    "items": [
      {
        "id", "device_name", "device_type", "fingerprint",
        "os_version", "app_version", "status", "ip_whitelist",
        "last_seen_at", "created_at"
      }
    ],
    "total", "page", "page_size"
  }
}
```

#### GET /api/v1/devices/:id
获取设备详情

#### PUT /api/v1/devices/:id
更新设备
```
Request:
{
  "device_name": "string (可选)",
  "ip_whitelist": ["192.168.1.0/24", "10.0.0.1"]  (可选)
}
```

#### DELETE /api/v1/devices/:id
吊销设备
```
Response 200: { "code": 0 }
```

### 3.5 隧道管理模块

#### GET /api/v1/tunnels
列出隧道
```
Query: ?page=1&page_size=20&status=active&protocol=tcp&device_id=xxx

Response 200:
{
  "data": {
    "items": [
      {
        "id", "name", "protocol", "local_host", "local_port",
        "remote_host", "remote_port", "domain", "status",
        "bandwidth_limit", "traffic_used", "traffic_limit",
        "expires_at", "device_name", "created_at"
      }
    ],
    "total", "page", "page_size"
  }
}
```

#### POST /api/v1/tunnels
创建隧道
```
Request:
{
  "name": "string (必填)",
  "protocol": "tcp|udp|http|https|websocket",
  "local_host": "127.0.0.1",
  "local_port": 8080,
  "remote_port": 10080,          // 可选, 0=自动分配
  "domain": "app.example.com",   // http/https 必填
  "device_id": "uuid",           // 可选, 默认当前设备
  "bandwidth_limit": 1048576,    // 可选, bytes/sec
  "traffic_limit": 10737418240,  // 可选, bytes
  "expires_at": "2026-12-31T23:59:59Z"  // 可选
}

Response 201:
{
  "data": {
    "id", "name", "protocol", "remote_host", "remote_port",
    "domain", "status", "created_at"
  }
}
```

#### PUT /api/v1/tunnels/:id
更新隧道
```
Request: 同创建，所有字段可选
```

#### DELETE /api/v1/tunnels/:id
删除隧道

#### POST /api/v1/tunnels/:id/start
启动隧道

#### POST /api/v1/tunnels/:id/stop
停止隧道

#### GET /api/v1/tunnels/:id/stats
获取隧道实时统计
```
Response 200:
{
  "data": {
    "bytes_sent": 1234567,
    "bytes_recv": 9876543,
    "active_connections": 5,
    "uptime_seconds": 3600,
    "latency_ms": 12.5
  }
}
```

### 3.6 流量统计模块

#### GET /api/v1/traffic/overview
流量概览
```
Query: ?period=hour|day|week|month

Response 200:
{
  "data": {
    "total_sent": 1234567890,
    "total_recv": 9876543210,
    "total_connections": 15000,
    "period_start": "2026-07-01T00:00:00Z",
    "period_end": "2026-07-03T23:59:59Z",
    "daily_stats": [
      { "date": "2026-07-01", "sent": 1000000, "recv": 2000000 }
    ]
  }
}
```

#### GET /api/v1/traffic/tunnel/:id
按隧道查询流量
```
Query: ?period=day&start=2026-07-01&end=2026-07-03
```

#### GET /api/v1/traffic/user/:id
按用户查询流量（管理员）

### 3.7 系统管理模块（仅管理员）

#### GET /api/v1/admin/users
管理员查看所有用户
```
Query: ?page=1&page_size=20&status=active&search=keyword
```

#### PUT /api/v1/admin/users/:id
管理员更新用户
```
Request:
{
  "role": "admin|user",
  "status": "active|suspended|banned",
  "max_tunnels": 3,
  "max_bandwidth": 10485760,
  "expires_at": "2026-12-31T23:59:59Z"
}
```

#### GET /api/v1/admin/system/stats
系统统计
```
Response 200:
{
  "data": {
    "total_users": 150,
    "active_users": 120,
    "total_tunnels": 450,
    "active_tunnels": 300,
    "total_bandwidth_used": 123456789,
    "server_uptime": 864000,
    "db_size": "256MB"
  }
}
```

#### GET /api/v1/admin/audit-logs
审计日志
```
Query: ?page=1&page_size=50&action=login&user_id=xxx&start=2026-07-01&end=2026-07-03
```

#### GET /api/v1/admin/system/config
获取系统配置

#### PUT /api/v1/admin/system/config
更新系统配置
```
Request:
{
  "default_max_tunnels": 3,
  "default_max_bandwidth": 10485760,
  "allowed_ports_range": [10000, 60000],
  "session_timeout": 7200,
  "maintenance_mode": false
}
```

### 3.8 配置导入导出模块

#### GET /api/v1/config/export
导出客户端配置
```
Headers: Authorization: Bearer <token>
Query: ?device_id=xxx (可选, 默认当前设备)

Response 200:
{
  "data": {
    "config": "加密后的配置 (base64)",
    "format": "yunonexus-v1",
    "created_at": "2026-07-03T12:00:00Z"
  }
}
```

#### POST /api/v1/config/import
导入客户端配置
```
Request:
{
  "config": "加密的配置包 (base64)",
  "device_name": "新设备名称"
}

Response 200:
{
  "data": {
    "tunnels_imported": 3,
    "device_created": true
  }
}
```

---

## 4. 客户端-服务端通信协议

### 4.1 传输层

所有客户端-服务端通信使用 **WebSocket over TLS (WSS)**。

连接地址: `wss://{server_host}:{port}/ws`

### 4.2 消息格式

所有消息使用 JSON 编码，外层包裹 AES-256-GCM 加密：

```
┌──────────────────────────────────────────────┐
│ WebSocket Frame (binary)                     │
│ ┌──────────────────────────────────────────┐ │
│ │ Header (16 bytes)                        │ │
│ │ ┌──────┬──────┬──────┬──────┐           │ │
│ │ │ Type │ Flag │ MsgID│ Resvd│           │ │
│ │ │ 2B   │ 2B   │ 4B   │ 8B   │           │ │
│ │ └──────┴──────┴──────┴──────┘           │ │
│ ├──────────────────────────────────────────┤ │
│ │ Nonce (12 bytes)                         │ │
│ ├──────────────────────────────────────────┤ │
│ │ Encrypted Payload (variable)             │ │
│ │ ├──────────────────────────────────────┤ │ │
│ │ │ JSON Message Body                    │ │ │
│ │ └──────────────────────────────────────┘ │ │
│ ├──────────────────────────────────────────┤ │
│ │ Auth Tag (16 bytes)                      │ │
│ ├──────────────────────────────────────────┤ │
│ │ Original Size (4 bytes, 可选, 压缩时)    │ │
│ └──────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘
```

### 4.3 消息类型

```go
// 消息类型定义
const (
    // 控制消息
    MsgTypeAuth          uint16 = 0x0001  // 认证请求
    MsgTypeAuthResponse  uint16 = 0x0002  // 认证响应
    MsgTypeHeartbeat     uint16 = 0x0003  // 心跳
    MsgTypeHeartbeatAck  uint16 = 0x0004  // 心跳响应

    // 隧道控制
    MsgTypeTunnelCreate  uint16 = 0x0010  // 创建隧道
    MsgTypeTunnelDelete  uint16 = 0x0011  // 删除隧道
    MsgTypeTunnelStart   uint16 = 0x0012  // 启动隧道
    MsgTypeTunnelStop    uint16 = 0x0013  // 停止隧道
    MsgTypeTunnelStatus  uint16 = 0x0014  // 隧道状态更新
    MsgTypeTunnelError   uint16 = 0x001F  // 隧道错误

    // 数据转发
    MsgTypeDataOpen      uint16 = 0x0020  // 打开数据通道
    MsgTypeDataClose     uint16 = 0x0021  // 关闭数据通道
    MsgTypeDataTransfer  uint16 = 0x0022  // 数据传输
    MsgTypeDataPing      uint16 = 0x0023  // 数据通道心跳
    MsgTypeDataPong      uint16 = 0x0024  // 数据通道心跳响应

    // 配置
    MsgTypeConfigPush    uint16 = 0x0030  // 服务端推送配置
    MsgTypeConfigPull    uint16 = 0x0031  // 客户端拉取配置
    MsgTypeConfigSync    uint16 = 0x0032  // 配置同步确认

    // 统计
    MsgTypeStatsReport   uint16 = 0x0040  // 流量统计上报
    MsgTypeStatsQuery    uint16 = 0x0041  // 统计查询
)
```

### 4.4 消息结构

#### 认证消息

```json
{
  "type": 1,
  "msg_id": 1,
  "payload": {
    "token": "JWT session token",
    "device_id": "uuid",
    "device_fingerprint": "sha256 hash",
    "client_version": "1.0.0",
    "platform": "windows|linux|android",
    "capabilities": ["tcp", "udp", "http", "https", "websocket"]
  }
}
```

#### 认证响应

```json
{
  "type": 2,
  "msg_id": 1,
  "payload": {
    "success": true,
    "server_version": "1.0.0",
    "heartbeat_interval": 30,
    "tunnels": [
      {
        "id": "uuid",
        "name": "web-server",
        "protocol": "http",
        "local_host": "127.0.0.1",
        "local_port": 8080,
        "remote_port": 10080,
        "domain": "app.example.com",
        "bandwidth_limit": 1048576
      }
    ]
  }
}
```

#### 隧道创建请求

```json
{
  "type": 16,
  "msg_id": 10,
  "payload": {
    "tunnel": {
      "name": "my-web",
      "protocol": "http",
      "local_host": "127.0.0.1",
      "local_port": 8080,
      "remote_port": 0,
      "domain": "myapp.example.com",
      "bandwidth_limit": 1048576
    }
  }
}
```

#### 隧道创建响应

```json
{
  "type": 16,
  "msg_id": 10,
  "payload": {
    "success": true,
    "tunnel": {
      "id": "uuid",
      "remote_port": 10080,
      "remote_host": "server.example.com",
      "status": "active"
    }
  }
}
```

#### 数据通道打开

```json
{
  "type": 32,
  "msg_id": 100,
  "payload": {
    "tunnel_id": "uuid",
    "connection_id": "uuid",
    "protocol": "tcp",
    "source_addr": "203.0.113.1:54321",
    "target_addr": "127.0.0.1:8080"
  }
}
```

#### 数据传输

```json
// 实际传输使用二进制格式
// Header + ConnectionID(16B) + CompressedSize(4B) + OriginalSize(4B) + Data(variable)
```

### 4.5 数据转发流程

```
公网用户 → 服务端(端口监听) → 加密封装 → WSS隧道 → 客户端(解密) → 内网服务

1. 公网用户连接服务端 remote_port
2. 服务端创建新 connection_id，发送 DataOpen 消息给客户端
3. 客户端连接本地 local_host:local_port
4. 双向数据传输 (DataTransfer 消息, zlib 压缩 + AES-256-GCM 加密)
5. 任一端断开 → 发送 DataClose 消息
6. 服务端关闭公网连接
```

### 4.6 心跳机制

```
客户端 → 服务端: Heartbeat (每 30 秒)
服务端 → 客户端: HeartbeatAck (响应)
超时 90 秒无响应 → 判定离线，断开所有隧道
```

---

## 5. 安全机制

### 5.1 认证安全

| 机制 | 实现 |
|------|------|
| 密码存储 | Argon2id (memory=64MB, iterations=3, parallelism=4) |
| 会话管理 | JWT (RS256, 有效期 2h) + Refresh Token (有效期 7d) |
| 两步验证 | TOTP (RFC 6238, 6 位, 30s 间隔) |
| 设备指纹 | 硬件特征组合哈希 (SHA-256) |
| 登录保护 | 同 IP 5 次失败后锁定 15 分钟 |
| 密码策略 | 最少 8 位, 包含大小写+数字+特殊字符 |

### 5.2 传输安全

| 层级 | 机制 |
|------|------|
| TLS | TLS 1.3, 仅允许强密码套件 |
| 应用层加密 | AES-256-GCM (每消息独立 nonce) |
| 密钥协商 | X25519 (Curve25519 ECDH) |
| 消息完整性 | AES-GCM Auth Tag (16 字节) |
| 重放保护 | 消息 ID 递增 + 时间窗口检查 (30s) |

### 5.3 数据安全

```
密钥管理流程:

1. 服务端启动 → 生成 X25519 密钥对
2. 客户端连接 → 双方交换公钥
3. ECDH 计算共享密钥 → HKDF 派生会话密钥
4. 每条消息 → 生成随机 nonce (12B) → AES-256-GCM 加密
5. 数据压缩 → zlib 压缩 (level 6) 后再加密
```

### 5.4 网络安全

| 机制 | 说明 |
|------|------|
| IP 白名单 | 设备级 IP 白名单过滤 |
| 端口限制 | 仅开放必要端口, 端口范围可配置 |
| 连接限制 | 单设备最大连接数限制 |
| 带宽限制 | 用户级 + 隧道级双重带宽限制 |
| DDoS 防护 | 连接速率限制 + 连接数上限 |
| 流量伪装 | HTTPS + WebSocket 传输, 模拟正常 Web 流量 |

### 5.5 配置安全

```
客户端配置存储:

1. AES-256-GCM 加密配置文件
2. 密钥来源: 设备指纹派生 (HKDF)
3. 导入导出: 配置加密打包, 支持跨设备迁移
4. 配置包含: 服务器地址、token、隧道列表 (token 不加密存储于导出包)

存储路径:
- Windows: %APPDATA%/YunoNexus/config.enc
- Linux:   ~/.config/yunonexus/config.enc
- Android: 应用私有存储目录
```

### 5.6 审计日志

| 事件 | 记录内容 |
|------|---------|
| 登录/登出 | IP, 设备, 时间 |
| 隧道创建/删除 | 用户, 隧道信息, IP |
| 配置变更 | 变更内容, 操作人 |
| 权限变更 | 角色变更, 被操作人 |
| 异常事件 | 暴力破解尝试, 异常连接 |

---

## 6. Docker 部署架构

### 6.1 Docker Compose 架构

```yaml
version: '3.8'

services:
  # Nginx 反向代理
  nginx:
    image: nginx:1.25-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/conf.d:/etc/nginx/conf.d
      - ./nginx/ssl:/etc/nginx/ssl
      - ./web-dist:/usr/share/nginx/html:ro
    depends_on:
      - api
    restart: always
    networks:
      - yunonexus-net

  # Go API 服务
  api:
    build:
      context: ./server
      dockerfile: Dockerfile
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=yunonexus
      - DB_USER=yunonexus
      - DB_PASSWORD=${DB_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - REDIS_URL=redis://redis:6379
      - LOG_LEVEL=info
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: always
    networks:
      - yunonexus-net
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'

  # Go 隧道代理服务
  proxy:
    build:
      context: ./server
      dockerfile: Dockerfile.proxy
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=yunonexus
      - DB_USER=yunonexus
      - DB_PASSWORD=${DB_PASSWORD}
      - PROXY_PORT=9443
      - TUNNEL_PORT_START=30000
      - TUNNEL_PORT_END=30100
      - API_INTERNAL_URL=http://api:8080
    ports:
      - "9443:9443"     # WSS 客户端连接
      - "30000-30100:30000-30100"  # 隧道端口范围
    depends_on:
      postgres:
        condition: service_healthy
    restart: always
    networks:
      - yunonexus-net
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '2.0'

  # PostgreSQL
  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=yunonexus
      - POSTGRES_USER=yunonexus
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./db/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U yunonexus"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: always
    networks:
      - yunonexus-net
    deploy:
      resources:
        limits:
          memory: 512M

  # Redis (缓存/限速)
  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: always
    networks:
      - yunonexus-net

volumes:
  postgres-data:
  redis-data:

networks:
  yunonexus-net:
    driver: bridge
```

### 6.2 Nginx 配置

```nginx
# nginx/conf.d/default.conf

upstream api_backend {
    server api:8080;
}

server {
    listen 80;
    server_name _;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate     /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols       TLSv1.3;
    ssl_ciphers         TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256;
    ssl_prefer_server_ciphers on;

    # Web 管理界面
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }

    # API 接口
    location /api/ {
        proxy_pass http://api_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 限速
        limit_req zone=api burst=20 nodelay;
    }

    # WebSocket 客户端连接
    location /ws {
        proxy_pass http://proxy:9443;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}

# 限速区域定义
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
```

### 6.3 服务端 Dockerfile

```dockerfile
# Dockerfile (API 服务)
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080
ENTRYPOINT ["/app/server"]
```

```dockerfile
# Dockerfile.proxy (隧道代理)
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/proxy ./cmd/proxy

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/proxy /app/proxy

EXPOSE 9443
EXPOSE 30000-30100
ENTRYPOINT ["/app/proxy"]
```

---

## 7. Go 服务端模块划分

### 7.1 目录结构

```
server/
├── cmd/
│   ├── server/main.go          # API 服务入口
│   └── proxy/main.go           # 隧道代理入口
├── internal/
│   ├── config/
│   │   ├── config.go           # 配置加载 (环境变量/文件)
│   │   └── config_test.go
│   ├── database/
│   │   ├── db.go               # 数据库连接池
│   │   ├── migrations/         # SQL 迁移文件
│   │   └── queries/            # 预编译查询
│   │       ├── user.sql.go
│   │       ├── tunnel.sql.go
│   │       ├── device.sql.go
│   │       └── traffic.sql.go
│   ├── auth/
│   │   ├── jwt.go              # JWT 生成/验证
│   │   ├── password.go         # Argon2id 哈希
│   │   ├── totp.go             # TOTP 两步验证
│   │   └── middleware.go       # 认证中间件
│   ├── api/
│   │   ├── router.go           # 路由注册
│   │   ├── handler/
│   │   │   ├── auth.go         # 认证处理器
│   │   │   ├── user.go         # 用户管理
│   │   │   ├── device.go       # 设备管理
│   │   │   ├── tunnel.go       # 隧道管理
│   │   │   ├── traffic.go      # 流量统计
│   │   │   ├── config.go       # 配置导入导出
│   │   │   └── admin.go        # 管理员接口
│   │   ├── middleware/
│   │   │   ├── cors.go         # CORS
│   │   │   ├── ratelimit.go    # 速率限制
│   │   │   └── audit.go        # 审计日志
│   │   └── response/
│   │       └── response.go     # 统一响应格式
│   ├── proxy/
│   │   ├── server.go           # WebSocket 服务器
│   │   ├── client.go           # 客户端连接管理
│   │   ├── connection.go       # 单个连接抽象
│   │   ├── tunnel.go           # 隧道管理器
│   │   ├── protocol/
│   │   │   ├── message.go      # 消息结构定义
│   │   │   ├── codec.go        # 消息编解码
│   │   │   ├── encrypt.go      # AES-256-GCM 加密
│   │   │   └── compress.go     # zlib 压缩
│   │   ├── forwarder/
│   │   │   ├── tcp.go          # TCP 转发器
│   │   │   ├── udp.go          # UDP 转发器
│   │   │   ├── http.go         # HTTP 转发器
│   │   │   └── websocket.go    # WebSocket 转发器
│   │   └── dispatcher/
│   │       └── dispatcher.go   # 消息分发器
│   ├── security/
│   │   ├── fingerprint.go      # 设备指纹生成
│   │   ├── ipfilter.go         # IP 白名单过滤
│   │   └── crypto.go           # 密钥管理
│   └── util/
│       ├── logger.go           # 日志 (slog)
│       ├── errors.go           # 错误定义
│       └── helpers.go          # 工具函数
├── pkg/
│   ├── portalloc/
│   │   └── allocator.go        # 端口分配器
│   └── ratelimit/
│       └── limiter.go          # 令牌桶限流
├── migrations/
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── go.mod
└── go.sum
```

### 7.2 核心模块说明

#### proxy/server.go - WebSocket 服务器

```go
// 核心接口定义
type ProxyServer interface {
    // 启动 WebSocket 监听
    Listen(addr string) error
    
    // 处理新客户端连接
    HandleConnection(conn *websocket.Conn)
    
    // 广播消息给所有客户端
    Broadcast(msg *Message)
    
    // 获取在线客户端
    GetOnlineClients() []*Client
}
```

#### proxy/tunnel.go - 隧道管理器

```go
type TunnelManager interface {
    // 创建隧道监听
    CreateTunnel(tunnel *Tunnel) error
    
    // 停止隧道
    StopTunnel(tunnelID string) error
    
    // 处理隧道数据流
    HandleStream(tunnelID string, conn net.Conn) error
    
    // 获取隧道统计
    GetStats(tunnelID string) *TunnelStats
    
    // 带宽限速检查
    CheckBandwidth(tunnelID string) bool
}
```

#### proxy/protocol/encrypt.go - 加密模块

```go
type Encryptor interface {
    // 加密消息
    Encrypt(plaintext []byte) (*EncryptedMessage, error)
    
    // 解密消息
    Decrypt(msg *EncryptedMessage) ([]byte, error)
    
    // 使用 X25519 协商会话密钥
    KeyExchange(theirPublicKey []byte) error
}
```

---

## 8. Tauri 客户端模块划分

### 8.1 目录结构

```
client-tauri/
├── src-tauri/
│   ├── src/
│   │   ├── main.rs
│   │   ├── lib.rs
│   │   ├── tunnel/
│   │   │   ├── mod.rs
│   │   │   ├── manager.rs       # 隧道管理 (CGO 调用 Go)
│   │   │   ├── watcher.rs       # 隧道状态监控
│   │   │   └── bridge.rs        # Rust-Go 桥接层
│   │   ├── auth/
│   │   │   ├── mod.rs
│   │   │   ├── login.rs         # 登录逻辑
│   │   │   └── session.rs       # 会话管理
│   │   ├── config/
│   │   │   ├── mod.rs
│   │   │   ├── storage.rs       # 加密配置存储
│   │   │   └── import_export.rs # 导入导出
│   │   ├── notification/
│   │   │   ├── mod.rs
│   │   │   └── windows.rs       # Windows 通知
│   │   ├── updater/
│   │   │   ├── mod.rs
│   │   │   └── checker.rs       # 自动更新检查
│   │   ├── system/
│   │   │   ├── mod.rs
│   │   │   ├── tray.rs          # 系统托盘
│   │   │   └── autostart.rs     # 开机自启
│   │   └── state/
│   │       └── mod.rs           # 全局状态管理
│   ├── go/                      # Go 隧道引擎 (CGO)
│   │   ├── tunnel_engine.go     # 隧道核心逻辑
│   │   ├── protocol/            # 协议实现
│   │   │   ├── message.go
│   │   │   ├── encrypt.go
│   │   │   └── compress.go
│   │   ├── forwarder/           # 转发器
│   │   │   ├── tcp.go
│   │   │   ├── udp.go
│   │   │   ├── http.go
│   │   │   └── websocket.go
│   │   └── bridge/              # C 桥接
│   │       └── bridge.go        # CGO 导出函数
│   ├── Cargo.toml
│   └── tauri.conf.json
├── src/                          # 前端 (React)
│   ├── components/
│   │   ├── Dashboard.tsx         # 仪表盘
│   │   ├── TunnelList.tsx        # 隧道列表
│   │   ├── TunnelForm.tsx        # 隧道表单
│   │   ├── Login.tsx             # 登录页
│   │   ├── Settings.tsx          # 设置页
│   │   ├── DeviceInfo.tsx        # 设备信息
│   │   └── TrafficChart.tsx      # 流量图表
│   ├── hooks/
│   │   ├── useTunnel.ts          # 隧道操作
│   │   ├── useAuth.ts            # 认证状态
│   │   └── useStats.ts           # 统计数据
│   ├── services/
│   │   ├── api.ts                # API 调用
│   │   └── tauri.ts              # Tauri IPC 调用
│   ├── store/
│   │   └── index.ts              # 状态管理
│   ├── App.tsx
│   ├── main.tsx
│   └── index.html
└── package.json
```

### 8.2 Rust 模块职责

| 模块 | 职责 | 关键函数 |
|------|------|---------|
| tunnel/bridge.rs | Rust ↔ Go 桥接 | `start_tunnel`, `stop_tunnel`, `get_stats` |
| auth/login.rs | 登录/登出/刷新 | `login`, `logout`, `refresh_token` |
| config/storage.rs | 加密配置存储 | `save_config`, `load_config`, `encrypt_file` |
| notification/windows.rs | Windows 系统通知 | `show_notification(title, body)` |
| updater/checker.rs | 自动更新 | `check_update`, `download_update`, `apply_update` |
| system/tray.rs | 系统托盘 | `create_tray`, `update_tray_icon` |

### 8.3 Tauri 命令接口 (Rust → JS)

```rust
#[tauri::command]
async fn login(email: String, password: String, device_name: String) -> Result<Session, String>

#[tauri::command]
async fn logout() -> Result<(), String>

#[tauri::command]
async fn create_tunnel(config: TunnelConfig) -> Result<Tunnel, String>

#[tauri::command]
async fn delete_tunnel(id: String) -> Result<(), String>

#[tauri::command]
async fn start_tunnel(id: String) -> Result<(), String>

#[tauri::command]
async fn stop_tunnel(id: String) -> Result<(), String>

#[tauri::command]
async fn get_tunnels() -> Result<Vec<Tunnel>, String>

#[tauri::command]
async fn get_tunnel_stats(id: String) -> Result<TunnelStats, String>

#[tauri::command]
async fn export_config() -> Result<String, String>

#[tauri::command]
async fn import_config(data: String) -> Result<ImportResult, String>

#[tauri::command]
async fn check_update() -> Result<UpdateInfo, String>

#[tauri::command]
async fn get_device_info() -> Result<DeviceInfo, String>
```

---

## 9. Flutter Android 客户端模块划分

### 9.1 目录结构

```
client-android/
├── lib/
│   ├── main.dart
│   ├── app/
│   │   ├── app.dart              # MaterialApp 配置
│   │   ├── routes.dart           # 路由定义
│   │   └── theme.dart            # 主题配置
│   ├── core/
│   │   ├── go_bridge/
│   │   │   ├── go_bridge.dart    # Go 引擎 FFI 绑定
│   │   │   └── tunnel_engine.dart
│   │   ├── protocol/
│   │   │   ├── message.dart      # 消息结构
│   │   │   ├── encrypt.dart      # 加密模块
│   │   │   └── compress.dart     # 压缩模块
│   │   ├── forwarder/
│   │   │   ├── tcp_forwarder.dart
│   │   │   ├── udp_forwarder.dart
│   │   │   └── http_forwarder.dart
│   │   ├── network/
│   │   │   ├── api_client.dart   # HTTP 客户端
│   │   │   ├── ws_client.dart    # WebSocket 客户端
│   │   │   └── interceptors.dart # 拦截器
│   │   └── storage/
│   │       ├── secure_storage.dart # 加密存储
│   │       └── config_manager.dart
│   ├── features/
│   │   ├── auth/
│   │   │   ├── login_page.dart
│   │   │   ├── login_provider.dart
│   │   │   └── auth_service.dart
│   │   ├── dashboard/
│   │   │   ├── dashboard_page.dart
│   │   │   ├── dashboard_provider.dart
│   │   │   ├── stats_card.dart
│   │   │   └── traffic_chart.dart
│   │   ├── tunnels/
│   │   │   ├── tunnel_list_page.dart
│   │   │   ├── tunnel_form_page.dart
│   │   │   ├── tunnel_detail_page.dart
│   │   │   ├── tunnel_provider.dart
│   │   │   └── tunnel_card.dart
│   │   ├── settings/
│   │   │   ├── settings_page.dart
│   │   │   ├── device_info_page.dart
│   │   │   └── settings_provider.dart
│   │   └── config/
│   │       ├── config_import_page.dart
│   │       └── config_export_page.dart
│   ├── shared/
│   │   ├── widgets/
│   │   │   ├── status_badge.dart
│   │   │   ├── loading_overlay.dart
│   │   │   └── confirm_dialog.dart
│   │   └── utils/
│   │       ├── validators.dart
│   │       └── formatters.dart
│   └── services/
│       ├── notification_service.dart  # Android 通知
│       ├── connectivity_service.dart  # 网络状态
│       ├── update_service.dart        # 自动更新
│       └── background_service.dart    # 后台服务
├── android/
│   └── app/
│       └── src/main/
│           └── kotlin/
│               └── com/yunonexus/
│                   ├── MainActivity.kt
│                   ├── NotificationHelper.kt  # Android 通知管理
│                   ├── VpnService.kt          # VPN 服务 (可选)
│                   └── BootReceiver.kt        # 开机自启
├── go/                                # Go 引擎源码
│   ├── tunnel_engine.go
│   ├── protocol/
│   ├── forwarder/
│   └── gomobile/                      # gomobile 绑定
│       └── bind.go
├── pubspec.yaml
└── Makefile                           # gomobile build 脚本
```

### 9.2 Go-Flutter 桥接

```dart
// go_bridge.dart
class GoBridge {
  static const _channel = MethodChannel('yunonexus/go_bridge');
  
  /// 初始化 Go 引擎
  static Future<void> init() async {
    await _channel.invokeMethod('init');
  }
  
  /// 启动隧道
  static Future<void> startTunnel(String config) async {
    await _channel.invokeMethod('startTunnel', {'config': config});
  }
  
  /// 停止隧道
  static Future<void> stopTunnel(String tunnelId) async {
    await _channel.invokeMethod('stopTunnel', {'tunnelId': tunnelId});
  }
  
  /// 获取隧道状态
  static Future<String> getTunnelStatus(String tunnelId) async {
    final result = await _channel.invokeMethod('getTunnelStatus', {
      'tunnelId': tunnelId,
    });
    return result as String;
  }
  
  /// 获取统计数据
  static Future<String> getStats(String tunnelId) async {
    final result = await _channel.invokeMethod('getStats', {
      'tunnelId': tunnelId,
    });
    return result as String;
  }
  
  /// 设置回调（状态变化/统计更新）
  static void setCallback(GoBridgeCallback callback) {
    _channel.setMethodCallHandler((call) async {
      switch (call.method) {
        case 'onStatusChange':
          callback.onStatusChange(
            call.arguments['tunnelId'],
            call.arguments['status'],
          );
          break;
        case 'onStatsUpdate':
          callback.onStatsUpdate(
            call.arguments['tunnelId'],
            call.arguments['stats'],
          );
          break;
        case 'onError':
          callback.onError(
            call.arguments['tunnelId'],
            call.arguments['error'],
          );
          break;
      }
    });
  }
}

abstract class GoBridgeCallback {
  void onStatusChange(String tunnelId, String status);
  void onStatsUpdate(String tunnelId, String stats);
  void onError(String tunnelId, String error);
}
```

### 9.3 Android 原生模块

```kotlin
// NotificationHelper.kt
class NotificationHelper(private val context: Context) {
    
    companion object {
        private const val CHANNEL_ID = "yunonexus_tunnel"
        private const val CHANNEL_NAME = "隧道状态"
    }
    
    init {
        createNotificationChannel()
    }
    
    private fun createNotificationChannel() {
        val channel = NotificationChannel(
            CHANNEL_ID, CHANNEL_NAME, NotificationManager.IMPORTANCE_LOW
        ).apply {
            description = "隧道连接状态通知"
        }
        val manager = context.getSystemService(NotificationManager::class.java)
        manager.createNotificationChannel(channel)
    }
    
    fun showTunnelStatus(tunnelName: String, status: String) {
        val notification = NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_tunnel)
            .setContentTitle("隧道: $tunnelName")
            .setContentText("状态: $status")
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setOngoing(true)
            .build()
        
        NotificationManagerCompat.from(context).notify(tunnelName.hashCode(), notification)
    }
    
    fun showError(tunnelName: String, error: String) {
        val notification = NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_error)
            .setContentTitle("隧道错误: $tunnelName")
            .setContentText(error)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setAutoCancel(true)
            .build()
        
        NotificationManagerCompat.from(context).notify(tunnelName.hashCode() + 1000, notification)
    }
}
```

---

## 10. React Web 管理界面模块划分

### 10.1 目录结构

```
web-admin/
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── vite-env.d.ts
│   ├── api/
│   │   ├── client.ts             # Axios 实例 (拦截器/token 刷新)
│   │   ├── auth.ts               # 认证 API
│   │   ├── users.ts              # 用户 API
│   │   ├── devices.ts            # 设备 API
│   │   ├── tunnels.ts            # 隧道 API
│   │   ├── traffic.ts            # 流量 API
│   │   ├── admin.ts              # 管理 API
│   │   └── config.ts             # 配置 API
│   ├── components/
│   │   ├── layout/
│   │   │   ├── MainLayout.tsx     # 主布局 (侧边栏+顶栏)
│   │   │   ├── Sidebar.tsx        # 侧边栏导航
│   │   │   ├── Header.tsx         # 顶栏 (用户信息/通知)
│   │   │   └── Breadcrumb.tsx     # 面包屑
│   │   ├── common/
│   │   │   ├── StatusBadge.tsx    # 状态标签
│   │   │   ├── ConfirmDialog.tsx  # 确认对话框
│   │   │   ├── LoadingSpinner.tsx
│   │   │   ├── EmptyState.tsx
│   │   │   ├── ErrorBoundary.tsx
│   │   │   └── CopyButton.tsx     # 复制到剪贴板
│   │   ├── charts/
│   │   │   ├── TrafficLineChart.tsx    # 流量折线图
│   │   │   ├── TrafficBarChart.tsx     # 流量柱状图
│   │   │   ├── ConnectionGauge.tsx     # 连接数仪表盘
│   │   │   └── ProtocolPieChart.tsx    # 协议分布饼图
│   │   └── forms/
│   │       ├── TunnelForm.tsx     # 隧道创建/编辑表单
│   │       ├── UserForm.tsx       # 用户编辑表单
│   │       └── DeviceForm.tsx     # 设备编辑表单
│   ├── pages/
│   │   ├── Login.tsx              # 登录页
│   │   ├── Dashboard.tsx          # 仪表盘
│   │   ├── tunnels/
│   │   │   ├── TunnelList.tsx     # 隧道列表
│   │   │   └── TunnelDetail.tsx   # 隧道详情
│   │   ├── devices/
│   │   │   ├── DeviceList.tsx     # 设备列表
│   │   │   └── DeviceDetail.tsx   # 设备详情
│   │   ├── users/
│   │   │   ├── UserList.tsx       # 用户列表 (管理员)
│   │   │   └── UserDetail.tsx     # 用户详情 (管理员)
│   │   ├── traffic/
│   │   │   └── TrafficStats.tsx   # 流量统计
│   │   ├── settings/
│   │   │   ├── ProfileSettings.tsx  # 个人设置
│   │   │   └── SystemSettings.tsx   # 系统设置 (管理员)
│   │   └── audit/
│   │       └── AuditLogs.tsx      # 审计日志 (管理员)
│   ├── hooks/
│   │   ├── useAuth.ts             # 认证状态
│   │   ├── useTunnels.ts          # 隧道数据
│   │   ├── useDevices.ts          # 设备数据
│   │   ├── useTraffic.ts          # 流量数据
│   │   ├── useWebSocket.ts        # WebSocket 连接
│   │   └── usePagination.ts       # 分页逻辑
│   ├── store/
│   │   ├── authStore.ts           # 认证状态 (Zustand)
│   │   ├── tunnelStore.ts         # 隧道状态
│   │   └── uiStore.ts             # UI 状态
│   ├── types/
│   │   ├── api.ts                 # API 响应类型
│   │   ├── tunnel.ts              # 隧道类型
│   │   ├── user.ts                # 用户类型
│   │   └── device.ts              # 设备类型
│   ├── utils/
│   │   ├── format.ts              # 格式化工具
│   │   ├── validation.ts          # 表单验证
│   │   ├── storage.ts             # 本地存储
│   │   └── constants.ts           # 常量定义
│   └── styles/
│       ├── global.css             # 全局样式
│       └── theme.ts               # Ant Design 主题配置
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
└── .env
```

### 10.2 核心页面组件说明

#### Dashboard.tsx - 仪表盘

```
┌─────────────────────────────────────────────────────┐
│  仪表盘                                              │
├───────────┬───────────┬───────────┬─────────────────┤
│ 总用户数  │ 活跃隧道   │ 总流量    │  服务运行时间    │
│   150     │   45      │ 12.3 GB   │  7 天 12 小时    │
├───────────┴───────────┴───────────┴─────────────────┤
│                  流量趋势图 (7天/30天)                │
│  [折线图: 上行/下行流量]                               │
├──────────────────────────┬──────────────────────────┤
│    协议分布 (饼图)        │    连接数趋势 (柱状图)     │
├──────────────────────────┴──────────────────────────┤
│  最近活动                                            │
│  - 用户 xxx 创建了隧道 web-server (HTTP)             │
│  - 设备 xxx 上线                                     │
│  - 隧道 api-tunnel 带宽超过 80% 阈值                  │
└─────────────────────────────────────────────────────┘
```

#### TunnelList.tsx - 隧道列表

```
┌─────────────────────────────────────────────────────┐
│  隧道管理                    [+ 创建隧道] [筛选 ▼]   │
├─────────────────────────────────────────────────────┤
│ 名称      │ 协议  │ 远程端口 │ 状态  │ 流量    │ 操作│
├───────────┼───────┼─────────┼───────┼─────────┼─────┤
│ web-app   │ HTTPS │ 10043   │ ●在线  │ 2.3 GB  │ ⋮   │
│ api-svc   │ TCP   │ 10080   │ ●在线  │ 890 MB  │ ⋮   │
│ db-access │ TCP   │ 10543   │ ○离线  │ 0       │ ⋮   │
├───────────┴───────┴─────────┴───────┴─────────┴─────┤
│ 分页: < 1 2 3 ... 10 >    显示 1-20 / 共 45 条       │
└─────────────────────────────────────────────────────┘
```

### 10.3 技术选型

| 类别 | 选择 | 说明 |
|------|------|------|
| 框架 | React 18 | 函数式组件 + Hooks |
| 构建 | Vite 5 | 快速开发体验 |
| 语言 | TypeScript 5 | 类型安全 |
| UI 库 | Ant Design 5 | 企业级组件 |
| 图表 | @ant-design/charts | 流量可视化 |
| 状态管理 | Zustand | 轻量级状态管理 |
| 路由 | React Router 6 | SPA 路由 |
| HTTP | Axios | 请求拦截/token 刷新 |
| 表单 | React Hook Form | 表单验证 |
| 国际化 | i18next | 多语言支持 |
| WebSocket | socket.io-client | 实时数据推送 |

### 10.4 路由配置

```typescript
const routes = [
  { path: '/login', element: <Login /> },
  {
    path: '/',
    element: <MainLayout />,
    children: [
      { index: true, element: <Dashboard /> },
      { path: 'tunnels', element: <TunnelList /> },
      { path: 'tunnels/:id', element: <TunnelDetail /> },
      { path: 'devices', element: <DeviceList /> },
      { path: 'devices/:id', element: <DeviceDetail /> },
      { path: 'traffic', element: <TrafficStats /> },
      { path: 'settings', element: <ProfileSettings /> },
      // 管理员路由
      { path: 'admin/users', element: <UserList />, roles: ['admin'] },
      { path: 'admin/users/:id', element: <UserDetail />, roles: ['admin'] },
      { path: 'admin/system', element: <SystemSettings />, roles: ['admin'] },
      { path: 'admin/audit', element: <AuditLogs />, roles: ['admin'] },
    ],
  },
];
```

---

## 附录

### A. 环境变量配置

```bash
# 服务端配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=yunonexus
DB_USER=yunonexus
DB_PASSWORD=your_secure_password

JWT_SECRET=your_jwt_secret_key
JWT_EXPIRY=2h
REFRESH_EXPIRY=7d

REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=your_redis_password

# 代理配置
PROXY_PORT=9443
TUNNEL_PORT_RANGE=30000-30100
HEARTBEAT_INTERVAL=30

# 安全配置
LOGIN_MAX_ATTEMPTS=5
LOGIN_LOCKOUT_DURATION=900
SESSION_TIMEOUT=7200

# 日志
LOG_LEVEL=info
LOG_FORMAT=json
```

### B. 隧道端口分配策略

```
默认端口范围: 30000 - 30100 (100 个端口)
HTTP/HTTPS: 域名绑定, 无需端口
TCP/UDP: 从端口池分配
自动分配: 查找第一个未使用的端口
手动指定: 验证端口可用性后分配
```

### C. 流量限制算法

```
算法: 令牌桶 (Token Bucket)
参数:
  - rate: 每秒令牌数 (bytes/sec)
  - burst: 突发容量 (默认 2x rate)
实现:
  - 每个隧道独立令牌桶
  - 每条消息消耗对应字节数的令牌
  - 令牌不足时拒绝或延迟
```

### D. 设备指纹生成算法

```
Windows:
  - Machine GUID (注册表)
  - MAC 地址 (主网卡)
  - CPU ID
  - 主板序列号
  - 磁盘序列号
  SHA256(组合值) → 设备指纹

Linux:
  - /etc/machine-id
  - MAC 地址
  - CPU info
  - DMI 信息
  SHA256(组合值) → 设备指纹

Android:
  - Android ID
  - Build.BOARD + Build.DEVICE
  - Build.MANUFACTURER + Build.MODEL
  - IMEI (需权限) 或 MAC 地址
  SHA256(组合值) → 设备指纹
```
