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

//Cache
func Cache(name string, seccnd int64, f func() interface{}) interface{} {
	c := cache[name]
	if c == nil {
		c = &cacheStruct{}
	}
	if time.Now().Unix()-c.time > seccnd {
		if i := atomic.AddInt32(&c.lock, 1); i == 1 {
			c.date = f()
			atomic.StoreInt32(&c.lock, 0)
			return c.date
		}
	}
	return c.date
}
