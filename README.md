# YunoNexus

轻量级、高安全性的内网穿透工具

## 功能特性

- 支持 TCP/UDP/HTTP/HTTPS/WebSocket 协议转发
- 双层加密：TLS 1.3 + AES-256-GCM
- 设备指纹识别，防止账号盗用
- IP白名单、到期时间、带宽限制
- 流量伪装：HTTPS + WebSocket，模拟正常网页流量
- 多端支持：Windows、Linux、Android
- Web管理界面 + 实时仪表盘
- 自动更新、系统通知

## 技术栈

- **服务端**: Go + Echo + PostgreSQL + Redis
- **Web管理**: React + TypeScript + Ant Design
- **桌面客户端**: Tauri 2.x + Go
- **Android客户端**: Flutter 3.x + Go
- **部署**: Docker + Nginx

## 快速开始

### 服务端部署

```bash
# 克隆项目
git clone https://github.com/your-username/YunoNexus.git
cd YunoNexus

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，设置数据库密码、JWT密钥等

# 启动服务
docker-compose up -d
```

### 访问管理界面

浏览器访问 `https://your-server-ip`，使用管理员账号登录。

## 文档

- [架构设计文档](docs/architecture.md)
- [API文档](docs/api/)
- [用户手册](docs/user-guide/)
- [开发指南](docs/dev-guide/)

## 安全说明

- 所有敏感配置使用 AES-256-GCM 加密存储
- 数据库连接使用 SSL
- API接口需要JWT认证
- 支持两步验证（TOTP）

## 许可证

MIT License
