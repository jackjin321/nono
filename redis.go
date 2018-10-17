package nono

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

//Redis 封装的redis结构
type Redis struct {
	url     string
	pwd     string
	session []*redis.Client
	lock    []*sync.Mutex
	db      int
}

//Limit 输入(键值,次数,总时间)进行限制,如果没有限制返回true,被限制了返回false
func (t *Redis) Limit(key string, times int64, extime time.Duration) bool {
	r := t.GetRedisByDb(5)
	cmd := r.Incr("limit." + key)
	if cmd.Val() > times {
		return false
	}
	if cmd.Val() == 1 {
		r.Expire("limit."+key, extime)
	}
	return true
}

//Unlimit 解除某个键的限制
func (t *Redis) Unlimit(key string) bool {
	r := t.GetRedisByDb(5)
	if cmd := r.Del("limit." + key); cmd.Err() == nil {
		return true
	}
	return true
}

//Unmarshal 返回某些特定的函数,暂时忘记使用方式,别用
func (t *Redis) Unmarshal(s []string, resuslt interface{}) error {
	ss := "[" + strings.Join(s, ",") + "]"
	err := json.Unmarshal([]byte(ss), &resuslt)
	return err
}

//GetID 在db9里进行id的递增运算并返回一个唯一的id
func (t *Redis) GetID(key string) int64 {
	r := t.GetRedisByDb(9)
	cmd := r.Incr("limit." + key)
	if cmd.Err() == nil {
		return cmd.Val()
	}
	return -1
}

//LockWithExprie 带过期时间的全局锁
func (t *Redis) LockWithExprie(s string, extime time.Duration) bool {
	r := t.GetRedisByDb(4)
	for {
		if cmd := r.SetNX("lock."+s, true, extime); cmd.Val() == true {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
}

//Lock 全局锁10秒
func (t *Redis) Lock(s string) bool {
	return t.LockWithExprie(s, 10*time.Second)
}

//IsLocked 确定是够有全局锁
func (t *Redis) IsLocked(s string) bool {
	r := t.GetRedisByDb(4)
	cmd := r.Get("lock." + s)
	if len(cmd.Val()) > 0 {
		return true
	}
	return false
}

//LockNoWait 全局锁,但是不都塞直接返回
func (t *Redis) LockNoWait(s string) bool {
	r := t.GetRedisByDb(4)
	if cmd := r.SetNX("lock."+s, true, 10*time.Second); cmd.Val() == true {
		return true
	}
	return false
}

//Unlock 解除全局锁
func (t *Redis) Unlock(s string) bool {
	r := t.GetRedisByDb(4)
	if cmd := r.Del("lock." + s); cmd.Err() == nil {
		return true
	}
	return true
}

//NewRedis 输入密码地址和db号码,返回封装的redis
func NewRedis(url string, pwd string, db int) *Redis {
	t := &Redis{
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

//GetRedis 返回默认db的redis.client
func (t *Redis) GetRedis() *redis.Client {
	return t.GetRedisByDb(t.db)
}

//GetRedisByDb 返回指定db号码的redis.client
func (t *Redis) GetRedisByDb(i int) *redis.Client {
	t.lock[i].Lock()
	defer t.lock[i].Unlock()
	if t.session[i] != nil && t.session[i].Ping().Err() == nil {
		return t.session[i]
	}
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

//
func (t *Redis) newRedisClient(db int) *redis.Client {
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
		}
		fmt.Println("Connect Redis:", t.url, t.pwd, pong.Err())
		time.Sleep(1 * time.Second)
	}
}
