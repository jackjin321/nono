package nono

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/globalsign/mgo"
	"golang.org/x/crypto/ssh/terminal"
)

func Readline(s string) (result string) {
	fmt.Println(s)
	fmt.Scanln(&result)
	return result
}
func ReadPwd(s string) (result string) {
	fmt.Println(s)
	pwd, err := terminal.ReadPassword(0)
	if err != nil {
		return err.Error()
	}
	return string(pwd)
}

const (
	INFO = "[INFO]"
	ERR  = "[ERROR]"
	WARN = "[WARN]"
	IMP  = "[IMP]"
)

var logs *Logg

func init() {
	logs = &Logg{db: "logs", col: "unclassfied"}
	logs.ch = make(chan interface{}, 99999)
	//TODO
	//logs.mongoURL = "mongodb://logsuser:logsuserpwd@10.0.0.49:13149"
}

type Logg struct {
	db       string
	col      string
	mongoURL string
	stop     bool
	ch       chan interface{}
}
type Logs struct {
	Tm interface{}
	Tp interface{}
	M  interface{}
}

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
func Println(v ...interface{}) {
	logs.pl(v...)
}
func Printerr(v ...interface{}) {
	Println("[ERROR]", v)
}
func Printinfo(v ...interface{}) {
	Println("[INFO]", v)
}
func Printwarn(v ...interface{}) {
	Println("[WARN]", v)
}
func PrintDebug(v ...interface{}) {
	Println("[DEBUG]", v)
}
func AllOutPut(v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		fmt.Println("EEERROR" + err.Error())
	} else {
		fmt.Println(string(js))
	}
}
