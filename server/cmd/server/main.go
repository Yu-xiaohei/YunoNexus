package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yunonexus/server/internal/config"
	"github.com/yunonexus/server/internal/database"
	"github.com/yunonexus/server/internal/api/handler"
	"github.com/yunonexus/server/internal/api/middleware/auth"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 连接数据库
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 创建Echo实例
	e := echo.New()
	e.HideBanner = true

	// 中间件
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 健康检查
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API路由组
	api := e.Group("/api/v1")

	// 公开路由（无需认证）
	authHandler := handler.NewAuthHandler(db, cfg)
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/refresh", authHandler.RefreshToken)

	// 需要认证的路由
	protected := api.Group("")
	protected.Use(auth.JWTMiddleware(cfg.JWT.Secret))

	// 用户相关
	userHandler := handler.NewUserHandler(db)
	protected.GET("/users/me", userHandler.GetProfile)
	protected.PUT("/users/me", userHandler.UpdateProfile)

	// 设备相关
	deviceHandler := handler.NewDeviceHandler(db)
	protected.GET("/devices", deviceHandler.List)
	protected.GET("/devices/:id", deviceHandler.Get)
	protected.PUT("/devices/:id", deviceHandler.Update)
	protected.DELETE("/devices/:id", deviceHandler.Delete)

	// 隧道相关
	tunnelHandler := handler.NewTunnelHandler(db, cfg)
	protected.GET("/tunnels", tunnelHandler.List)
	protected.POST("/tunnels", tunnelHandler.Create)
	protected.GET("/tunnels/:id", tunnelHandler.Get)
	protected.PUT("/tunnels/:id", tunnelHandler.Update)
	protected.DELETE("/tunnels/:id", tunnelHandler.Delete)
	protected.POST("/tunnels/:id/start", tunnelHandler.Start)
	protected.POST("/tunnels/:id/stop", tunnelHandler.Stop)
	protected.GET("/tunnels/:id/stats", tunnelHandler.GetStats)

	// 流量统计
	trafficHandler := handler.NewTrafficHandler(db)
	protected.GET("/traffic/overview", trafficHandler.Overview)
	protected.GET("/traffic/tunnel/:id", trafficHandler.ByTunnel)
	protected.GET("/traffic/user/:id", trafficHandler.ByUser)

	// 配置导入导出
	configHandler := handler.NewConfigHandler(db)
	protected.GET("/config/export", configHandler.Export)
	protected.POST("/config/import", configHandler.Import)

	// 管理员路由
	admin := protected.Group("/admin")
	admin.Use(auth.AdminMiddleware())

	adminHandler := handler.NewAdminHandler(db)
	admin.GET("/users", adminHandler.ListUsers)
	admin.PUT("/users/:id", adminHandler.UpdateUser)
	admin.GET("/system/stats", adminHandler.SystemStats)
	admin.GET("/audit-logs", adminHandler.AuditLogs)
	admin.GET("/system/config", adminHandler.GetSystemConfig)
	admin.PUT("/system/config", adminHandler.UpdateSystemConfig)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	
	// 优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	log.Printf("YunoNexus API服务已启动: %s", addr)

	<-ctx.Done()
	log.Println("正在关闭服务器...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("服务器已关闭")
}
