// pkg/cache/cache.go
package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Common error
var ErrKeyNotFound = errors.New("key not found")

// Cache interface for multiple implementations
type Cache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
	Has(key string) bool
}

// ======================
// MemoryCache
// ======================

type MemoryCache struct {
	items map[string]memoryItem
	mu    sync.RWMutex
}

type memoryItem struct {
	value      interface{}
	expiration time.Time
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]memoryItem),
	}
	// Start cleanup goroutine
	go cache.cleanup()
	return cache
}

func (c *MemoryCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, ErrKeyNotFound
	}
	if time.Now().After(item.expiration) {
		return nil, ErrKeyNotFound
	}
	return item.value, nil
}

func (c *MemoryCache) Set(key string, value interface{}, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = memoryItem{
		value:      value,
		expiration: time.Now().Add(expiration),
	}
	return nil
}

func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; !exists {
		return ErrKeyNotFound
	}
	delete(c.items, key)
	return nil
}

func (c *MemoryCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}
	return time.Now().Before(item.expiration)
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// ======================
// RedisCache (placeholder)
// ======================

type RedisCache struct {
	// Redis connection would be stored here
}

func NewRedisCache(addr string, password string, db int) (*RedisCache, error) {
	// Connect to Redis (not implemented)
	return &RedisCache{}, nil
}

func (c *RedisCache) Get(key string) (interface{}, error) {
	return nil, errors.New("redis cache not implemented")
}
func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	return errors.New("redis cache not implemented")
}
func (c *RedisCache) Delete(key string) error {
	return errors.New("redis cache not implemented")
}
func (c *RedisCache) Has(key string) bool {
	return false
}

// ======================
// FileCache (minimal working version)
// ======================

type FileCache struct {
	path string
	mu   sync.RWMutex
}

func NewFileCache(path string) (*FileCache, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %v", err)
	}
	return &FileCache{path: path}, nil
}

func (c *FileCache) filename(key string) string {
	return filepath.Join(c.path, fmt.Sprintf("%s.cache", key))
}

func (c *FileCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	file := c.filename(key)
	data, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	var item struct {
		Value      interface{} `json:"value"`
		Expiration int64       `json:"expiration"`
	}
	if err := json.Unmarshal(data, &item); err != nil {
		os.Remove(file)
		return nil, err
	}

	if time.Now().Unix() > item.Expiration {
		os.Remove(file)
		return nil, ErrKeyNotFound
	}
	return item.Value, nil
}

func (c *FileCache) Set(key string, value interface{}, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := struct {
		Value      interface{} `json:"value"`
		Expiration int64       `json:"expiration"`
	}{
		Value:      value,
		Expiration: time.Now().Add(expiration).Unix(),
	}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return os.WriteFile(c.filename(key), data, 0644)
}

func (c *FileCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file := c.filename(key)
	if err := os.Remove(file); err != nil {
		if os.IsNotExist(err) {
			return ErrKeyNotFound
		}
		return err
	}
	return nil
}

func (c *FileCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	file := c.filename(key)
	data, err := os.ReadFile(file)
	if err != nil {
		return false
	}

	var item struct {
		Expiration int64 `json:"expiration"`
	}
	if err := json.Unmarshal(data, &item); err != nil {
		os.Remove(file)
		return false
	}

	if time.Now().Unix() > item.Expiration {
		os.Remove(file)
		return false
	}
	return true
}
