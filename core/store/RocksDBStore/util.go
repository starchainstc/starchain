package RocksDBStore

import (
	"bytes"
	"encoding/binary"
	"unsafe"
	"math"
	"sync"
	"sync/atomic"
	"fmt"
)

var (
	SEP = []byte{','}
	KEY = []byte{'+'} // Key Prefix
	SOK = []byte{'['} // Start of Key
	EOK = []byte{']'} // End of Key
)

type ElementType byte

const (
	STRING    ElementType = 's'
	HASH                  = 'h'
	LIST                  = 'l'
	SORTEDSET             = 'z'
	NONE                  = '0'
)

func (e ElementType) String() string {
	switch byte(e) {
	case 's':
		return "string"
	case 'h':
		return "hash"
	case 'l':
		return "list"
	case 'z':
		return "sortedset"
	case 'e':
		return "set" // not design
	default:
		return "none"
	}
}

type IterDirection int

const (
	IterForward IterDirection = iota
	IterBackward
)

// 字节范围
const (
	MINBYTE byte = 0
	MAXBYTE byte = math.MaxUint8
)

func rawKey(key []byte, t ElementType) []byte {
	return bytes.Join([][]byte{KEY, key, SEP, []byte{byte(t)}}, nil)
}

// 范围判断 min <= v <= max
func between(v, min, max []byte) bool {
	return bytes.Compare(v, min) >= 0 && bytes.Compare(v, max) <= 0
}

// 复制数组
func copyBytes(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

// 使用二进制存储整形
func Int64ToBytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func SplitKeyName(key []byte) (string, string) {
	k := string(key)
	length := len(key)
	okString := string(k[1 : length-2])
	ttype := string(k[length-1 : length])
	return okString, ttype
}

func Str2bytes(s string) []byte {
	ptr := (*[2]uintptr)(unsafe.Pointer(&s))
	btr := [3]uintptr{ptr[0], ptr[1], ptr[1]}
	return *(*[]byte)(unsafe.Pointer(&btr))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}


func IncrSignal(val int64) int64 {
	return atomic.AddInt64((*int64)(&val), 1)
}

// Atomic Counter, simple enough, easy to Incr/Decr
// c := Counter(0)
// c.SetCount(100)
// c.Incr(1) or c.Decr(1)
// fmt.Println(c) or c.Count()
type Counter int64

func (c *Counter) SetCount(val int64) {
	atomic.StoreInt64((*int64)(c), val)
}

func (c *Counter) Count() int64 {
	return atomic.LoadInt64((*int64)(c))
}

func (c *Counter) Incr(delta int64) int64 {
	return atomic.AddInt64((*int64)(c), delta)
}

func (c *Counter) Decr(delta int64) int64 {
	return atomic.AddInt64((*int64)(c), delta*-1)
}

func (c *Counter) String() string {
	return fmt.Sprint(c.Count())
}

// Counter Collection
// factory := NewFactory()
// factory.Get("set").Incr(1)
// factory.Get("get").Incr(1)
// factory.Get("del").Incr(1)
// factory.Get("total").Incr(3)
type Counters struct {
	table map[string]*Counter
	mu    sync.Mutex
}

func NewCounters() *Counters {
	return &Counters{
		table: make(map[string]*Counter),
	}
}

// Get or auto create a Counter by name
func (f *Counters) C(name string) *Counter {
	var c *Counter
	var ok bool
	if c, ok = f.table[name]; !ok {
		f.mu.Lock()
		if c, ok = f.table[name]; !ok {
			tmp := Counter(0)
			c = &tmp
			f.table[name] = c
		}
		f.mu.Unlock()
	}
	return c
}
