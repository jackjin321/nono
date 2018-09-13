package nono

import (
	"testing"
	"time"
)

func Test测试增加日志(t *testing.T) {
	logs.mongoURL = "127.0.0.1:27018"
	SetCollAndStart("test")
	for i := 0; i < 20; i++ {
		Println(IMP, i)

	}
	time.Sleep(60 * time.Second)
	logs.mongoURL = "127.0.0.1:27017"
	time.Sleep(10 * time.Second)
}
