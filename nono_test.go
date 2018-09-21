package nono

import (
	"fmt"
	"testing"
	"time"
)

func Test_Any(tt *testing.T) {
	a := ""
	for i := 0; i < 1000; i++ {
		a += "a"
	}
	now := time.Now()
	b := ""
	for i := 0; i < 1000000; i++ {
		b = AesEncrypt(a, "abc")
	}
	fmt.Println(time.Since(now).Seconds())
	now = time.Now()
	for i := 0; i < 1000000; i++ {
		AesDecrypt(b, "abc")
	}
	fmt.Println(time.Since(now).Seconds())
}
