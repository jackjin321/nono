package nono

import (
	"sync"
	"sync/atomic"
	"time"
)

type cacheStruct struct {
	id   int64
	time int64
	lock int32
	date interface{}
}

var cache sync.Map
var onceClearCache sync.Once

func clearCache() {
	const cacheing = 120
	for {
		time.Sleep(cacheing * time.Second)
		now := time.Now().Unix()
		cache.Range(func(key, value interface{}) bool {
			c, ok := value.(*cacheStruct)
			if !ok || now-c.time > cacheing {
				cache.Delete(key)
			}
			return false
		})
	}
}

//Cache 输入名字,过期时间(秒),一致的id,如果 id相同,那么即使超时了也不重新读取 需要运行的函数,如果超时,那么运行函数并把返回值缓存到名字里
func Cache(name string, seccnd int64, f func(id int64) (interface{}, int64)) interface{} {
	go onceClearCache.Do(clearCache)
RE:
	cc, _ := cache.LoadOrStore(name, &cacheStruct{})
	//c := cache[name]
	//if c == nil {
	//	c = &cacheStruct{}
	//	cache[name] = c
	//}
	c, ok := cc.(*cacheStruct)
	if !ok {
		cache.Delete(name)
		goto RE
	}
	if f == nil {
		return c.date
	}
	//如果id相等,直接返回

	if time.Now().Unix()-c.time > seccnd {
		if i := atomic.AddInt32(&c.lock, 1); i == 1 {
			result, idd := f(c.id)
			if idd != 0 && idd == c.id {
				c.time = time.Now().Unix()
			}
			if result != nil {
				c.time = time.Now().Unix()
				c.date = result
				c.id = idd
			}
			atomic.StoreInt32(&c.lock, 0)
			return c.date
		}
	}
	return c.date
}
