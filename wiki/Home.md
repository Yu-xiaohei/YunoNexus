# YUNO Nexus

> 于的小窝 - 专属于你的轻量级内网穿透小工具
> 
> 版本: 0.2.0 | by Yu_xiaohei

## 简介

YUNO Nexus 是一个轻量级、高安全性的内网穿透工具，支持 TCP/UDP/HTTP/HTTPS/WebSocket 协议转发，让内网设备拥有公网访问能力。

## 文档导航

- [架构设计](Architecture.md) - 系统架构和技术栈
- [API文档](API-Reference.md) - 接口说明
- [错误码](Error-Codes.md) - 错误码体系
- [更新日志](CHANGELOG.md) - 版本历史

## 功能特性

- 支持 TCP/UDP/HTTP/HTTPS/WebSocket 协议转发
- 双层加密：TLS 1.3 + AES-256-GCM
- 设备指纹识别，防止账号盗用
- IP白名单、到期时间、带宽限制
- 流量伪装：HTTPS + WebSocket，模拟正常网页流量
- 多端支持：Windows、Linux、Android
- Web管理界面 + 实时仪表盘
- 自动更新、系统通知

## 快速开始

```bash
# 克隆项目
git clone https://github.com/Yu-xiaohei/YunoNexus.git
cd YunoNexus

# 配置环境变量
cp .env.example .env

# 启动服务
docker-compose up -d
```

访问 `https://dev.yxhmc.cn/nexus` 使用管理界面。

## 作者

**Yu_xiaohei**
