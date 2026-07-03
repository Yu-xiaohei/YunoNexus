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
	echoAuth "github.com/yunonexus/server/internal/api/middleware/auth"
)

func main() {
	cfg := config.Load()

	// 连接数据库
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 健康检查
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "ok",
			"version": cfg.Server.Mode,
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	api := e.Group("/api/v1")

	// 公开路由
	authHandler := handler.NewAuthHandler(db, cfg)
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/refresh", authHandler.RefreshToken)

	// 需要认证的路由
	protected := api.Group("")
	protected.Use(echoAuth.JWTMiddleware(cfg.JWT.Secret))

	// 用户信息
	protected.GET("/users/me", func(c echo.Context) error {
		userID := c.Get("user_id").(string)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    map[string]string{"user_id": userID},
		})
	})

	// 隧道管理
	tunnelHandler := handler.NewTunnelHandler(db, cfg)
	protected.GET("/tunnels", tunnelHandler.List)
	protected.POST("/tunnels", tunnelHandler.Create)
	protected.GET("/tunnels/:id", tunnelHandler.Get)
	protected.PUT("/tunnels/:id", tunnelHandler.Update)
	protected.DELETE("/tunnels/:id", tunnelHandler.Delete)
	protected.POST("/tunnels/:id/start", tunnelHandler.Start)
	protected.POST("/tunnels/:id/stop", tunnelHandler.Stop)
	protected.GET("/tunnels/:id/stats", tunnelHandler.GetStats)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	log.Printf("YUNO Nexus Server 已启动: %s", addr)

	<-ctx.Done()
	log.Println("正在关闭...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	e.Shutdown(shutdownCtx)
	log.Println("已关闭")
}
