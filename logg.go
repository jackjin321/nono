package nono

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/globalsign/mgo"
	"golang.org/x/crypto/ssh/terminal"
)

//Readline 真漂亮
func Readline(s string) (result string) {
	fmt.Println(s)
	fmt.Scanln(&result)
	return result
}

//ReadPwd 隐藏的模式读取字符串
func ReadPwd(s string) (result string) {
	fmt.Println(s)
	pwd, err := terminal.ReadPassword(0)
	if err != nil {
		return err.Error()
	}
	return string(pwd)
}

const (
	//INFO 信息
	INFO = "[INFO]"
	//ERR 错误
	ERR = "[ERROR]"
	//WARN 警告
	WARN = "[WARN]"
	//IMP 重要
	IMP = "[IMP]"
)

var logs *Logg

func init() {
	logs = &Logg{db: "logs", col: "unclassfied"}
	logs.ch = make(chan interface{}, 99999)
	//TODO
	//logs.mongoURL = "mongodb://logsuser:logsuserpwd@10.0.0.49:13149"
}

// Logg 日志结构
type Logg struct {
	db       string
	col      string
	mongoURL string
	stop     bool
	ch       chan interface{}
}

//Logs 日志结构2
type Logs struct {
	Tm interface{}
	Tp interface{}
	M  interface{}
}

//SetCollAndStart 按照输入coll和mongourl启动日志记录
func SetCollAndStart(coll string, mongoURL string) {
	logs.col = coll
	logs.mongoURL = mongoURL
	if logs.mongoURL != "" {
		go logs.start()
	}
}
func (t *Logg) push(v []interface{}) {
	if len(v) == 0 {
		return
	}
	s, err := mgo.Dial(t.mongoURL)
	if !Noerr(err) { //连接失败就把Log加入到列表末尾
		log.Println("LOGERR", err.Error())
		for _, vv := range v {
			t.ch <- vv
		}
		return
	}
	defer s.Close()
	c := s.DB(t.db).C(t.col)
	c.Insert(v...)
}
func (t *Logg) start() {
	tm := time.NewTimer(time.Second) //超时
	pushLog := []interface{}{}       //需要增加的Log

	push := func() { //增加log并清空
		t.push(pushLog)
		pushLog = []interface{}{}
	}
	for {
		tm.Reset(time.Second)
		select {
		case s := <-t.ch:
			pushLog = append(pushLog, s)
			if len(pushLog) > 500 {
				go push()
			}
		case <-tm.C:
			go push()
		}
	}

}
func (t *Logg) pl(v ...interface{}) {
	log.Println(v...)
	if len(v) == 0 {
		return
	}
	temp := Logs{}
	temp.Tm = Time2S(time.Now())
	if len(v) == 1 {
		temp.M = v[0]
	} else {
		temp.Tp = v[0]
		temp.M = fmt.Sprintln(v...)
	}
	if len(t.mongoURL) > 10 {
		t.ch <- temp
	}
}

//Println 输出行
func Println(v ...interface{}) {
	logs.pl(v...)
}

//Printerr 输出错误
func Printerr(v ...interface{}) {
	Println("[ERROR]", v)
}

//Printinfo 输出正常日志
func Printinfo(v ...interface{}) {
	Println("[INFO]", v)
}

//Printwarn 输出警告
func Printwarn(v ...interface{}) {
	Println("[WARN]", v)
}

//PrintDebug 输出调试
func PrintDebug(v ...interface{}) {
	Println("[DEBUG]", v)
}

//AllOutPut 把一个对象格式化然后输出文本
func AllOutPut(v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		fmt.Println("EEERROR" + err.Error())
	} else {
		fmt.Println(string(js))
	}
}
