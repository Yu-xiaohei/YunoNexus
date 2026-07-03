-- YunoNexus 数据库初始化
-- 版本: 001_init
-- 日期: 2026-07-03

-- 启用UUID扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 启用pgcrypto扩展（用于加密函数）
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- 用户表
-- ============================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(64) NOT NULL UNIQUE,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    salt            VARCHAR(64) NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'user',
    max_tunnels     INT NOT NULL DEFAULT 3,
    max_bandwidth   BIGINT NOT NULL DEFAULT 10485760,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    expires_at      TIMESTAMPTZ,
    two_factor_key  VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- ============================================
-- 设备表
-- ============================================
CREATE TABLE devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name     VARCHAR(128) NOT NULL,
    device_type     VARCHAR(20) NOT NULL,
    fingerprint     VARCHAR(64) NOT NULL,
    os_version      VARCHAR(64),
    app_version     VARCHAR(20),
    public_key      TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    ip_whitelist    INET[] DEFAULT '{}',
    last_seen_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, fingerprint)
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_fingerprint ON devices(fingerprint);

-- ============================================
-- 会话表
-- ============================================
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id       UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token           VARCHAR(64) NOT NULL UNIQUE,
    refresh_token   VARCHAR(64) NOT NULL UNIQUE,
    ip_address      INET NOT NULL,
    user_agent      TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_device_id ON sessions(device_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- ============================================
-- 隧道表
-- ============================================
CREATE TABLE tunnels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id       UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name            VARCHAR(128) NOT NULL,
    protocol        VARCHAR(10) NOT NULL,
    local_host      VARCHAR(255) NOT NULL,
    local_port      INT NOT NULL,
    remote_host     VARCHAR(255),
    remote_port     INT NOT NULL,
    domain          VARCHAR(255),
    status          VARCHAR(20) NOT NULL DEFAULT 'inactive',
    error_message   TEXT,
    config          JSONB DEFAULT '{}',
    bandwidth_limit BIGINT,
    traffic_used    BIGINT DEFAULT 0,
    traffic_limit   BIGINT,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, remote_port),
    UNIQUE(user_id, domain),
    CONSTRAINT chk_protocol CHECK (protocol IN ('tcp','udp','http','https','websocket')),
    CONSTRAINT chk_ports CHECK (remote_port > 0 AND remote_port < 65536)
);

CREATE INDEX idx_tunnels_user_id ON tunnels(user_id);
CREATE INDEX idx_tunnels_device_id ON tunnels(device_id);
CREATE INDEX idx_tunnels_status ON tunnels(status);

-- ============================================
-- 端口绑定表
-- ============================================
CREATE TABLE port_bindings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tunnel_id       UUID NOT NULL REFERENCES tunnels(id) ON DELETE CASCADE,
    port            INT NOT NULL UNIQUE,
    protocol        VARCHAR(10) NOT NULL,
    description     VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_port_bindings_port ON port_bindings(port);

-- ============================================
-- 流量日志表（按月分区）
-- ============================================
CREATE TABLE traffic_logs (
    id              BIGSERIAL,
    tunnel_id       UUID NOT NULL REFERENCES tunnels(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id),
    bytes_sent      BIGINT NOT NULL DEFAULT 0,
    bytes_recv      BIGINT NOT NULL DEFAULT 0,
    connections     INT NOT NULL DEFAULT 0,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, recorded_at)
) PARTITION BY RANGE (recorded_at);

-- 创建当前月份的分区
CREATE TABLE traffic_logs_2026_07 PARTITION OF traffic_logs
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE TABLE traffic_logs_2026_08 PARTITION OF traffic_logs
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

CREATE TABLE traffic_logs_2026_09 PARTITION OF traffic_logs
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE TABLE traffic_logs_2026_10 PARTITION OF traffic_logs
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');

CREATE TABLE traffic_logs_2026_11 PARTITION OF traffic_logs
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');

CREATE TABLE traffic_logs_2026_12 PARTITION OF traffic_logs
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

CREATE INDEX idx_traffic_logs_tunnel_time ON traffic_logs(tunnel_id, recorded_at);
CREATE INDEX idx_traffic_logs_user_time ON traffic_logs(user_id, recorded_at);

-- ============================================
-- 端口分配表
-- ============================================
CREATE TABLE port_allocations (
    port            INT PRIMARY KEY,
    allocated_to    UUID REFERENCES tunnels(id) ON DELETE SET NULL,
    allocated_by    UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================
-- 审计日志表
-- ============================================
CREATE TABLE audit_logs (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    action          VARCHAR(64) NOT NULL,
    resource_type   VARCHAR(32),
    resource_id     UUID,
    ip_address      INET,
    user_agent      TEXT,
    details         JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- ============================================
-- 系统配置表
-- ============================================
CREATE TABLE system_configs (
    key             VARCHAR(64) PRIMARY KEY,
    value           JSONB NOT NULL,
    description     TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by      UUID REFERENCES users(id)
);

-- 插入默认配置
INSERT INTO system_configs (key, value, description) VALUES
    ('default_max_tunnels', '3', '默认最大隧道数'),
    ('default_max_bandwidth', '10485760', '默认最大带宽(bytes/sec)'),
    ('allowed_ports_range', '{"start": 30000, "end": 30100}', '允许的端口范围'),
    ('session_timeout', '7200', '会话超时时间(秒)'),
    ('maintenance_mode', 'false', '维护模式');

-- ============================================
-- 加密密钥表
-- ============================================
CREATE TABLE encryption_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_type        VARCHAR(32) NOT NULL,
    key_hash        VARCHAR(64) NOT NULL,
    public_key      TEXT,
    encrypted_key   TEXT,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotated_at      TIMESTAMPTZ
);

CREATE INDEX idx_encryption_keys_type ON encryption_keys(key_type);

-- ============================================
-- 创建管理员用户（默认密码: admin123）
-- 注意：生产环境请立即修改密码
-- ============================================
INSERT INTO users (username, email, password_hash, salt, role) VALUES
    ('admin', 'admin@yunonexus.local', 
     crypt('admin123', gen_salt('bf', 10)),
     'default_salt',
     'admin');
