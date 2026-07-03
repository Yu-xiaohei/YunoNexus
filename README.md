# YUNO Nexus

> 于的小窝 - 专属于你的轻量级内网穿透小工具
> 
> 版本: 0.1.0 | by Yu_xiaohei

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

- **服务端**: YUNO Nexus Server (Go + Echo + PostgreSQL + Redis)
- **Web管理**: React + TypeScript + Ant Design
- **桌面客户端**: YUNO Nexus (Tauri 2.x + Go)
- **Android客户端**: YUNO Nexus (Flutter 3.x + Go)
- **部署**: Docker + Nginx

## 快速开始

### 服务端部署

```bash
# 克隆项目
git clone https://github.com/Yu-xiaohei/YunoNexus.git
cd YunoNexus

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，设置数据库密码、JWT密钥等

# 启动服务
docker-compose up -d
```

### 访问管理界面

浏览器访问 `https://dev.yxhmc.cn/nexus`，使用管理员账号登录。

## 版本历史

### 0.1.0 (2026-07-03)
- 初始版本
- 项目结构搭建
- 服务端基础框架
- 数据库设计

## 文档

- [架构设计文档](docs/architecture.md)
- [完整架构文档](docs/architecture-full.md)
- [开发任务分解](docs/development-tasks.md)
- [API文档](docs/api/)
- [用户手册](docs/user-guide/)
- [开发指南](docs/dev-guide/)

## 安全说明

- 所有敏感配置使用 AES-256-GCM 加密存储
- 数据库连接使用 SSL
- API接口需要JWT认证
- 支持两步验证（TOTP）

## 版本号规则

格式: `n.n.n[-beta]`
- 第一位: 大版本（新功能组）
- 第二位: 小更新（功能完善/修改）
- 第三位: 每次变动+1
- 后缀: -beta 表示测试版

## 作者

**Yu_xiaohei**

## 许可证

MIT License
