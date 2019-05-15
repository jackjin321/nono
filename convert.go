package nono

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"time"

	mathETH "github.com/ethereum/go-ethereum/common/math"
)

//Noerr 这个可以定位err的地址
func Noerr(err error) bool {
	return NoerrByDepth(err, 2)
}

// NoerrByDepth 按照深度来输出错误
func NoerrByDepth(err error, depth int) bool {
	if err != nil {
		//pc:线程,file:文件名,line:行号 f.name:调用函数
		pc, file, line, ok := runtime.Caller(depth)
		if ok {
			f := runtime.FuncForPC(pc)
			log.Println("[ERROR]", f.Name(), err.Error())
			log.Println("[ERROR]", file, line)
		}
		return false
	}
	return true
}

//S2f 字符串转换float64忽略错误,如果错误返回0
func S2f(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0.0
	}
	return f
}

//S2i 字符串转换成int64,错误返回0
func S2i(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		i = 0
	}
	return i
}

//Hex2i 16进制字符串转换成int,这里必须确认数字不会越界,否则 还是用big.int的好
func Hex2i(s string) int64 {
	b := Hex2Big(s)
	if b == nil {
		return 0
	}
	if b.IsInt64() {
		return b.Int64()
	}
	return 0
}

//Hex2Big 16进制字符串转换成int
func Hex2Big(s string) (b *big.Int) {
	if len(s) <= 2 {
		return nil
	}
	if string(s[:2]) == "0x" {
		b, _ = new(big.Int).SetString(s, 0)
	} else {
		b, _ = new(big.Int).SetString(s, 16)
	}

	return
}

//I2Hex 1
func I2Hex(i int64) string {
	return fmt.Sprintf("0x%02x", i)
}

//Sha 进行sha256加密
func Sha(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	md := hash.Sum(nil)
	return hex.EncodeToString(md)
}

//Sha512 进行sha512加密
func Sha512(s string) string {
	hash := sha512.New()
	hash.Write([]byte(s))
	md := hash.Sum(nil)
	return hex.EncodeToString(md)
}

//GetIP 吧192.168.1.1:3389 这样的字符串返回ipbuff
func GetIP(s string) string {
	sArray := strings.Split(s, `:`)
	if len(sArray) == 2 {
		return sArray[0]
	}
	return s
}

//F2s 暂时别用
func F2s(f interface{}, n string) string {
	switch f.(type) {
	case float64, float32:
		ff := fmt.Sprintf("%."+n+"f", f)
		return ff
	case int64:
		return strconv.FormatInt(f.(int64), 10)
	case int:
		return strconv.Itoa(f.(int))
	}
	return ""
}

//I2s 暂时别用
func I2s(f interface{}) string {
	switch f.(type) {
	case int64:
		return strconv.FormatInt(f.(int64), 10)
	case int:
		return strconv.Itoa(f.(int))
	}
	return ""
}

//Time2S 从各种时间格式转换成字符串包含以下
//float32/64,可以直接去掉小数位的unix时间
//time
//int32/64
//string格式的unix时间
func Time2S(f interface{}) string {
	var unix int64
	switch f.(type) {
	case time.Time:
		unix = f.(time.Time).Unix()
	case string:
		ss := f.(string)
		unix = S2i(ss)
		if unix == 0 {
			if len(ss) > 8 {
				ss = string(ss[:8])
			}
			unix = Hex2i(ss)
		}
	case int:
		unix = int64(f.(int))
	case float32:
		unix = int64(f.(float32))
	case float64:
		unix = int64(f.(float64))
	case int64:
		unix = f.(int64)
	}
	for unix > 8537950667 {
		unix = unix / 1000
	}
	if unix == 0 {
		return fmt.Sprint(f)
	}
	tm := time.Unix(unix, 0)
	return tm.Format("2006-01-02 15:04:05")
}

//Bsonid2time 通过mongodb的_id转换成可以
func Bsonid2time(s string) string {
	s = string(s[:8])
	i := Hex2i(s)
	return Time2S(i)
}

//S2Time 2006-01-02 15:04:05的格式转换成time
func S2Time(s string) time.Time {
	lo, _ := time.LoadLocation("Local")
	tm, _ := time.ParseInLocation("2006-01-02 15:04:05", s, lo)
	return tm
}

//StartAtTime 定时启动,会阻塞,当hour设置>24时候,直接返回
func StartAtTime(hours int, minute int) bool {
	if hours > 24 || minute > 60 {
		return true
	}
	now := time.Now()
	var t time.Time
	if now.Hour() > hours {
		t = now.Add(24 * time.Hour)
	} else if now.Hour() == hours {
		if now.Minute() > minute {
			t = now.Add(24 * time.Hour)
		} else {
			t = now
		}
	} else {
		t = now
	}
	log.Println("Will Start At:", hours, ":", minute)
	t = time.Date(t.Year(), t.Month(), t.Day(), hours, minute, 0, 0, t.Location())
	s := time.NewTimer(t.Sub(now))
	<-s.C
	//println(time.Since(now).String())
	return true
}

var pow256 = mathETH.BigPow(2, 256)

//Hex2Diff 暂时别用
func Hex2Diff(s string) *big.Int {
	bt := Hex2b(s)
	i := new(big.Int).Div(pow256, new(big.Int).SetBytes(bt))
	return i
}

//Diff2Hex 暂时别用
func Diff2Hex(diff int64) string {
	diff1 := big.NewInt(diff)
	diff2 := new(big.Int).Div(pow256, diff1)
	return B2hex(diff2.Bytes())
}

//B2hex 暂时别用
func B2hex(b []byte) string {
	hex := hex.EncodeToString(b)
	// Prefer output of "0x0" instead of "0x"
	if len(hex) == 0 {
		hex = "0"
	}
	return "0x" + hex
}

//Hex2b 暂时别用.
func Hex2b(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	b, _ := hex.DecodeString(s)
	return b
}
