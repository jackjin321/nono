package nono

import (
	"sync/atomic"
	"time"
)

type cacheStruct struct {
	time int64
	lock int32
	date interface{}
}

var cache = make(map[string]*cacheStruct)

//CacheRead 仅仅读取返回值，如果
func CacheRead(name string) interface{} {
	return Cache(name, 0, nil)
}

//Cache 输入名字,过期时间(秒), 需要运行的函数,如果超时,那么运行函数并把返回值缓存到名字里
func Cache(name string, seccnd int64, f func() interface{}) interface{} {
	c := cache[name]
	if c == nil {
		c = &cacheStruct{}
		cache[name] = c
	}
	if f == nil {
		return c.date
	}
	if time.Now().Unix()-c.time > seccnd {
		if i := atomic.AddInt32(&c.lock, 1); i == 1 {
			c.time = time.Now().Unix()
			c.date = f()
			atomic.StoreInt32(&c.lock, 0)
			return c.date
		}
	}
	return c.date
}
