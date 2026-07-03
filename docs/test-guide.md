# YUNO Nexus 第一阶段 - 手动测试指南

> 版本: 0.1.0  
> 日期: 2026-07-04

---

## 测试前准备

### 启动服务

**终端1 - API服务：**
```powershell
cd F:\Yuno-Develop\YunoNexus\server
.\server.exe
```
看到 `YUNO Nexus Server 已启动: 0.0.0.0:8080` 表示成功

**终端2 - 代理服务：**
```powershell
cd F:\Yuno-Develop\YunoNexus\server
.\proxy.exe
```
看到 `代理服务已启动: :9443` 表示成功

### 测试工具

使用 PowerShell 终端，复制粘贴命令执行。

---

## 测试项目

### 1. 健康检查

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/health"
```

**预期结果：**
```json
{
  "status": "ok",
  "version": "release",
  "time": "2026-..."
}
```

---

### 2. 用户注册

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" -Method POST -ContentType "application/json" -Body '{"username":"testuser","email":"test@example.com","password":"Test1234!","device_name":"我的电脑","device_type":"windows","device_fingerprint":"abc123def456"}'
```

**预期结果：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user": { "id": "...", "username": "testuser", "email": "test@example.com" },
    "session": { "access_token": "...", "refresh_token": "..." },
    "device": { "id": "...", "device_name": "我的电脑" }
  }
}
```

---

### 3. 用户登录

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -ContentType "application/json" -Body '{"email":"test@example.com","password":"Test1234!","device_name":"我的电脑","device_type":"windows","device_fingerprint":"abc123def456"}'
```

**预期结果：**
- 返回 `code: 0`
- 包含 `access_token` 和 `refresh_token`

---

### 4. 登录失败锁定

**连续5次错误密码登录：**
```powershell
# 复制这行运行5次
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -ContentType "application/json" -Body '{"email":"test@example.com","password":"WrongPass1!","device_name":"我的电脑","device_type":"windows","device_fingerprint":"abc123def456"}'
```

**预期结果：**
- 前4次：返回 `code: -2599`，提示"您的账号或密码输入错误"
- 第5次：返回 `code: -2598`，提示"您的输入错误次数过多，请在 X 分钟后重试"

---

### 5. 认证访问

**先获取token：**
```powershell
$token = (Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -ContentType "application/json" -Body '{"email":"test@example.com","password":"Test1234!","device_name":"我的电脑","device_type":"windows","device_fingerprint":"abc123def456"}').data.session.access_token
```

**访问需要认证的接口：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users/me" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
```json
{
  "code": 0,
  "data": { "user_id": "..." }
}
```

---

### 6. 设备列表

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/devices" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
- 返回 `code: 0`
- `data.items` 包含设备数组
- `data.total` 显示设备总数

**记录返回的设备ID，后续测试会用到：**
```powershell
$deviceId = (Invoke-RestMethod -Uri "http://localhost:8080/api/v1/devices" -Headers @{"Authorization"="Bearer $token"}).data.items[0].id
Write-Host "设备ID: $deviceId"
```

---

### 7. 设备详情

**命令（用上面获取的真实ID）：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/devices/$deviceId" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
- 返回 `code: 0`
- 包含设备详细信息

---

### 8. 设备更新

**命令（替换DEVICE_ID）：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/devices/DEVICE_ID" -Method PUT -ContentType "application/json" -Headers @{"Authorization"="Bearer $token"} -Body '{"device_name":"更新后的名称"}'
```

**预期结果：**
- 返回 `code: 0`
- `device_name` 已更新

---

### 9. 创建隧道

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tunnels" -Method POST -ContentType "application/json" -Headers @{"Authorization"="Bearer $token"} -Body '{"name":"测试隧道","protocol":"tcp","local_host":"127.0.0.1","local_port":8080}'
```

**预期结果：**
```json
{
  "code": 0,
  "data": {
    "id": "...",
    "name": "测试隧道",
    "protocol": "tcp",
    "remote_port": 30000,
    "status": "inactive"
  }
}
```

**记录返回的隧道ID：**
```powershell
$tunnelId = (Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tunnels" -Method POST -ContentType "application/json" -Headers @{"Authorization"="Bearer $token"} -Body '{"name":"测试隧道","protocol":"tcp","local_host":"127.0.0.1","local_port":8080}').data.id
Write-Host "隧道ID: $tunnelId"
```

---

### 10. 隧道列表

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tunnels" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
- 返回 `code: 0`
- `data.items` 包含隧道数组

---

### 11. 隧道详情

**命令（用上面获取的真实ID）：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tunnels/$tunnelId" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
- 返回 `code: 0`
- 包含隧道详细信息

---

### 12. 启动隧道

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tunnels/$tunnelId/start" -Method POST -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
```json
{ "code": 0, "message": "success" }
```

---

### 13. 隧道统计

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tunnels/$tunnelId/stats" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
```json
{
  "code": 0,
  "data": {
    "status": "active",
    "protocol": "tcp",
    "remote_port": 30000,
    "bytes_sent": 0,
    "bytes_recv": 0
  }
}
```

---

### 14. 停止隧道

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tunnels/$tunnelId/stop" -Method POST -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
```json
{ "code": 0, "message": "success" }
```

---

### 15. 删除隧道

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tunnels/$tunnelId" -Method DELETE -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
```json
{ "code": 0, "message": "success" }
```

---

### 16. 流量概览

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/traffic/overview" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
```json
{
  "code": 0,
  "data": {
    "total_sent": 0,
    "total_recv": 0,
    "total_connections": 0
  }
}
```

---

### 17. 配置导出

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/config/export" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
- 返回 `code: 0`
- 包含 `version`、`user_id`、`tunnels` 等字段

---

### 18. 配置导入

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/config/import" -Method POST -ContentType "application/json" -Headers @{"Authorization"="Bearer $token"} -Body '{"config":"test"}'
```

**预期结果：**
```json
{
  "code": 0,
  "data": { "imported": 0, "message": "配置导入功能开发中" }
}
```

---

### 19. 管理员接口

**先设置管理员权限（数据库操作）：**
```powershell
docker exec docker-postgres-1 psql -U yunonexus -d yunonexus -c "UPDATE users SET role='admin' WHERE email='test@example.com';"
```

**管理员用户列表：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/users" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
- 返回 `code: 0`
- 包含用户列表

**系统统计：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/system/stats" -Headers @{"Authorization"="Bearer $token"}
```

**预期结果：**
```json
{
  "code": 0,
  "data": {
    "total_users": 1,
    "active_users": 1,
    "total_tunnels": 0,
    "active_tunnels": 0
  }
}
```

---

### 20. 错误处理

**未认证访问：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users/me"
```

**预期结果：**
- 返回 `code: 1001`，提示"缺少认证令牌"

**错误密码：**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -ContentType "application/json" -Body '{"email":"test@example.com","password":"WrongPass1!","device_name":"test","device_type":"windows","device_fingerprint":"test"}'
```

**预期结果：**
- 返回 `code: -2599`，提示"您的账号或密码输入错误"

---

### 21. 代理服务健康检查

**命令：**
```powershell
Invoke-RestMethod -Uri "http://localhost:9443/health"
```

**预期结果：**
```json
{ "status": "ok" }
```

---

## 测试记录表

| 序号 | 测试项 | 结果 | 备注 |
|------|--------|------|------|
| 1 | 健康检查 | | |
| 2 | 用户注册 | | |
| 3 | 用户登录 | | |
| 4 | 登录失败锁定 | | |
| 5 | 认证访问 | | |
| 6 | 设备列表 | | |
| 7 | 设备详情 | | |
| 8 | 设备更新 | | |
| 9 | 创建隧道 | | |
| 10 | 隧道列表 | | |
| 11 | 隧道详情 | | |
| 12 | 启动隧道 | | |
| 13 | 隧道统计 | | |
| 14 | 停止隧道 | | |
| 15 | 删除隧道 | | |
| 16 | 流量概览 | | |
| 17 | 配置导出 | | |
| 18 | 配置导入 | | |
| 19 | 管理员接口 | | |
| 20 | 错误处理 | | |
| 21 | 代理服务 | | |

---

## 注意事项

1. **token有效期**：access_token 2小时后过期，需要重新登录
2. **锁定时间**：连续5次错误密码后锁定5分钟
3. **重启清除**：服务重启后锁定记录会清除
4. **管理员权限**：管理员接口需要先设置 role='admin'
