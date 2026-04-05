package cache

import (
	"sync"
	"time"
)

// TokenCache Token统计缓存
type TokenCache struct {
	mu   sync.RWMutex
	data map[string]*CacheEntry
	ttl  time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Data      interface{}
	Timestamp time.Time
}

// NewTokenCache 创建新的Token缓存
func NewTokenCache(ttl time.Duration) *TokenCache {
	cache := &TokenCache{
		data: make(map[string]*CacheEntry),
		ttl:  ttl,
	}
	// 启动清理协程
	go cache.cleanup()
	return cache
}

// Get 获取缓存值
func (c *TokenCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Since(entry.Timestamp) > c.ttl {
		return nil, false
	}

	return entry.Data, true
}

// Set 设置缓存值
func (c *TokenCache) Set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
	}
}

// Delete 删除缓存值
func (c *TokenCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// Clear 清空所有缓存
func (c *TokenCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheEntry)
}

// cleanup 定期清理过期缓存
func (c *TokenCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.data {
			if time.Since(entry.Timestamp) > c.ttl {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// 全局缓存实例
var (
	// TokenStatsCache 会话统计缓存（5分钟TTL）
	TokenStatsCache = NewTokenCache(5 * time.Minute)
	// ProjectTokenStatsCache 项目统计缓存（10分钟TTL）
	ProjectTokenStatsCache = NewTokenCache(10 * time.Minute)
	// GlobalTokenStatsCache 全局统计缓存（15分钟TTL）
	GlobalTokenStatsCache = NewTokenCache(15 * time.Minute)
)
