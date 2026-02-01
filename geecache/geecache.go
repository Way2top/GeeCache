package geecache

import (
	"fmt"
	"log"
	"sync"
)

// Getter 满足 Getter 接口需要实现一个 Get 方法，可以通过一个 key 获取数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 是一种“长这样签名的函数”的类型： 接收 string，返回 ([]byte, error)
type GetterFunc func(key string) ([]byte, error)

// Get 给 GetterFunc 这个类型实现 Get 方法，这样 GetterFunc 就实现了 Getter 接口
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 是一个缓存命名空间
type Group struct {
	name      string // 每个 Group 都拥有一个唯一的名称
	getter    Getter // 缓存未命中时候获取数据源的回调函数
	mainCache cache  // cache.go 中实现的并发缓存
}

var (
	mu     sync.RWMutex // 用 RWMutex 是因为 groups 是读多写少的全局表，读锁可并发，提高性能
	groups = make(map[string]*Group)
)

// NewGroup 初始化 Group，要求必须传入 getter，否则初始化会失败
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	// 这里强制 getter 非空，如果为空则无法创建 Group
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup 根据 name 找到对应的 Group
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	// 先排除 key 为 空的情况，以免污染缓存空间
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// 从 mainCache 中查找缓存，如果命中缓存，直接返回
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	// 如果没命中，通过 load 拿数据
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 暂时直接从本地拿数据，等到后面添加上从其他数据源拿数据
	// 期望的流程是：本地缓存 --> 远程缓存 --> 本地数据源
	// 远程缓存后续加上，暂时直接从本地数据源拿
	return g.getLocally(key)
}

// getLocally 调用用户回调函数 g.getter.Get() 获取数据，并将元数据添加到缓存 mainCache 中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
