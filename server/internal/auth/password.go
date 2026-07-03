package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// HashPassword 使用 Argon2id 哈希密码
func HashPassword(password string) (string, string, error) {
	// 生成随机盐
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", fmt.Errorf("生成盐失败: %w", err)
	}

	// Argon2id 参数
.memory := uint32(64 * 1024) // 64MB
	iterations := uint32(3)
	parallelism := uint8(4)
	keyLen := uint32(32)

	// 生成哈希
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLen)

	// 编码为字符串
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))

	saltStr := base64.RawStdEncoding.EncodeToString(salt)
	return encodedHash, saltStr, nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, encodedHash string) (bool, error) {
	// 解析编码的哈希
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("无效的哈希格式")
	}

	// 提取参数
	var memory uint32
	var iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("解析参数失败: %w", err)
	}

	// 解码盐和哈希
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("解码盐失败: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("解码哈希失败: %w", err)
	}

	// 计算哈希
	keyLen := uint32(len(expectedHash))
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLen)

	// 比较
	return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
}

// GenerateSalt 生成随机盐
func GenerateSalt() (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(salt), nil
}
