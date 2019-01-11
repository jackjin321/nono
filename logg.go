package nono

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	INFO = "[I]"
	//ERR 错误
	ERR = "[E]"
	//WARN 警告
	WARN = "[W]"
	//IMP 重要
	IMP = "[P]"
)

// var logs *Logg
var lg *logg
var once sync.Once

func init() {
	// logs = &Logg{db: "logs", col: "unclassfied"}
	// logs.ch = make(chan interface{}, 99999)
	//TODO
	//logs.mongoURL = "mongodb://logsuser:logsuserpwd@10.0.0.49:13149"
	lg = &logg{}
	//lg.SaveFile()
	log.SetOutput(lg)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type logg struct {
	name  string
	write *os.File
}

func (t *logg) Write(w []byte) (n int, err error) {
	n, err = fmt.Printf("%s", w)
	once.Do(t.SaveFile)
	return t.write.Write(w)
}

// func (t *logg) saveRds() {
// 	r := nono.NewRedis("127.0.0.1:6379", "", 0)
// }
func (t *logg) SaveFile() {
	t.getname()
	now := time.Now()
	filename := t.name + "-" + now.Format("2006-01-02") + ".log"
	var logfile *os.File
	var err error
	logfile, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		//fmt.Println(err)
		logfile, err = os.Create(filename)
		if err != nil {
			fmt.Println("error save log", err)
		}
	}
	//logfile.Write([]byte(strings.Join(os.Args, "-")))
	lastlog := logfile
	t.write = logfile

	go func() {
		for {
			if time.Now().Day() != now.Day() {
				name := t.name + "-" + time.Now().Format("2006-01-02")
				file, err := os.Create(name)
				if err != nil {
					fmt.Println("error save log", err)
					time.Sleep(1 * time.Second)
					continue
				}
				file.Write([]byte(strings.Join(os.Args, "-")))
				t.write = file
				time.Sleep(1 * time.Hour)
				lastlog.Close()
				lastlog = file
				now = time.Now()
			}
			time.Sleep(360 * time.Second)
		}
	}()
}
func (t *logg) getname() (s string) {
	for i, arg := range os.Args {
		if i == 0 {
			s = filepath.Base(arg)
		} else {
			s = s + " " + arg
		}
	}
	t.name = s
	return
}

// // Logg 日志结构
// type Logg struct {
// 	db       string
// 	col      string
// 	mongoURL string
// 	stop     bool
// 	ch       chan interface{}
// }

// //Logs 日志结构2
// type Logs struct {
// 	Tm interface{}
// 	Tp interface{}
// 	M  interface{}
// }

// //SetCollAndStart 按照输入coll和mongourl启动日志记录
// func SetCollAndStart(coll string, mongoURL string) {
// 	logs.col = coll
// 	logs.mongoURL = mongoURL
// 	if logs.mongoURL != "" {
// 		go logs.start()
// 	}
// }
// func (t *Logg) push(v []interface{}) {
// 	if len(v) == 0 {
// 		return
// 	}
// 	s, err := mgo.Dial(t.mongoURL)
// 	if !Noerr(err) { //连接失败就把Log加入到列表末尾
// 		log.Println("LOGERR", err.Error())
// 		for _, vv := range v {
// 			t.ch <- vv
// 		}
// 		return
// 	}
// 	defer s.Close()
// 	c := s.DB(t.db).C(t.col)
// 	c.Insert(v...)
// }
// func (t *Logg) start() {
// 	tm := time.NewTimer(time.Second) //超时
// 	pushLog := []interface{}{}       //需要增加的Log

// 	push := func() { //增加log并清空
// 		t.push(pushLog)
// 		pushLog = []interface{}{}
// 	}
// 	for {
// 		tm.Reset(time.Second)
// 		select {
// 		case s := <-t.ch:
// 			pushLog = append(pushLog, s)
// 			if len(pushLog) > 500 {
// 				go push()
// 			}
// 		case <-tm.C:
// 			go push()
// 		}
// 	}

// }
// func (t *Logg) pl(v ...interface{}) {
// 	log.Println(v...)
// 	if len(v) == 0 {
// 		return
// 	}
// 	temp := Logs{}
// 	temp.Tm = Time2S(time.Now())
// 	if len(v) == 1 {
// 		temp.M = v[0]
// 	} else {
// 		temp.Tp = v[0]
// 		temp.M = fmt.Sprintln(v...)
// 	}
// 	if len(t.mongoURL) > 10 {
// 		t.ch <- temp
// 	}
// }

// //Println 输出行
// func Println(v ...interface{}) {
// 	logs.pl(v...)
// }
// func getline() string {
// 	_, file, line, _ := runtime.Caller(2)
// 	return fmt.Sprintln(file, line)
// }

// //Printerr 输出错误
// func Printerr(v ...interface{}) {
// 	Println("[ERROR]", getline(), v)
// }

// //Printinfo 输出正常日志
// func Printinfo(v ...interface{}) {
// 	Println("[INFO]", getline(), v)
// }

// //Printwarn 输出警告
// func Printwarn(v ...interface{}) {
// 	Println("[WARN]", getline(), v)
// }

// //PrintDebug 输出调试
// func PrintDebug(v ...interface{}) {
// 	Println("[DEBUG]", getline(), v)
// }

// //AllOutPut 把一个对象格式化然后输出文本
// func AllOutPut(v interface{}) {
// 	js, err := json.Marshal(v)
// 	if err != nil {
// 		fmt.Println("EEERROR" + err.Error())
// 	} else {
// 		fmt.Println(string(js))
// 	}
// }
