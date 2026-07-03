package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/yunonexus/server/internal/models"
)

// TunnelQueries 隧道相关查询
type TunnelQueries struct {
	DB *DB
}

// NewTunnelQueries 创建隧道查询实例
func NewTunnelQueries(db *DB) *TunnelQueries {
	return &TunnelQueries{DB: db}
}

// Create 创建隧道
func (q *TunnelQueries) Create(ctx context.Context, tunnel *models.Tunnel) error {
	query := `
		INSERT INTO tunnels (user_id, device_id, name, protocol, local_host, local_port, 
		                     remote_host, remote_port, domain, status, config, bandwidth_limit, traffic_limit, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	config := "{}"
	if tunnel.Config != nil && *tunnel.Config != "" {
		config = *tunnel.Config
	}

	return q.DB.Pool.QueryRow(ctx, query,
		tunnel.UserID, tunnel.DeviceID, tunnel.Name, tunnel.Protocol,
		tunnel.LocalHost, tunnel.LocalPort, tunnel.RemoteHost, tunnel.RemotePort,
		tunnel.Domain, tunnel.Status, config, tunnel.BandwidthLimit, tunnel.TrafficLimit, tunnel.ExpiresAt,
	).Scan(&tunnel.ID, &tunnel.CreatedAt, &tunnel.UpdatedAt)
}

// GetByID 根据ID获取隧道
func (q *TunnelQueries) GetByID(ctx context.Context, id string) (*models.Tunnel, error) {
	query := `
		SELECT id, user_id, device_id, name, protocol, local_host, local_port, 
		       remote_host, remote_port, domain, status, error_message, bandwidth_limit,
		       traffic_used, traffic_limit, expires_at, created_at, updated_at
		FROM tunnels WHERE id = $1`

	tunnel := &models.Tunnel{}
	err := q.DB.Pool.QueryRow(ctx, query, id).Scan(
		&tunnel.ID, &tunnel.UserID, &tunnel.DeviceID, &tunnel.Name, &tunnel.Protocol,
		&tunnel.LocalHost, &tunnel.LocalPort, &tunnel.RemoteHost, &tunnel.RemotePort,
		&tunnel.Domain, &tunnel.Status, &tunnel.ErrorMessage, &tunnel.BandwidthLimit,
		&tunnel.TrafficUsed, &tunnel.TrafficLimit, &tunnel.ExpiresAt,
		&tunnel.CreatedAt, &tunnel.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("隧道不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询隧道失败: %w", err)
	}

	return tunnel, nil
}

// List 列出用户隧道
func (q *TunnelQueries) List(ctx context.Context, userID string, offset, limit int) ([]*models.Tunnel, int, error) {
	var total int
	err := q.DB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM tunnels WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, device_id, name, protocol, local_host, local_port, 
		       remote_host, remote_port, domain, status, bandwidth_limit,
		       traffic_used, traffic_limit, expires_at, created_at
		FROM tunnels 
		WHERE user_id = $1
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := q.DB.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tunnels []*models.Tunnel
	for rows.Next() {
		t := &models.Tunnel{}
		err := rows.Scan(
			&t.ID, &t.UserID, &t.DeviceID, &t.Name, &t.Protocol,
			&t.LocalHost, &t.LocalPort, &t.RemoteHost, &t.RemotePort,
			&t.Domain, &t.Status, &t.BandwidthLimit, &t.TrafficUsed,
			&t.TrafficLimit, &t.ExpiresAt, &t.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		tunnels = append(tunnels, t)
	}

	return tunnels, total, nil
}

// Update 更新隧道
func (q *TunnelQueries) Update(ctx context.Context, tunnel *models.Tunnel) error {
	query := `
		UPDATE tunnels 
		SET name = $2, protocol = $3, local_host = $4, local_port = $5,
		    remote_host = $6, remote_port = $7, domain = $8, status = $9,
		    bandwidth_limit = $10, traffic_limit = $11, expires_at = $12, updated_at = NOW()
		WHERE id = $1`

	_, err := q.DB.Pool.Exec(ctx, query,
		tunnel.ID, tunnel.Name, tunnel.Protocol, tunnel.LocalHost, tunnel.LocalPort,
		tunnel.RemoteHost, tunnel.RemotePort, tunnel.Domain, tunnel.Status,
		tunnel.BandwidthLimit, tunnel.TrafficLimit, tunnel.ExpiresAt,
	)
	return err
}

// UpdateStatus 更新隧道状态
func (q *TunnelQueries) UpdateStatus(ctx context.Context, id, status, errorMsg string) error {
	query := `UPDATE tunnels SET status = $2, error_message = $3, updated_at = NOW() WHERE id = $1`
	_, err := q.DB.Pool.Exec(ctx, query, id, status, errorMsg)
	return err
}

// Delete 删除隧道
func (q *TunnelQueries) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tunnels WHERE id = $1`
	_, err := q.DB.Pool.Exec(ctx, query, id)
	return err
}

// CountByUser 统计用户隧道数
func (q *TunnelQueries) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := q.DB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM tunnels WHERE user_id = $1", userID).Scan(&count)
	return count, err
}

// AllocatePort 分配端口
func (q *TunnelQueries) AllocatePort(ctx context.Context, port int, tunnelID, userID string) error {
	query := `
		INSERT INTO port_allocations (port, allocated_to, allocated_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (port) DO UPDATE SET allocated_to = $2, allocated_by = $3`
	_, err := q.DB.Pool.Exec(ctx, query, port, tunnelID, userID)
	return err
}

// ReleasePort 释放端口
func (q *TunnelQueries) ReleasePort(ctx context.Context, port int) error {
	query := `DELETE FROM port_allocations WHERE port = $1`
	_, err := q.DB.Pool.Exec(ctx, query, port)
	return err
}

// FindAvailablePort 查找可用端口
func (q *TunnelQueries) FindAvailablePort(ctx context.Context, startPort, endPort int) (int, error) {
	query := `
		SELECT port FROM port_allocations 
		WHERE port >= $1 AND port <= $2 
		ORDER BY port`
	
	rows, err := q.DB.Pool.Query(ctx, query, startPort, endPort)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	usedPorts := make(map[int]bool)
	for rows.Next() {
		var port int
		if err := rows.Scan(&port); err != nil {
			return 0, err
		}
		usedPorts[port] = true
	}

	for port := startPort; port <= endPort; port++ {
		if !usedPorts[port] {
			return port, nil
		}
	}

	return 0, fmt.Errorf("没有可用端口")
}
