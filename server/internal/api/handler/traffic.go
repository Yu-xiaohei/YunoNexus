package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yunonexus/server/internal/database"
	"github.com/yunonexus/server/internal/api/response"
)

// TrafficHandler 流量统计处理器
type TrafficHandler struct {
	TrafficQueries *database.TrafficQueries
}

// NewTrafficHandler 创建流量统计处理器
func NewTrafficHandler(db *database.DB) *TrafficHandler {
	return &TrafficHandler{
		TrafficQueries: database.NewTrafficQueries(db),
	}
}

// Overview 流量概览
func (h *TrafficHandler) Overview(c echo.Context) error {
	userID := c.Get("user_id").(string)

	stats, err := h.TrafficQueries.GetUserStats(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "查询流量统计失败")
	}

	return response.Success(c, http.StatusOK, stats)
}

// ByTunnel 按隧道查询流量
func (h *TrafficHandler) ByTunnel(c echo.Context) error {
	userID := c.Get("user_id").(string)
	tunnelID := c.Param("id")

	// 验证隧道归属
	tunnelQueries := database.NewTunnelQueries(h.TrafficQueries.DB)
	tunnel, err := tunnelQueries.GetByID(c.Request().Context(), tunnelID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "隧道不存在")
	}

	if tunnel.UserID != userID {
		return response.Error(c, http.StatusForbidden, "无权访问此隧道")
	}

	stats, err := h.TrafficQueries.GetTunnelStats(c.Request().Context(), tunnelID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "查询流量统计失败")
	}

	return response.Success(c, http.StatusOK, stats)
}

// ByUser 按用户查询流量（管理员）
func (h *TrafficHandler) ByUser(c echo.Context) error {
	targetUserID := c.Param("id")

	stats, err := h.TrafficQueries.GetUserStats(c.Request().Context(), targetUserID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "查询流量统计失败")
	}

	return response.Success(c, http.StatusOK, stats)
}
