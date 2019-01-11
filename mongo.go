package nono

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Mongo nono mongo struct
type Mongo struct {
	url      string         //连接字符串
	db       string         //数据库
	sessions []*mgo.Session //线程
	lock     []*sync.Mutex  //线程锁
	index    int            //上面数组的当前最大值
	limit    int            //超过limt/s增加线程
	max      int            //最大线程数
	times    int64          //当前调用数
	closed   bool
}

//Close all session and return
func (t *Mongo) Close() {
	t.closed = true
	for i := 0; i < t.index; i++ {
		//fmt.Println(len(t.sessions))
		t.sessions[i].Close()
	}
}

//Idbson return time by objectid
func (t *Mongo) Idbson(min int64, max int64) bson.M {

	to16 := func(i int64) bson.ObjectId {
		s := fmt.Sprintf("%x", i) + "0000000000000000" //16个0
		return bson.ObjectIdHex(s)
	}
	quary := bson.M{}
	if min != 0 {
		quary["$gte"] = to16(min)
	}
	if max != 0 {
		quary["$lte"] = to16(max)
	}
	return bson.M{"_id": quary}
}

//Insert ii must be a []
func (t *Mongo) Insert(coll string, ii interface{}) bool {
	c := t.GetMongo(coll)
	v := reflect.ValueOf(ii)
	if v.Kind() != reflect.Slice {
		return false
	}
	f := func(inter []interface{}) {
	RE:
		err := c.Insert(inter...)
		if err != nil {
			log.Println("insert err,RE:", err)
			goto RE
		}
	}
	l := v.Len()
	inter := []interface{}{}
	for i := 0; i < l; i++ {
		if i >= 5000 && i%5000 == 0 {
			f(inter)
			inter = []interface{}{}
		}
		iii := v.Index(i).Interface()
		inter = append(inter, iii)
	}
	if len(inter) > 0 {
		f(inter)
	}
	return true
}

//Or return bson.M{"$or":value}
func (t *Mongo) Or(value map[string]string) (bs bson.M) {
	findBson := []bson.M{}
	for k, v := range value {
		if k != "" && v != "" {
			findBson = append(findBson, bson.M{k: v})
		}
	}
	return bson.M{"$or": findBson}
}

//Set return bson.M{"$set":value}
func (t *Mongo) Set(field string, value interface{}) (bs bson.M) {
	return bson.M{"$set": bson.M{field: value}}
}

//Inc return bson.M{"$inc":value}
func (t *Mongo) Inc(field string, value interface{}) (bs bson.M) {
	return bson.M{"$inc": bson.M{field: value}}
}

//Range return bson.M{"$gte":"$lte"},can nil
func (t *Mongo) Range(field string, min interface{}, max interface{}) (bs bson.M) {
	rg := bson.M{}
	if min != nil {
		rg["$gte"] = min
	}
	if max != nil {
		rg["$lte"] = max
	}
	return bson.M{field: rg}
}

//GetID return inc id from 6000001
func (t *Mongo) GetID(which string) int64 {
	c := t.GetMongo("index")
	change := t.Change(t.Inc("id", 1))
	var result struct {
		Name string
		ID   int64
	}
	_, err := c.Find(bson.M{"name": which}).Apply(change, &result)
	if err != nil && err.Error() == "not found" {
		result.Name = which
		result.ID = 6000001
		c.Upsert(bson.M{"name": which}, bson.M{"$set": result})
		return result.ID
	} else if err != nil {
		fmt.Println(err)
		return -1
	}
	return result.ID
}

//Change return change after bs
func (t *Mongo) Change(bs bson.M) mgo.Change {
	change := mgo.Change{
		Update:    bs,
		ReturnNew: true,
	}
	return change
}

// NewMongo 1
func NewMongo(url, db string) *Mongo {
	log.Println("connect to mongo:", db)
	max := 18
	t := &Mongo{
		url:   url,
		db:    db,
		limit: 258 * 60,
		max:   max,
		index: 1,
		times: 0,
	}
	t.sessions = make([]*mgo.Session, max)
	t.lock = make([]*sync.Mutex, max)
	for i := 0; i < 1; i++ {
		s := t.newMongo()
		var lock sync.Mutex
		t.sessions[i] = s
		t.lock[i] = &lock
	}
	go t.incSession()
	log.Println("mongo Connected")
	return t
}
func (t *Mongo) incSession() {
	var last int64
	for {
		now := atomic.LoadInt64(&t.times)
		if last == 0 {
			last = now
		} else {
			if now-last > int64(t.limit*len(t.sessions)) && t.index < t.max {
				s := t.newMongo()
				var lock sync.Mutex
				t.sessions[t.index+1] = s
				t.lock[t.index+1] = &lock
				t.index++
			}
			if t.index >= t.max || t.closed == true {
				return
			}
			last = now
		}
		time.Sleep(1 * time.Minute)
	}
}

//GetMongo params coll return mgo.Coll
func (t *Mongo) GetMongo(coll string) *mgo.Collection {
	return t.GetMongoByDB(t.db, coll)
}

//GetSession return Session for mgo
func (t *Mongo) GetSession() *mgo.Session {
	return t.GetMongo("a").Database.Session
}

//GetMongoClone 返回一个Session的Clone
func (t *Mongo) GetMongoClone(coll string) (*mgo.Session, *mgo.Collection) {
	s := t.GetSession().Clone()
	m := s.DB(t.db).C(coll)
	return s, m
}

//GetMongoByDB params coll return mgo.Coll
func (t *Mongo) GetMongoByDB(db, coll string) *mgo.Collection {
	i := rand.Intn(t.index)
	t.lock[i].Lock()
	defer t.lock[i].Unlock()
	if t.sessions[i].Ping() != nil {
		s := t.newMongo()
		if t.sessions[i] != nil {
			temp := t.sessions[i]
			go func() {
				time.Sleep(30 * time.Second)
				temp.Close()
			}()
		}
		t.sessions[i] = s
	}
	atomic.AddInt64(&t.times, int64(1))
	return t.sessions[i].DB(db).C(coll)
}
func (t *Mongo) newMongo() (s *mgo.Session) {
	for {
		s, err := mgo.Dial(t.url)
		if err != nil {
			time.Sleep(2 * time.Second)

			fmt.Println(runtime.Caller(3))
			log.Println("mongoDial", t.url, err.Error())
		} else {
			return s
		}
	}
}
