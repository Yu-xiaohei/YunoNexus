package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yunonexus/server/internal/database"
	"github.com/yunonexus/server/internal/api/response"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	UserQueries   *database.UserQueries
	TunnelQueries *database.TunnelQueries
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(db *database.DB) *AdminHandler {
	return &AdminHandler{
		UserQueries:   database.NewUserQueries(db),
		TunnelQueries: database.NewTunnelQueries(db),
	}
}

// ListUsers 用户列表（管理员）
func (h *AdminHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	users, total, err := h.UserQueries.List(c.Request().Context(), offset, pageSize)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "查询用户失败")
	}

	return response.PaginatedSuccess(c, users, total, page, pageSize)
}

// UpdateUser 更新用户（管理员）
func (h *AdminHandler) UpdateUser(c echo.Context) error {
	userID := c.Param("id")

	var req struct {
		Role         string `json:"role"`
		Status       string `json:"status"`
		MaxTunnels   *int   `json:"max_tunnels"`
		MaxBandwidth *int64 `json:"max_bandwidth"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "请求参数错误")
	}

	ctx := c.Request().Context()

	user, err := h.UserQueries.GetByID(ctx, userID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "用户不存在")
	}

	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != "" {
		user.Status = req.Status
	}
	if req.MaxTunnels != nil {
		user.MaxTunnels = *req.MaxTunnels
	}
	if req.MaxBandwidth != nil {
		user.MaxBandwidth = *req.MaxBandwidth
	}

	if err := h.UserQueries.Update(ctx, user); err != nil {
		return response.Error(c, http.StatusInternalServerError, "更新用户失败")
	}

	return response.Success(c, http.StatusOK, user)
}

// SystemStats 系统统计
func (h *AdminHandler) SystemStats(c echo.Context) error {
	ctx := c.Request().Context()

	// 获取用户总数
	users, userTotal, err := h.UserQueries.List(ctx, 0, 1)
	if err != nil {
		userTotal = 0
	}
	_ = users

	// 获取隧道总数
	// 简化处理，后续可以优化
	stats := map[string]interface{}{
		"total_users":    userTotal,
		"active_users":   userTotal,
		"total_tunnels":  0,
		"active_tunnels": 0,
	}

	return response.Success(c, http.StatusOK, stats)
}

// AuditLogs 审计日志
func (h *AdminHandler) AuditLogs(c echo.Context) error {
	// TODO: 实现审计日志查询
	return response.Success(c, http.StatusOK, map[string]interface{}{
		"items":  []interface{}{},
		"total":  0,
		"page":   1,
		"page_size": 20,
	})
}

// GetSystemConfig 获取系统配置
func (h *AdminHandler) GetSystemConfig(c echo.Context) error {
	// TODO: 实现系统配置查询
	return response.Success(c, http.StatusOK, map[string]interface{}{
		"default_max_tunnels":   3,
		"default_max_bandwidth": 10485760,
		"session_timeout":       7200,
	})
}

// UpdateSystemConfig 更新系统配置
func (h *AdminHandler) UpdateSystemConfig(c echo.Context) error {
	// TODO: 实现系统配置更新
	return response.Success(c, http.StatusOK, nil)
}
