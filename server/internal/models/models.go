package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Salt         string     `json:"-" db:"salt"`
	Role         string     `json:"role" db:"role"`
	MaxTunnels   int        `json:"max_tunnels" db:"max_tunnels"`
	MaxBandwidth int64      `json:"max_bandwidth" db:"max_bandwidth"`
	Status       string     `json:"status" db:"status"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
	TwoFactorKey *string    `json:"-" db:"two_factor_key"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Device 设备模型
type Device struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	DeviceName  string     `json:"device_name" db:"device_name"`
	DeviceType  string     `json:"device_type" db:"device_type"`
	Fingerprint string     `json:"fingerprint" db:"fingerprint"`
	OSVersion   *string    `json:"os_version" db:"os_version"`
	AppVersion  *string    `json:"app_version" db:"app_version"`
	PublicKey   *string    `json:"public_key" db:"public_key"`
	Status      string     `json:"status" db:"status"`
	IPWhitelist []string   `json:"ip_whitelist" db:"ip_whitelist"`
	LastSeenAt  *time.Time `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Session 会话模型
type Session struct {
	ID               string    `json:"id" db:"id"`
	DeviceID         string    `json:"device_id" db:"device_id"`
	UserID           string    `json:"user_id" db:"user_id"`
	Token            string    `json:"token" db:"token"`
	RefreshToken     string    `json:"refresh_token" db:"refresh_token"`
	IPAddress        string    `json:"ip_address" db:"ip_address"`
	UserAgent        string    `json:"user_agent" db:"user_agent"`
	ExpiresAt        time.Time `json:"expires_at" db:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at" db:"refresh_expires_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Tunnel 隧道模型
type Tunnel struct {
	ID             string     `json:"id" db:"id"`
	UserID         string     `json:"user_id" db:"user_id"`
	DeviceID       string     `json:"device_id" db:"device_id"`
	Name           string     `json:"name" db:"name"`
	Protocol       string     `json:"protocol" db:"protocol"`
	LocalHost      string     `json:"local_host" db:"local_host"`
	LocalPort      int        `json:"local_port" db:"local_port"`
	RemoteHost     string     `json:"remote_host" db:"remote_host"`
	RemotePort     int        `json:"remote_port" db:"remote_port"`
	Domain         string     `json:"domain" db:"domain"`
	Status         string     `json:"status" db:"status"`
	ErrorMessage   string     `json:"error_message" db:"error_message"`
	Config         string     `json:"config" db:"config"`
	BandwidthLimit *int64     `json:"bandwidth_limit" db:"bandwidth_limit"`
	TrafficUsed    int64      `json:"traffic_used" db:"traffic_used"`
	TrafficLimit   *int64     `json:"traffic_limit" db:"traffic_limit"`
	ExpiresAt      *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// AuditLog 审计日志模型
type AuditLog struct {
	ID           int64     `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	Details      string    `json:"details" db:"details"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// TrafficLog 流量日志模型
type TrafficLog struct {
	ID          int64     `json:"id" db:"id"`
	TunnelID    string    `json:"tunnel_id" db:"tunnel_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	BytesSent   int64     `json:"bytes_sent" db:"bytes_sent"`
	BytesRecv   int64     `json:"bytes_recv" db:"bytes_recv"`
	Connections int       `json:"connections" db:"connections"`
	RecordedAt  time.Time `json:"recorded_at" db:"recorded_at"`
}

// SystemConfig 系统配置模型
type SystemConfig struct {
	Key         string    `json:"key" db:"key"`
	Value       string    `json:"value" db:"value"`
	Description string    `json:"description" db:"description"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
}

// EncryptionKey 加密密钥模型
type EncryptionKey struct {
	ID           string     `json:"id" db:"id"`
	KeyType      string     `json:"key_type" db:"key_type"`
	KeyHash      string     `json:"key_hash" db:"key_hash"`
	PublicKey    string     `json:"public_key" db:"public_key"`
	EncryptedKey string     `json:"encrypted_key" db:"encrypted_key"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	RotatedAt    *time.Time `json:"rotated_at" db:"rotated_at"`
}
