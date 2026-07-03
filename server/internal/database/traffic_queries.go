package database

import (
	"context"

	"github.com/yunonexus/server/internal/models"
)

// TrafficQueries 流量统计查询
type TrafficQueries struct {
	DB *DB
}

// NewTrafficQueries 创建流量统计查询实例
func NewTrafficQueries(db *DB) *TrafficQueries {
	return &TrafficQueries{DB: db}
}

// GetUserStats 获取用户流量统计
func (q *TrafficQueries) GetUserStats(ctx context.Context, userID string) (*models.TrafficStats, error) {
	query := `
		SELECT COALESCE(SUM(bytes_sent), 0) as total_sent,
		       COALESCE(SUM(bytes_recv), 0) as total_recv,
		       COALESCE(SUM(connections), 0) as total_connections
		FROM traffic_logs 
		WHERE user_id = $1`

	stats := &models.TrafficStats{}
	err := q.DB.Pool.QueryRow(ctx, query, userID).Scan(
		&stats.TotalSent, &stats.TotalRecv, &stats.TotalConnections,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTunnelStats 获取隧道流量统计
func (q *TrafficQueries) GetTunnelStats(ctx context.Context, tunnelID string) (*models.TrafficStats, error) {
	query := `
		SELECT COALESCE(SUM(bytes_sent), 0) as total_sent,
		       COALESCE(SUM(bytes_recv), 0) as total_recv,
		       COALESCE(SUM(connections), 0) as total_connections
		FROM traffic_logs 
		WHERE tunnel_id = $1`

	stats := &models.TrafficStats{}
	err := q.DB.Pool.QueryRow(ctx, query, tunnelID).Scan(
		&stats.TotalSent, &stats.TotalRecv, &stats.TotalConnections,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// RecordTraffic 记录流量
func (q *TrafficQueries) RecordTraffic(ctx context.Context, log *models.TrafficLog) error {
	query := `
		INSERT INTO traffic_logs (tunnel_id, user_id, bytes_sent, bytes_recv, connections)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := q.DB.Pool.Exec(ctx, query,
		log.TunnelID, log.UserID, log.BytesSent, log.BytesRecv, log.Connections,
	)

	return err
}
