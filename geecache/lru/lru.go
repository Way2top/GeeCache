package lru

import "container/list"

type Cache struct {
	maxBytes  int64      // 允许使用的最大内存
	nbytes    int64      // 当前已使用的内存
	ll        *list.List // Go 标准库实现的双向链表, 链表里存的不是 key 或 value，而是后面的 entry
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数
}

// entry 是链表中存储的内容
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

// New 用于初始化 Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 是当缓存需要被使用的时候，根据 key 查找对应缓存，并将其存储节点移动到队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 用于移除最近最少访问的节点（队首）
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 取到队首节点
	if ele != nil {
		c.ll.Remove(ele) // 删除队首节点
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key) // 删除 c.cache 中的映射关系（这里就体现了前面 entry 需要存储 key 的作用了）
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 如果 onEvicted 不为空，就调用该函数，将删除的值传出去
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 用于修改或者更新缓存
func (c *Cache) Add(key string, value Value) {
	// 如果传入的 key 存在于 c.cache 中，就将节点移动到队尾，然后修改元素的值
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else { // 如果 key 不存在于 c.cache 中，就新建一个节点，然后移动到队尾
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 每次新增或者修改元素之后都检查一下当前使用的缓存是否超过了最大缓存，如果超过了，移除队首元素
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len 用于计算 c.ll 的长度，也即链表的长度，所以返回的结果就是缓存的个数
func (c *Cache) Len() int {
	return c.ll.Len() // 这里的 c.ll.Len() 是 GO 标准库内置的方法
}
