package geecache

// ByteView 仅包含一个 []byte
type ByteView struct {
	b []byte
}

// Len 返回 ByteView 的长度，同时也需要实现这个方法从而实现 Value 接口
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一个拷贝后的数据，防止外部修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 返回缓存数据的只读字符串表示
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
