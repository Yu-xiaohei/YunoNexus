package auth

import (
	"sync"
	"time"
)

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	Count     int
	LockedAt  *time.Time
	FirstFail time.Time
}

// LoginLimiter 登录限速器
type LoginLimiter struct {
	mu       sync.RWMutex
	attempts map[string]*LoginAttempt
	maxTry   int
	lockTime time.Duration
}

// NewLoginLimiter 创建登录限速器
func NewLoginLimiter(maxTry int, lockTime time.Duration) *LoginLimiter {
	limiter := &LoginLimiter{
		attempts: make(map[string]*LoginAttempt),
		maxTry:   maxTry,
		lockTime: lockTime,
	}

	// 定期清理过期记录
	go limiter.cleanup()

	return limiter
}

// IsLocked 检查是否被锁定
func (l *LoginLimiter) IsLocked(identifier string) (bool, time.Duration) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempt, exists := l.attempts[identifier]
	if !exists {
		return false, 0
	}

	if attempt.LockedAt != nil {
		remain := l.lockTime - time.Since(*attempt.LockedAt)
		if remain > 0 {
			return true, remain
		}
		// 锁定已过期
		return false, 0
	}

	return false, 0
}

// RecordFailure 记录失败
func (l *LoginLimiter) RecordFailure(identifier string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	attempt, exists := l.attempts[identifier]
	if !exists {
		attempt = &LoginAttempt{
			FirstFail: time.Now(),
		}
		l.attempts[identifier] = attempt
	}

	attempt.Count++

	if attempt.Count >= l.maxTry {
		now := time.Now()
		attempt.LockedAt = &now
	}
}

// RecordSuccess 记录成功，清除失败记录
func (l *LoginLimiter) RecordSuccess(identifier string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.attempts, identifier)
}

// GetFailCount 获取失败次数
func (l *LoginLimiter) GetFailCount(identifier string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempt, exists := l.attempts[identifier]
	if !exists {
		return 0
	}
	return attempt.Count
}

// cleanup 定期清理过期记录
func (l *LoginLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for id, attempt := range l.attempts {
			// 锁定已过期且超过锁定期两倍时间，删除记录
			if attempt.LockedAt != nil && now.Sub(*attempt.LockedAt) > l.lockTime*2 {
				delete(l.attempts, id)
			}
			// 无锁定且超过1小时，删除记录
			if attempt.LockedAt == nil && now.Sub(attempt.FirstFail) > time.Hour {
				delete(l.attempts, id)
			}
		}
		l.mu.Unlock()
	}
}
