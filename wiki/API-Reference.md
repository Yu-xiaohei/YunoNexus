# API 接口文档

## 基础信息

- 基础路径: `/api/v1`
- 认证方式: Bearer Token (JWT)
- 内容类型: `application/json`

## 认证接口

### POST /api/v1/auth/register
注册新用户

### POST /api/v1/auth/login
用户登录

### POST /api/v1/auth/refresh
刷新Token

## 用户接口

### GET /api/v1/users/me
获取当前用户信息

## 设备接口

### GET /api/v1/devices
设备列表

### GET /api/v1/devices/:id
设备详情

### PUT /api/v1/devices/:id
更新设备

### DELETE /api/v1/devices/:id
吊销设备

## 隧道接口

### GET /api/v1/tunnels
隧道列表

### POST /api/v1/tunnels
创建隧道

### GET /api/v1/tunnels/:id
隧道详情

### PUT /api/v1/tunnels/:id
更新隧道

### DELETE /api/v1/tunnels/:id
删除隧道

### POST /api/v1/tunnels/:id/start
启动隧道

### POST /api/v1/tunnels/:id/stop
停止隧道

### GET /api/v1/tunnels/:id/stats
隧道统计

## 流量接口

### GET /api/v1/traffic/overview
流量概览

### GET /api/v1/traffic/tunnel/:id
按隧道查询流量

## 配置接口

### GET /api/v1/config/export
导出配置

### POST /api/v1/config/import
导入配置

## 管理员接口

### GET /api/v1/admin/users
用户列表

### PUT /api/v1/admin/users/:id
更新用户

### GET /api/v1/admin/system/stats
系统统计

### GET /api/v1/admin/audit-logs
审计日志

### GET /api/v1/admin/system/config
系统配置

### PUT /api/v1/admin/system/config
更新系统配置
