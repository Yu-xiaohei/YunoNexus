package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"math"
	"strings"
	"time"
)

// TOTP 两步验证
type TOTP struct {
	Secret    string
	Issuer    string
	Account   string
	Digits    int
	Period    int
	Algorithm string
}

// NewTOTP 创建新的TOTP实例
func NewTOTP(secret, issuer, account string) *TOTP {
	return &TOTP{
		Secret:    secret,
		Issuer:    issuer,
		Account:   account,
		Digits:    6,
		Period:    30,
		Algorithm: "SHA1",
	}
}

// GenerateSecret 生成随机密钥
func GenerateSecret() string {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		// 如果随机数生成失败，使用时间戳+随机数作为后备
		for i := range secret {
			secret[i] = byte(time.Now().UnixNano() >> uint(i*3) & 0xff)
		}
	}
	return base32.StdEncoding.EncodeToString(secret)
}

// GenerateCode 生成TOTP代码
func (t *TOTP) GenerateCode() (string, error) {
	// 计算时间步
	timeStep := time.Now().Unix() / int64(t.Period)

	// 将时间步转换为字节
	timeBytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		timeBytes[i] = byte(timeStep & 0xff)
		timeStep >>= 8
	}

	// 解码密钥
	secretBytes, err := base32.StdEncoding.DecodeString(strings.ToUpper(t.Secret))
	if err != nil {
		return "", fmt.Errorf("解码密钥失败: %w", err)
	}

	// 计算HMAC-SHA1
	mac := hmac.New(sha1.New, secretBytes)
	mac.Write(timeBytes)
	hash := mac.Sum(nil)

	// 动态截断
	offset := hash[len(hash)-1] & 0x0f
	code := int32((int32(hash[offset])&0x7f)<<24 |
		(int32(hash[offset+1])&0xff)<<16 |
		(int32(hash[offset+2])&0xff)<<8 |
		(int32(hash[offset+3])&0xff))

	// 生成指定长度的代码
	code = code % int32(math.Pow10(t.Digits))
	return fmt.Sprintf("%0*d", t.Digits, code), nil
}

// VerifyCode 验证TOTP代码
func (t *TOTP) VerifyCode(code string, window int) bool {
	// 检查当前时间步和前后window个时间步
	for i := -window; i <= window; i++ {
		timeStep := time.Now().Unix()/int64(t.Period) + int64(i)

		// 将时间步转换为字节
		timeBytes := make([]byte, 8)
		tempTime := timeStep
		for j := 7; j >= 0; j-- {
			timeBytes[j] = byte(tempTime & 0xff)
			tempTime >>= 8
		}

		// 解码密钥
		secretBytes, err := base32.StdEncoding.DecodeString(strings.ToUpper(t.Secret))
		if err != nil {
			continue
		}

		// 计算HMAC-SHA1
		mac := hmac.New(sha1.New, secretBytes)
		mac.Write(timeBytes)
		hash := mac.Sum(nil)

		// 动态截断
		offset := hash[len(hash)-1] & 0x0f
		expectedCode := int32((int32(hash[offset])&0x7f)<<24 |
			(int32(hash[offset+1])&0xff)<<16 |
			(int32(hash[offset+2])&0xff)<<8 |
			(int32(hash[offset+3])&0xff))

		expectedCode = expectedCode % int32(math.Pow10(t.Digits))
		expectedStr := fmt.Sprintf("%0*d", t.Digits, expectedCode)

		if code == expectedStr {
			return true
		}
	}

	return false
}

// GetProvisioningURI 获取配置URI（用于生成二维码）
func (t *TOTP) GetProvisioningURI() string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		t.Issuer, t.Account, t.Secret, t.Issuer, t.Digits, t.Period)
}
