package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yunonexus/server/internal/database"
	"github.com/yunonexus/server/internal/api/response"
)

// ConfigHandler 配置导入导出处理器
type ConfigHandler struct {
	TunnelQueries *database.TunnelQueries
	DeviceQueries *database.DeviceQueries
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(db *database.DB) *ConfigHandler {
	return &ConfigHandler{
		TunnelQueries: database.NewTunnelQueries(db),
		DeviceQueries: database.NewDeviceQueries(db),
	}
}

// Export 导出配置
func (h *ConfigHandler) Export(c echo.Context) error {
	userID := c.Get("user_id").(string)

	// 获取用户所有隧道
	tunnels, _, err := h.TunnelQueries.List(c.Request().Context(), userID, 0, 100)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "查询隧道失败")
	}

	// 构建导出数据
	exportData := map[string]interface{}{
		"version":  "1.0",
		"user_id":  userID,
		"tunnels":  tunnels,
		"exported_at": "2026-07-03",
	}

	return response.Success(c, http.StatusOK, exportData)
}

// Import 导入配置
func (h *ConfigHandler) Import(c echo.Context) error {
	var req struct {
		Config string `json:"config" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "请求参数错误")
	}

	// TODO: 解析配置并导入隧道
	// 目前简化处理

	return response.Success(c, http.StatusOK, map[string]interface{}{
		"imported": 0,
		"message":  "配置导入功能开发中",
	})
}
