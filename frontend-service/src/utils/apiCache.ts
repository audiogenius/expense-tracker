// Simple in-memory cache for API requests
interface CacheItem {
  data: any
  timestamp: number
  ttl: number
}

class APICache {
  private cache = new Map<string, CacheItem>()
  private maxSize = 100 // Maximum number of cached items

  // Generate cache key from URL and params
  private getCacheKey(url: string, params?: Record<string, any>): string {
    const sortedParams = params ? Object.keys(params)
      .sort()
      .map(key => `${key}=${params[key]}`)
      .join('&') : ''
    
    return `${url}${sortedParams ? `?${sortedParams}` : ''}`
  }

  // Get cached data
  get(url: string, params?: Record<string, any>): any | null {
    const key = this.getCacheKey(url, params)
    const item = this.cache.get(key)
    
    if (!item) {
      return null
    }
    
    // Check if expired
    if (Date.now() - item.timestamp > item.ttl) {
      this.cache.delete(key)
      return null
    }
    
    return item.data
  }

  // Set cached data
  set(url: string, data: any, ttl: number = 2 * 60 * 1000, params?: Record<string, any>): void {
    const key = this.getCacheKey(url, params)
    
    // Remove oldest items if cache is full
    if (this.cache.size >= this.maxSize) {
      const oldestKey = this.cache.keys().next().value
      this.cache.delete(oldestKey)
    }
    
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl
    })
  }

  // Clear cache
  clear(): void {
    this.cache.clear()
  }

  // Clear expired items
  cleanup(): void {
    const now = Date.now()
    for (const [key, item] of this.cache.entries()) {
      if (now - item.timestamp > item.ttl) {
        this.cache.delete(key)
      }
    }
  }

  // Get cache size
  size(): number {
    return this.cache.size
  }

  // Clear cache for specific pattern
  clearPattern(pattern: string): void {
    for (const key of this.cache.keys()) {
      if (key.includes(pattern)) {
        this.cache.delete(key)
      }
    }
  }
}

// Global cache instance
export const apiCache = new APICache()

// Cleanup expired items every 5 minutes
setInterval(() => {
  apiCache.cleanup()
}, 5 * 60 * 1000)

// Cache TTL constants
export const CACHE_TTL = {
  TRANSACTIONS: 2 * 60 * 1000, // 2 minutes
  CATEGORIES: 10 * 60 * 1000,  // 10 minutes
  BALANCE: 1 * 60 * 1000,      // 1 minute
  SUGGESTIONS: 5 * 60 * 1000, // 5 minutes
} as const
