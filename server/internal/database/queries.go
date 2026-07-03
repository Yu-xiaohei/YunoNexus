package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/yunonexus/server/internal/models"
)

// UserQueries 用户相关查询
type UserQueries struct {
	DB *DB
}

// NewUserQueries 创建用户查询实例
func NewUserQueries(db *DB) *UserQueries {
	return &UserQueries{DB: db}
}

// Create 创建用户
func (q *UserQueries) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, salt, role, max_tunnels, max_bandwidth, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return q.DB.Pool.QueryRow(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.Salt,
		user.Role, user.MaxTunnels, user.MaxBandwidth, user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID 根据ID获取用户
func (q *UserQueries) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, salt, role, max_tunnels, max_bandwidth, 
		       status, expires_at, two_factor_key, created_at, updated_at
		FROM users WHERE id = $1`

	user := &models.User{}
	err := q.DB.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Salt,
		&user.Role, &user.MaxTunnels, &user.MaxBandwidth, &user.Status,
		&user.ExpiresAt, &user.TwoFactorKey, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("用户不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (q *UserQueries) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, salt, role, max_tunnels, max_bandwidth, 
		       status, expires_at, two_factor_key, created_at, updated_at
		FROM users WHERE email = $1`

	user := &models.User{}
	err := q.DB.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Salt,
		&user.Role, &user.MaxTunnels, &user.MaxBandwidth, &user.Status,
		&user.ExpiresAt, &user.TwoFactorKey, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("用户不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// Update 更新用户
func (q *UserQueries) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET username = $2, email = $3, role = $4, max_tunnels = $5, 
		    max_bandwidth = $6, status = $7, expires_at = $8, updated_at = NOW()
		WHERE id = $1`

	_, err := q.DB.Pool.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.Role,
		user.MaxTunnels, user.MaxBandwidth, user.Status, user.ExpiresAt,
	)

	return err
}

// UpdatePassword 更新密码
func (q *UserQueries) UpdatePassword(ctx context.Context, userID, passwordHash, salt string) error {
	query := `UPDATE users SET password_hash = $2, salt = $3, updated_at = NOW() WHERE id = $1`
	_, err := q.DB.Pool.Exec(ctx, query, userID, passwordHash, salt)
	return err
}

// UpdateTwoFactor 更新两步验证
func (q *UserQueries) UpdateTwoFactor(ctx context.Context, userID, twoFactorKey string) error {
	query := `UPDATE users SET two_factor_key = $2, updated_at = NOW() WHERE id = $1`
	_, err := q.DB.Pool.Exec(ctx, query, userID, twoFactorKey)
	return err
}

// List 列出用户
func (q *UserQueries) List(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	// 获取总数
	var total int
	err := q.DB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询用户列表
	query := `
		SELECT id, username, email, role, max_tunnels, max_bandwidth, 
		       status, expires_at, created_at, updated_at
		FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	rows, err := q.DB.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Role,
			&user.MaxTunnels, &user.MaxBandwidth, &user.Status,
			&user.ExpiresAt, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

// DeviceQueries 设备相关查询
type DeviceQueries struct {
	DB *DB
}

// NewDeviceQueries 创建设备查询实例
func NewDeviceQueries(db *DB) *DeviceQueries {
	return &DeviceQueries{DB: db}
}

// Create 创建设备
func (q *DeviceQueries) Create(ctx context.Context, device *models.Device) error {
	query := `
		INSERT INTO devices (user_id, device_name, device_type, fingerprint, os_version, app_version, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return q.DB.Pool.QueryRow(ctx, query,
		device.UserID, device.DeviceName, device.DeviceType,
		device.Fingerprint, device.OSVersion, device.AppVersion, device.Status,
	).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)
}

// GetByID 根据ID获取设备
func (q *DeviceQueries) GetByID(ctx context.Context, id string) (*models.Device, error) {
	query := `
		SELECT id, user_id, device_name, device_type, fingerprint, os_version, 
		       app_version, public_key, status, ip_whitelist, last_seen_at, created_at, updated_at
		FROM devices WHERE id = $1`

	device := &models.Device{}
	err := q.DB.Pool.QueryRow(ctx, query, id).Scan(
		&device.ID, &device.UserID, &device.DeviceName, &device.DeviceType,
		&device.Fingerprint, &device.OSVersion, &device.AppVersion,
		&device.PublicKey, &device.Status, &device.IPWhitelist,
		&device.LastSeenAt, &device.CreatedAt, &device.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("设备不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	return device, nil
}

// GetByFingerprint 根据指纹获取设备
func (q *DeviceQueries) GetByFingerprint(ctx context.Context, userID, fingerprint string) (*models.Device, error) {
	query := `
		SELECT id, user_id, device_name, device_type, fingerprint, os_version, 
		       app_version, public_key, status, ip_whitelist, last_seen_at, created_at, updated_at
		FROM devices WHERE user_id = $1 AND fingerprint = $2`

	device := &models.Device{}
	err := q.DB.Pool.QueryRow(ctx, query, userID, fingerprint).Scan(
		&device.ID, &device.UserID, &device.DeviceName, &device.DeviceType,
		&device.Fingerprint, &device.OSVersion, &device.AppVersion,
		&device.PublicKey, &device.Status, &device.IPWhitelist,
		&device.LastSeenAt, &device.CreatedAt, &device.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("设备不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	return device, nil
}

// UpdateLastSeen 更新最后在线时间
func (q *DeviceQueries) UpdateLastSeen(ctx context.Context, deviceID string) error {
	query := `UPDATE devices SET last_seen_at = $2, updated_at = NOW() WHERE id = $1`
	_, err := q.DB.Pool.Exec(ctx, query, deviceID, time.Now())
	return err
}

// List 列出设备
func (q *DeviceQueries) List(ctx context.Context, userID string, offset, limit int) ([]*models.Device, int, error) {
	// 获取总数
	var total int
	err := q.DB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM devices WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询设备列表
	query := `
		SELECT id, user_id, device_name, device_type, fingerprint, os_version, 
		       app_version, status, last_seen_at, created_at
		FROM devices 
		WHERE user_id = $1
		ORDER BY last_seen_at DESC NULLS LAST
		LIMIT $2 OFFSET $3`

	rows, err := q.DB.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		device := &models.Device{}
		err := rows.Scan(
			&device.ID, &device.UserID, &device.DeviceName, &device.DeviceType,
			&device.Fingerprint, &device.OSVersion, &device.AppVersion,
			&device.Status, &device.LastSeenAt, &device.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		devices = append(devices, device)
	}

	return devices, total, nil
}

// Delete 删除设备
func (q *DeviceQueries) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM devices WHERE id = $1`
	_, err := q.DB.Pool.Exec(ctx, query, id)
	return err
}
