package nono

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"time"
)

//GetMongoURL 获取本地的mongourl地址mongo.txt
func GetMongoURL() string {
	bt, err := ioutil.ReadFile("mongo.txt")
	if err != nil {
		log.Panic(err)
	}
	var result struct {
		Mongo string
	}
	_ = json.Unmarshal(bt, &result)
	return result.Mongo
}

//DBURL 1
type DBURL struct {
	Mongo    string
	Redis    string
	RedisPwd string
}

//GetDBURL 获取本地的mongourl地址mongo.txt
func GetDBURL() DBURL {
	bt, err := ioutil.ReadFile("mongo.txt")
	if err != nil {
		log.Panic(err)
	}
	var result DBURL
	_ = json.Unmarshal(bt, &result)
	return result
}

//Rand 返回最大值和最小值（包含）的随机数
func Rand(l, h int64) int64 {
	seed := rand.NewSource(time.Now().UnixNano())
	rd := rand.New(seed)
	m := h - l + 1
	if m <= 0 || m >= 1e18 {
		return h
	}
	return l + rd.Int63n(m)
}

//RandF 返回最大值和最小值浮点（包含）的随机数，默认最高保留10位精度，当字符串变长时会减少精度
//当精度输入小于10的时候,默认理解为输入错误,输入了位数而不是1en的格式
func RandF(l, h, decimal float64) float64 {
	if decimal < 10 {
		decimal = math.Pow10(int(decimal))
	}
	return float64(Rand(int64(l*decimal), int64(h*decimal))) / decimal
}
