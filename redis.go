package nono

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

type NonoRedis struct {
	url     string
	pwd     string
	session []*redis.Client
	lock    []*sync.Mutex
	db      int
}

//GetCache getCache,if string=="",cache need reset
func (t *NonoRedis) GetCache(key string) (string, error, bool) {
	r := t.GetRedisByDb(7)
	result := r.Get(key)
	return result.String(), result.Err(), false
}
func (t *NonoRedis) SetCache(key string, value func() string) error {
	return nil
}

//var cache *redis.Client
func (t *NonoRedis) Limit(key string, times int64, extime time.Duration) bool {
	r := t.GetRedisByDb(11)

	if cmd := r.Incr("limit." + key); cmd.Val() > times {
		return false
	} else {
		if cmd.Val() == 1 {
			r.Expire("limit."+key, extime)
		}
		return true
	}
}
func (t *NonoRedis) Unmarshal(s []string, resuslt interface{}) error {
	ss := "[" + strings.Join(s, ",") + "]"
	err := json.Unmarshal([]byte(ss), &resuslt)
	return err
}
func (t *NonoRedis) GetID(key string) int64 {
	r := t.GetRedisByDb(10)
	cmd := r.Incr("limit." + key)
	if cmd.Err() == nil {
		return cmd.Val()
	}
	return -1
}
func (t *NonoRedis) Unlimit(key string) bool {
	r := t.GetRedisByDb(11)
	if cmd := r.Del("limit." + key); cmd.Err() == nil {
		return true
	}
	return true
}
func (t *NonoRedis) LockWithExprie(s string, extime time.Duration) bool {
	r := t.GetRedisByDb(12)
	for {
		if cmd := r.SetNX("lock."+s, true, extime); cmd.Val() == true {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (t *NonoRedis) Lock(s string) bool {
	return t.LockWithExprie(s, 10*time.Second)
}
func (t *NonoRedis) IsLocked(s string) bool {
	r := t.GetRedisByDb(12)
	cmd := r.Get("lock." + s)
	if len(cmd.Val()) > 0 {
		return true
	}
	return false
}
func (t *NonoRedis) LockNoWait(s string) bool {
	r := t.GetRedisByDb(12)
	if cmd := r.SetNX("lock."+s, true, 10*time.Second); cmd.Val() == true {
		return true
	}
	return false
}
func (t *NonoRedis) Unlock(s string) bool {
	r := t.GetRedisByDb(12)
	if cmd := r.Del("lock." + s); cmd.Err() == nil {
		return true
	}
	return true
}
func NewRedis(url string, pwd string, db int) *NonoRedis {
	t := &NonoRedis{
		url: url,
		pwd: pwd,
		db:  db,
	}
	t.session = make([]*redis.Client, 32)
	t.lock = make([]*sync.Mutex, 32)
	t.session[db] = t.newRedisClient(db)
	for i := 0; i < len(t.lock); i++ {
		var lock sync.Mutex
		t.lock[i] = &lock
	}
	return t
}
func (t *NonoRedis) GetRedis() *redis.Client {
	return t.GetRedisByDb(t.db)
}
func (t *NonoRedis) GetRedisByDb(i int) *redis.Client {
	t.lock[i].Lock()
	defer t.lock[i].Unlock()
	if t.session[i] != nil && t.session[i].Ping().Err() == nil {
		return t.session[i]
	} else {
		s := t.newRedisClient(i)
		if t.session[i] != nil {
			temp := *t.session[i]
			go func() {
				time.Sleep(30 * time.Second)
				temp.Close()
			}()
		}
		t.session[i] = s
		return t.session[i]
	}
}
func (t *NonoRedis) newRedisClient(db int) *redis.Client {
	//dd这里忽略了redis可能存在的故障的情况
	for {
		client := redis.NewClient(&redis.Options{
			Addr:     t.url,
			DB:       db,
			PoolSize: 10,
			Password: t.pwd,
		})
		pong := client.Ping()
		if pong.Err() == nil {
			fmt.Println("Connect Redis:", t.url, db, pong.Val())
			return client
		} else {
			fmt.Println("Connect Redis:", t.url, t.pwd, pong.Err())
			time.Sleep(1 * time.Second)
		}
	}
}
