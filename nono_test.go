package nono

import (
	"fmt"
	"testing"
)

func Test_Any(tt *testing.T) {
	for {
		a := ReadPwd("input pwd")
		fmt.Println(":::", a)
	}
}
