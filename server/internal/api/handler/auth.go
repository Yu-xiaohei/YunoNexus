package handler

import (
	"fmt"
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
	Limiter      *auth.LoginLimiter
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *database.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		UserQueries:   database.NewUserQueries(db),
		DeviceQueries: database.NewDeviceQueries(db),
		Config:        cfg,
		Limiter:       auth.NewLoginLimiter(cfg.Security.LoginMaxAttempts, cfg.Security.LoginLockoutDuration),
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
		return response.ErrorWithCode(c, http.StatusBadRequest, 2001, "请求参数错误")
	}

	// 验证密码强度
	if err := validatePassword(req.Password); err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, 2002, err.Error())
	}

	// 生成密码哈希
	passwordHash, salt, err := auth.HashPassword(req.Password)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 2091, "系统出现未知错误，请刷新页面重试或联系管理员")
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
		return response.ErrorWithCode(c, http.StatusConflict, 2041, "用户名或邮箱已存在")
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
		return response.ErrorWithCode(c, http.StatusInternalServerError, 2092, "系统出现未知错误，请刷新页面重试或联系管理员")
	}

	// 生成令牌
	tokenPair, err := auth.GenerateTokenPair(
		user.ID, user.Username, user.Role,
		h.Config.JWT.Secret,
		h.Config.JWT.Expiry,
		h.Config.JWT.RefreshExpiry,
	)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusInternalServerError, 2093, "系统出现未知错误，请刷新页面重试或联系管理员")
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
		return response.ErrorWithCode(c, http.StatusBadRequest, 2001, "请求参数错误")
	}

	ctx := c.Request().Context()
	identifier := req.Email

	// 检查是否被锁定
	if locked, remain := h.Limiter.IsLocked(identifier); locked {
		return response.ErrorWithCode(c, http.StatusTooManyRequests, 2051,
			fmt.Sprintf("登录尝试过多，请在 %.0f 分钟后重试", remain.Minutes()))
	}

	// 获取用户
	user, err := h.UserQueries.GetByEmail(ctx, req.Email)
	if err != nil {
		h.Limiter.RecordFailure(identifier)
		return response.ErrorWithCode(c, http.StatusUnauthorized, 2011, "您的账号或密码输入错误")
	}

	// 验证密码
	ok, verifyErr := auth.VerifyPassword(req.Password, user.PasswordHash)
	if !ok || verifyErr != nil {
		h.Limiter.RecordFailure(identifier)
		failCount := h.Limiter.GetFailCount(identifier)
		remaining := h.Config.Security.LoginMaxAttempts - failCount
		if remaining <= 0 {
			return response.ErrorWithCode(c, http.StatusTooManyRequests, 2051,
				fmt.Sprintf("登录尝试过多，请在 %.0f 分钟后重试", h.Config.Security.LoginLockoutDuration.Minutes()))
		}
		return response.ErrorWithCode(c, http.StatusUnauthorized, 2011, "您的账号或密码输入错误")
	}

	// 密码正确，清除失败记录
	h.Limiter.RecordSuccess(identifier)

	// 检查用户状态
	if user.Status != "active" {
		return response.ErrorWithCode(c, http.StatusForbidden, 2061, "当前账号已被禁用")
	}

	// 检查两步验证
	if user.TwoFactorKey != nil && *user.TwoFactorKey != "" {
		if req.TOTPCode == "" {
			return response.ErrorWithCode(c, http.StatusUnauthorized, 2012, "当前账号需要进行令牌安全认证")
		}

		totp := auth.NewTOTP(*user.TwoFactorKey, "YUNO Nexus", user.Email)
		if !totp.VerifyCode(req.TOTPCode, 1) {
			return response.ErrorWithCode(c, http.StatusUnauthorized, 2013, "认证失败，您的令牌代码错误")
		}
	}

	// 获取或创建设备
	device, err := h.DeviceQueries.GetByFingerprint(ctx, user.ID, req.DeviceFingerprint)
	if err != nil {
		device = &models.Device{
			UserID:      user.ID,
			DeviceName:  req.DeviceName,
			DeviceType:  req.DeviceType,
			Fingerprint: req.DeviceFingerprint,
			Status:      "active",
		}
		if createErr := h.DeviceQueries.Create(ctx, device); createErr != nil {
			return response.ErrorWithCode(c, http.StatusInternalServerError, 2092, "系统出现未知错误，请刷新页面重试或联系管理员")
		}
	} else {
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
		return response.ErrorWithCode(c, http.StatusInternalServerError, 2093, "系统出现未知错误，请刷新页面重试或联系管理员")
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
		return response.ErrorWithCode(c, http.StatusBadRequest, 2001, "请求参数错误")
	}

	// 验证刷新令牌
	_, err := auth.ValidateToken(req.RefreshToken, h.Config.JWT.Secret)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusUnauthorized, 2014, "刷新令牌无效或已过期")
	}

	// 生成新的访问令牌
	newAccessToken, err := auth.RefreshAccessToken(
		req.RefreshToken,
		h.Config.JWT.Secret,
		h.Config.JWT.Expiry,
	)
	if err != nil {
		return response.ErrorWithCode(c, http.StatusUnauthorized, 2015, "刷新令牌失败，请重新登录")
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"access_token": newAccessToken,
	})
}

// 验证密码强度
func validatePassword(password string) error {
	if len(password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "您设置的密码过于简单，密码长度不得少于8位，且密码至少包含字母、数字、特殊符号中的两种")
	}

	var hasLetter, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'):
			hasLetter = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	// 至少包含两种类型（字母、数字、特殊符号）
	types := 0
	if hasLetter {
		types++
	}
	if hasDigit {
		types++
	}
	if hasSpecial {
		types++
	}

	if types < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "密码必须包含字母、数字、特殊符号中的至少两种")
	}

	return nil
}
