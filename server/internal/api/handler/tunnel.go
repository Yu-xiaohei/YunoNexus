package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yunonexus/server/internal/config"
	"github.com/yunonexus/server/internal/database"
	"github.com/yunonexus/server/internal/models"
	"github.com/yunonexus/server/internal/api/response"
)

// TunnelHandler 隧道处理器
type TunnelHandler struct {
	TunnelQueries  *database.TunnelQueries
	UserQueries    *database.UserQueries
	DeviceQueries  *database.DeviceQueries
	Config         *config.Config
}

// NewTunnelHandler 创建隧道处理器
func NewTunnelHandler(db *database.DB, cfg *config.Config) *TunnelHandler {
	return &TunnelHandler{
		TunnelQueries: database.NewTunnelQueries(db),
		UserQueries:   database.NewUserQueries(db),
		DeviceQueries: database.NewDeviceQueries(db),
		Config:        cfg,
	}
}

// CreateTunnelRequest 创建隧道请求
type CreateTunnelRequest struct {
	Name           string  `json:"name" validate:"required"`
	Protocol       string  `json:"protocol" validate:"required,oneof=tcp udp http https websocket"`
	LocalHost      string  `json:"local_host" validate:"required"`
	LocalPort      int     `json:"local_port" validate:"required,min=1,max=65535"`
	RemotePort     int     `json:"remote_port"`
	Domain         string  `json:"domain"`
	DeviceID       string  `json:"device_id"`
	BandwidthLimit *int64  `json:"bandwidth_limit"`
	TrafficLimit   *int64  `json:"traffic_limit"`
}

// List 隧道列表
func (h *TunnelHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(string)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	tunnels, total, err := h.TunnelQueries.List(c.Request().Context(), userID, offset, pageSize)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 5091, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	return response.PaginatedSuccess(c, tunnels, total, page, pageSize)
}

// Create 创建隧道
func (h *TunnelHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req CreateTunnelRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, 5001, "请求参数错误")
	}

	ctx := c.Request().Context()

	// 获取设备ID
	deviceID := req.DeviceID
	if deviceID == "" {
		// 获取用户第一个设备
		devices, _, err := h.DeviceQueries.List(ctx, userID, 0, 1)
		if err != nil || len(devices) == 0 {
			return response.ErrorWithCode(c, http.StatusBadRequest, 5082, "请先注册设备")
		}
		deviceID = devices[0].ID
	}

	// 检查用户隧道数量限制
	user, err := h.UserQueries.GetByID(ctx, userID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusUnauthorized, 5011, "用户不存在")
	}

	count, err := h.TunnelQueries.CountByUser(ctx, userID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 5091, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	if count >= user.MaxTunnels {
		return response.ErrorWithCode(c, http.StatusForbidden, 5081, 
			"已达隧道数量上限（最多 "+strconv.Itoa(user.MaxTunnels)+" 个）")
	}

	// 分配端口
	remotePort := req.RemotePort
	if remotePort == 0 {
		port, err := h.TunnelQueries.FindAvailablePort(ctx, 
			h.Config.Security.AllowedPortsRange[0], 
			h.Config.Security.AllowedPortsRange[1])
		if err != nil {
			return response.ErrorWithCode(c, http.StatusServiceUnavailable, 5071, "没有可用端口")
		}
		remotePort = port
	}

	// 创建隧道
	var domain *string
	if req.Domain != "" {
		domain = &req.Domain
	}
	tunnel := &models.Tunnel{
		UserID:         userID,
		DeviceID:       deviceID,
		Name:           req.Name,
		Protocol:       req.Protocol,
		LocalHost:      req.LocalHost,
		LocalPort:      req.LocalPort,
		RemotePort:     remotePort,
		Domain:         domain,
		Status:         "inactive",
		BandwidthLimit: req.BandwidthLimit,
		TrafficLimit:   req.TrafficLimit,
	}

	if err := h.TunnelQueries.Create(ctx, tunnel); err != nil {
		log.Printf("创建隧道失败: %v", err)
		return response.ErrorWithCode(c, http.StatusInternalServerError, 5092, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	// 分配端口记录
	if err := h.TunnelQueries.AllocatePort(ctx, remotePort, tunnel.ID, userID); err != nil {
		// 端口分配失败不影响隧道创建
	}

	return response.Success(c, http.StatusCreated, tunnel)
}

// Get 获取隧道详情
func (h *TunnelHandler) Get(c echo.Context) error {
	userID := c.Get("user_id").(string)
	tunnelID := c.Param("id")

	tunnel, err := h.TunnelQueries.GetByID(c.Request().Context(), tunnelID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusNotFound, 5031, "隧道不存在")
	}

	if tunnel.UserID != userID {
		return response.ErrorWithCode(c, http.StatusForbidden, 5021, "当前用户权限不足")
	}

	return response.Success(c, http.StatusOK, tunnel)
}

// Update 更新隧道
func (h *TunnelHandler) Update(c echo.Context) error {
	userID := c.Get("user_id").(string)
	tunnelID := c.Param("id")

	ctx := c.Request().Context()

	tunnel, err := h.TunnelQueries.GetByID(ctx, tunnelID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusNotFound, 5031, "隧道不存在")
	}

	if tunnel.UserID != userID {
		return response.ErrorWithCode(c, http.StatusForbidden, 5022, "当前用户权限不足")
	}

	var req CreateTunnelRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, 5001, "请求参数错误")
	}

	var domain *string
	if req.Domain != "" {
		domain = &req.Domain
	}

	tunnel.Name = req.Name
	tunnel.Protocol = req.Protocol
	tunnel.LocalHost = req.LocalHost
	tunnel.LocalPort = req.LocalPort
	tunnel.Domain = domain
	tunnel.BandwidthLimit = req.BandwidthLimit
	tunnel.TrafficLimit = req.TrafficLimit

	if err := h.TunnelQueries.Update(ctx, tunnel); err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 5093, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	return response.Success(c, http.StatusOK, tunnel)
}

// Delete 删除隧道
func (h *TunnelHandler) Delete(c echo.Context) error {
	userID := c.Get("user_id").(string)
	tunnelID := c.Param("id")

	ctx := c.Request().Context()

	tunnel, err := h.TunnelQueries.GetByID(ctx, tunnelID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusNotFound, 5031, "隧道不存在")
	}

	if tunnel.UserID != userID {
		return response.ErrorWithCode(c, http.StatusForbidden, 5023, "当前用户权限不足")
	}

	// 释放端口
	h.TunnelQueries.ReleasePort(ctx, tunnel.RemotePort)

	if err := h.TunnelQueries.Delete(ctx, tunnelID); err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 5094, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	return response.Success(c, http.StatusOK, nil)
}

// Start 启动隧道
func (h *TunnelHandler) Start(c echo.Context) error {
	userID := c.Get("user_id").(string)
	tunnelID := c.Param("id")

	ctx := c.Request().Context()

	tunnel, err := h.TunnelQueries.GetByID(ctx, tunnelID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusNotFound, 5031, "隧道不存在")
	}

	if tunnel.UserID != userID {
		return response.ErrorWithCode(c, http.StatusForbidden, 5024, "当前用户权限不足")
	}

	if err := h.TunnelQueries.UpdateStatus(ctx, tunnelID, "active", ""); err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 5095, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	return response.Success(c, http.StatusOK, nil)
}

// Stop 停止隧道
func (h *TunnelHandler) Stop(c echo.Context) error {
	userID := c.Get("user_id").(string)
	tunnelID := c.Param("id")

	ctx := c.Request().Context()

	tunnel, err := h.TunnelQueries.GetByID(ctx, tunnelID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusNotFound, 5031, "隧道不存在")
	}

	if tunnel.UserID != userID {
		return response.ErrorWithCode(c, http.StatusForbidden, 5024, "当前用户权限不足")
	}

	if err := h.TunnelQueries.UpdateStatus(ctx, tunnelID, "inactive", ""); err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 5096, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	return response.Success(c, http.StatusOK, nil)
}

// GetStats 获取隧道统计
func (h *TunnelHandler) GetStats(c echo.Context) error {
	userID := c.Get("user_id").(string)
	tunnelID := c.Param("id")

	tunnel, err := h.TunnelQueries.GetByID(c.Request().Context(), tunnelID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusNotFound, 5031, "隧道不存在")
	}

	if tunnel.UserID != userID {
		return response.ErrorWithCode(c, http.StatusForbidden, 5021, "当前用户权限不足")
	}

	stats := map[string]interface{}{
		"bytes_sent":    tunnel.TrafficUsed,
		"bytes_recv":    0,
		"status":        tunnel.Status,
		"protocol":      tunnel.Protocol,
		"remote_port":   tunnel.RemotePort,
	}

	return response.Success(c, http.StatusOK, stats)
}
