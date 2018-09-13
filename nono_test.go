package nono

import (
	"fmt"
	"testing"
)

func Test_Any(tt *testing.T) {
	fmt.Println(Hex2Diff("0x0000000112e0be826d694b2e62d01511f12a6061fbaec8bc02357593e70e52ba").Int64())
}
