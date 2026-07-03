package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yunonexus/server/internal/database"
	"github.com/yunonexus/server/internal/api/response"
)

// DeviceHandler 设备处理器
type DeviceHandler struct {
	DeviceQueries *database.DeviceQueries
}

// NewDeviceHandler 创建设备处理器
func NewDeviceHandler(db *database.DB) *DeviceHandler {
	return &DeviceHandler{
		DeviceQueries: database.NewDeviceQueries(db),
	}
}

// List 设备列表
func (h *DeviceHandler) List(c echo.Context) error {
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

	devices, total, err := h.DeviceQueries.List(c.Request().Context(), userID, offset, pageSize)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 4091, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	return response.PaginatedSuccess(c, devices, total, page, pageSize)
}

// Get 获取设备详情
func (h *DeviceHandler) Get(c echo.Context) error {
	userID := c.Get("user_id").(string)
	deviceID := c.Param("id")

	device, err := h.DeviceQueries.GetByID(c.Request().Context(), deviceID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusNotFound, 4031, "设备不存在")
	}

	if device.UserID != userID {
		return response.ErrorWithCode(c, http.StatusForbidden, 4021, "当前用户权限不足")
	}

	return response.Success(c, http.StatusOK, device)
}

// Update 更新设备
func (h *DeviceHandler) Update(c echo.Context) error {
	userID := c.Get("user_id").(string)
	deviceID := c.Param("id")

	var req struct {
		DeviceName string `json:"device_name"`
	}

	if err := c.Bind(&req); err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, 4001, "请求参数错误")
	}

	ctx := c.Request().Context()

	device, err := h.DeviceQueries.GetByID(ctx, deviceID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusNotFound, 4031, "设备不存在")
	}

	if device.UserID != userID {
		return response.ErrorWithCode(c, http.StatusForbidden, 4022, "当前用户权限不足")
	}

	if req.DeviceName != "" {
		device.DeviceName = req.DeviceName
		// 直接执行SQL更新
		_, err = h.DeviceQueries.DB.Pool.Exec(ctx, 
			"UPDATE devices SET device_name = $1, updated_at = NOW() WHERE id = $2", 
			device.DeviceName, device.ID)
		if err != nil {
			return response.ErrorWithCode(c, http.StatusInternalServerError, 4092, "系统出现未知错误，请刷新页面重试或联系管理员")
		}
	}

	return response.Success(c, http.StatusOK, device)
}

// Delete 吊销设备
func (h *DeviceHandler) Delete(c echo.Context) error {
	userID := c.Get("user_id").(string)
	deviceID := c.Param("id")

	ctx := c.Request().Context()

	device, err := h.DeviceQueries.GetByID(ctx, deviceID)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusNotFound, 4031, "设备不存在")
	}

	if device.UserID != userID {
		return response.ErrorWithCode(c, http.StatusForbidden, 4023, "当前用户权限不足")
	}

	if err := h.DeviceQueries.Delete(ctx, deviceID); err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 4093, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	return response.Success(c, http.StatusOK, nil)
}
