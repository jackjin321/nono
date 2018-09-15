package nono

import (
	"fmt"
	"testing"
)

func Test_Any(tt *testing.T) {
	a := AesEncrypt(`{"name":"zhangsan","age":23,"email":"chentging@aliyun.com"}`, "321")
	fmt.Println(a)
	b := AesDecrypt(a, "321")
	fmt.Println(b)
}
