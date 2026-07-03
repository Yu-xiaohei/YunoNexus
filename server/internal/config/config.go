package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	// 服务器配置
	Server ServerConfig

	// 数据库配置
	Database DatabaseConfig

	// Redis配置
	Redis RedisConfig

	// JWT配置
	JWT JWTConfig

	// 代理配置
	Proxy ProxyConfig

	// 安全配置
	Security SecurityConfig

	// 日志配置
	Log LogConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string
	Port int
	Mode string // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// RedisConfig Redis配置
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Port             int
	TunnelPortRange  [2]int // [start, end]
	HeartbeatInterval int
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	LoginMaxAttempts   int
	LoginLockoutDuration time.Duration
	SessionTimeout     time.Duration
	AllowedPortsRange  [2]int
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string
	Format string
}

// Load 加载配置
func Load() *Config {
	// 尝试加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvInt("SERVER_PORT", 8080),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			Name:     getEnv("DB_NAME", "yunonexus"),
			User:     getEnv("DB_USER", "yunonexus"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", ""),
			Expiry:        getEnvDuration("JWT_EXPIRY", 2*time.Hour),
			RefreshExpiry: getEnvDuration("REFRESH_EXPIRY", 7*24*time.Hour),
		},
		Proxy: ProxyConfig{
			Port:              getEnvInt("PROXY_PORT", 9443),
			TunnelPortRange:   [2]int{
				getEnvInt("TUNNEL_PORT_START", 30000),
				getEnvInt("TUNNEL_PORT_END", 30100),
			},
			HeartbeatInterval: getEnvInt("HEARTBEAT_INTERVAL", 30),
		},
		Security: SecurityConfig{
			LoginMaxAttempts:    getEnvInt("LOGIN_MAX_ATTEMPTS", 5),
			LoginLockoutDuration: getEnvDuration("LOGIN_LOCKOUT_DURATION", 15*time.Minute),
			SessionTimeout:      getEnvDuration("SESSION_TIMEOUT", 2*time.Hour),
			AllowedPortsRange:   [2]int{
				getEnvInt("ALLOWED_PORT_START", 30000),
				getEnvInt("ALLOWED_PORT_END", 30100),
			},
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
