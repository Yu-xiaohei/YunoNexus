package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yunonexus/server/internal/auth"
	"github.com/yunonexus/server/internal/config"
	"github.com/yunonexus/server/internal/database"
	"github.com/yunonexus/server/internal/models"
	"github.com/yunonexus/server/internal/api/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	UserQueries  *database.UserQueries
	DeviceQueries *database.DeviceQueries
	Config       *config.Config
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *database.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		UserQueries:  database.NewUserQueries(db),
		DeviceQueries: database.NewDeviceQueries(db),
		Config:       cfg,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username          string `json:"username" validate:"required,min=3,max=64"`
	Email             string `json:"email" validate:"required,email"`
	Password          string `json:"password" validate:"required,min=8,max=128"`
	DeviceName        string `json:"device_name" validate:"required"`
	DeviceType        string `json:"device_type" validate:"required,oneof=windows linux android"`
	DeviceFingerprint string `json:"device_fingerprint" validate:"required"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email             string `json:"email" validate:"required,email"`
	Password          string `json:"password" validate:"required"`
	DeviceName        string `json:"device_name" validate:"required"`
	DeviceType        string `json:"device_type" validate:"required,oneof=windows linux android"`
	DeviceFingerprint string `json:"device_fingerprint" validate:"required"`
	TOTPCode          string `json:"totp_code"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	User    *models.User    `json:"user"`
	Session *auth.TokenPair `json:"session"`
	Device  *models.Device  `json:"device"`
}

// Register 用户注册
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "请求参数错误")
	}

	// 验证密码强度
	if err := validatePassword(req.Password); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	// 生成密码哈希
	passwordHash, salt, err := auth.HashPassword(req.Password)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "生成密码哈希失败")
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         "user",
		MaxTunnels:   3,
		MaxBandwidth: 10485760, // 10MB/s
		Status:       "active",
	}

	ctx := c.Request().Context()
	if err := h.UserQueries.Create(ctx, user); err != nil {
		return response.Error(c, http.StatusConflict, "用户名或邮箱已存在")
	}

	// 创建设备
	device := &models.Device{
		UserID:      user.ID,
		DeviceName:  req.DeviceName,
		DeviceType:  req.DeviceType,
		Fingerprint: req.DeviceFingerprint,
		Status:      "active",
	}

	if err := h.DeviceQueries.Create(ctx, device); err != nil {
		return response.Error(c, http.StatusInternalServerError, "创建设备失败")
	}

	// 生成令牌
	tokenPair, err := auth.GenerateTokenPair(
		user.ID, user.Username, user.Role,
		h.Config.JWT.Secret,
		h.Config.JWT.Expiry,
		h.Config.JWT.RefreshExpiry,
	)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "生成令牌失败")
	}

	return response.Success(c, http.StatusCreated, &AuthResponse{
		User:    user,
		Session: tokenPair,
		Device:  device,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "请求参数错误")
	}

	ctx := c.Request().Context()

	// 获取用户
	user, err := h.UserQueries.GetByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("登录失败 - 获取用户错误: %v", err)
		return response.Error(c, http.StatusUnauthorized, "邮箱或密码错误")
	}

	log.Printf("登录 - 用户ID: %s, 邮箱: %s", user.ID, user.Email)

	// 验证密码
	ok, verifyErr := auth.VerifyPassword(req.Password, user.PasswordHash)
	log.Printf("登录 - 密码验证结果: ok=%v, err=%v", ok, verifyErr)
	if !ok || verifyErr != nil {
		return response.Error(c, http.StatusUnauthorized, "邮箱或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return response.Error(c, http.StatusForbidden, "账号已被禁用")
	}

	// 检查两步验证
	if user.TwoFactorKey != nil && *user.TwoFactorKey != "" {
		if req.TOTPCode == "" {
			return response.Error(c, http.StatusUnauthorized, "需要两步验证")
		}

		totp := auth.NewTOTP(*user.TwoFactorKey, "YUNO Nexus", user.Email)
		if !totp.VerifyCode(req.TOTPCode, 1) {
			return response.Error(c, http.StatusUnauthorized, "两步验证代码错误")
		}
	}

	// 获取或创建设备
	device, err := h.DeviceQueries.GetByFingerprint(ctx, user.ID, req.DeviceFingerprint)
	if err != nil {
		log.Printf("登录 - 设备不存在，尝试创建: %v", err)
		// 设备不存在，创建新设备
		device = &models.Device{
			UserID:      user.ID,
			DeviceName:  req.DeviceName,
			DeviceType:  req.DeviceType,
			Fingerprint: req.DeviceFingerprint,
			Status:      "active",
		}
		if createErr := h.DeviceQueries.Create(ctx, device); createErr != nil {
			log.Printf("登录 - 创建设备失败: %v", createErr)
			return response.Error(c, http.StatusInternalServerError, "创建设备失败")
		}
		log.Printf("登录 - 设备创建成功: %s", device.ID)
	} else {
		log.Printf("登录 - 设备已存在: %s", device.ID)
		// 更新设备信息
		now := time.Now()
		device.LastSeenAt = &now
		h.DeviceQueries.UpdateLastSeen(ctx, device.ID)
	}

	// 生成令牌
	tokenPair, err := auth.GenerateTokenPair(
		user.ID, user.Username, user.Role,
		h.Config.JWT.Secret,
		h.Config.JWT.Expiry,
		h.Config.JWT.RefreshExpiry,
	)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "生成令牌失败")
	}

	return response.Success(c, http.StatusOK, &AuthResponse{
		User:    user,
		Session: tokenPair,
		Device:  device,
	})
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "请求参数错误")
	}

	// 验证刷新令牌
	_, err := auth.ValidateToken(req.RefreshToken, h.Config.JWT.Secret)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "无效的刷新令牌")
	}

	// 生成新的访问令牌
	newAccessToken, err := auth.RefreshAccessToken(
		req.RefreshToken,
		h.Config.JWT.Secret,
		h.Config.JWT.Expiry,
	)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "刷新令牌失败")
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"access_token": newAccessToken,
	})
}

// 验证密码强度
func validatePassword(password string) error {
	if len(password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "密码长度至少8位")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return echo.NewHTTPError(http.StatusBadRequest, "密码必须包含大小写字母、数字和特殊字符")
	}

	return nil
}
