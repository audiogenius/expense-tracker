package cache

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryCache represents in-memory cache with TTL
type MemoryCache struct {
	items map[string]cacheItem
	mutex sync.RWMutex
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoryCache creates new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves value from cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		return nil, false
	}

	return item.value, true
}

// Set stores value in cache with TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete removes value from cache
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear removes all items from cache
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]cacheItem)
}

// Size returns number of items in cache
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// cleanup removes expired items periodically
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiresAt) {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Cache keys for different data types
const (
	KeyUserBalance     = "user_balance_%d"
	KeyUserExpenses    = "user_expenses_%d_%s_%d" // user_id, period, page
	KeyUserCategories  = "user_categories_%d"
	KeyUserSuggestions = "user_suggestions_%d_%s" // user_id, query
	KeyCategories      = "categories_all"
	KeySubcategories   = "subcategories_%d" // category_id
)

// Helper functions for common cache operations
func (c *MemoryCache) GetUserBalance(userID int64) (interface{}, bool) {
	key := fmt.Sprintf(KeyUserBalance, userID)
	return c.Get(key)
}

func (c *MemoryCache) SetUserBalance(userID int64, balance interface{}) {
	key := fmt.Sprintf(KeyUserBalance, userID)
	c.Set(key, balance, 5*time.Minute)
}

func (c *MemoryCache) GetUserExpenses(userID int64, period string, page int) (interface{}, bool) {
	key := fmt.Sprintf(KeyUserExpenses, userID, period, page)
	return c.Get(key)
}

func (c *MemoryCache) SetUserExpenses(userID int64, period string, page int, expenses interface{}) {
	key := fmt.Sprintf(KeyUserExpenses, userID, period, page)
	c.Set(key, expenses, 5*time.Minute)
}

func (c *MemoryCache) GetUserSuggestions(userID int64, query string) (interface{}, bool) {
	key := fmt.Sprintf(KeyUserSuggestions, userID, query)
	return c.Get(key)
}

func (c *MemoryCache) SetUserSuggestions(userID int64, query string, suggestions interface{}) {
	key := fmt.Sprintf(KeyUserSuggestions, userID, query)
	c.Set(key, suggestions, 5*time.Minute)
}

func (c *MemoryCache) GetCategories() (interface{}, bool) {
	return c.Get(KeyCategories)
}

func (c *MemoryCache) SetCategories(categories interface{}) {
	c.Set(KeyCategories, categories, 10*time.Minute) // Categories change rarely
}

func (c *MemoryCache) GetSubcategories(categoryID int) (interface{}, bool) {
	key := fmt.Sprintf(KeySubcategories, categoryID)
	return c.Get(key)
}

func (c *MemoryCache) SetSubcategories(categoryID int, subcategories interface{}) {
	key := fmt.Sprintf(KeySubcategories, categoryID)
	c.Set(key, subcategories, 5*time.Minute)
}

// ClearPattern removes all items matching a pattern
func (c *MemoryCache) ClearPattern(pattern string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key := range c.items {
		if strings.Contains(key, pattern) {
			delete(c.items, key)
		}
	}
}