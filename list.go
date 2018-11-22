package nono

import (
	"sync/atomic"
)

//Quence 1
type Quence struct {
	c chan interface{}
	l int64
}

//NewQuence 新建表
func NewQuence(l int64) *Quence {
	return &Quence{make(chan interface{}, l), l}
}

//Push 推
func (t *Quence) Push(i interface{}) bool {
	select {
	case t.c <- i:
		atomic.AddInt64(&t.l, 1)
		return true
	default:
		return false
	}
}

//Pull 拉
func (t *Quence) Pull() (interface{}, bool) {
	//var result interface{}
	select {
	case result := <-t.c:
		atomic.AddInt64(&t.l, -1)
		return result, true
	default:
		return nil, false
	}
}

//Len 1
func (t *Quence) Len() int64 {
	return atomic.LoadInt64(&t.l)
}

//Close 1
func (t *Quence) Close() {
	close(t.c)
}
