package geecache

import (
	"geecache/geecache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex // 锁
	lru        *lru.Cache // 将 lru.go 的内容封装
	cacheBytes int64
}

// add 实例化 lru，封装 Add 方法，加锁
func (c *cache) add(key string, value ByteView) {
	// 上锁
	c.mu.Lock()
	defer c.mu.Unlock() // 退出时释放锁
	// 延迟初始化，可以提供性能（如果这个缓存组根本没被用到，就不浪费内存）
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// get 实例化 lru，封装 Get 方法，加锁
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 如果缓存从来没加过值，那么查询就直接返回“未命中”
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok // 这里加括号的原因：v 的静态类型是接口 Value，实际存的是 ByteView，需要类型断言取出具体类型
	}
	return
}
