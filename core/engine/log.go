// Package log provides 全局日志记录，内部使用beego/log
package engine

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/astaxie/beego/logs"
)

const (
	LOG_default          = 0 //
	LOG_console          = 1 //打印到控制台
	LOG_file             = 2 //打印到文件
	LOG_console_and_file = 3 //同时打印到控制台和文件

	log_level_info    = 1
	log_level_debug   = 2
	log_level_warning = 3
	log_level_error   = 4
)

var NLog *LogQueue = NewLog(LOG_console, "log.txt")

//func init() {
//	Init(LOG_console, "log.txt")
//}

type LogQueue struct {
	t        int
	filepath string
	queue    chan *LogOne
}

func (this *LogQueue) waite() {
	//	fmt.Println("90909090")
	//	for one := range this.queue {
	for {
		one := <-this.queue
		this.pro(one)
	}
}

func (this *LogQueue) pro(one *LogOne) {
	//	fmt.Println("000", one)
	level := " [?] "
	switch one.logLevel {
	case log_level_info:
		level = " [I] "
	case log_level_debug:
		level = " [D] "
	case log_level_warning:
		level = " [W] "
	case log_level_error:
		level = " [E] "
	}
	str := one.date.Format("2006-01-02 15:04:05.000") + level + fmt.Sprintf(one.format, one.content...) + "\n"
	if one.t == LOG_default {
		one.t = this.t
	}
	switch one.t {
	case LOG_console:
		//		fmt.Println(str)
	case LOG_file:
		tracefile(this.filepath, str)
	case LOG_console_and_file:
		fmt.Println(str)
		tracefile(this.filepath, str)
	}
}

func NewLog(t int, filepath string) *LogQueue {
	log := &LogQueue{
		t:        t,
		filepath: filepath,
		queue:    make(chan *LogOne, 10000),
	}
	go log.waite()
	return log
}

/*
	@t    int    打印日志类型，0=系统默认类型
*/
func (this *LogQueue) Info(t int, format string, str ...interface{}) {
	one := LogOne{
		date:     time.Now(),
		logLevel: log_level_info,
		t:        t,
		format:   format,
		content:  str,
	}
	select {
	case this.queue <- &one:
	default:
	}
}

/*
	@t    int    打印日志类型，0=系统默认类型
*/
func (this *LogQueue) Debug(t int, format string, str ...interface{}) {
	one := LogOne{
		date:     time.Now(),
		logLevel: log_level_debug,
		t:        t,
		format:   format,
		content:  str,
	}
	select {
	case this.queue <- &one:
	default:
	}
}

/*
	@t    int    打印日志类型，0=系统默认类型
*/
func (this *LogQueue) Error(t int, format string, str ...interface{}) {
	one := LogOne{
		date:     time.Now(),
		logLevel: log_level_error,
		t:        t,
		format:   format,
		content:  str,
	}
	select {
	case this.queue <- &one:
		//		fmt.Println("---")
		//		this.pro(&one)
	default:
		//		fmt.Println("---2")
	}
	//	fmt.Println("---3")
}

type LogOne struct {
	date     time.Time
	logLevel int
	t        int           //日志输出类型
	format   string        //
	content  []interface{} //
}

//打印内容到文件中
//tracefile(fmt.Sprintf("receive:%s",v))
func tracefile(filepath, str_content string) {
	fd, _ := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	//	fd_time := time.Now().Format("2006-01-02 15:04:05")
	//	fd_content := strings.Join([]string{"======", fd_time, "=====", str_content, "\n"}, "")
	buf := []byte(str_content)
	fd.Write(buf)
	fd.Close()
}

//import (
//	"errors"

//	"github.com/astaxie/beego/logs"
//)
//beego Log
var Log *logs.BeeLogger

func init() {
	GlobalInit("file", `{"filename":"logs/log.txt"}`, "", 1000)
	//	GlobalInit("console", "", "debug", 1)
}

func GlobalInit(kind, path, level string, length int) error {
	if Log == nil {
		Log = logs.NewLogger(int64(length))
	}
	Log.EnableFuncCallDepth(true)
	err := Log.SetLogger(kind, path)
	if err != nil {
		return err
	}

	switch level {
	case "debug":
		Log.SetLevel(logs.LevelDebug)
	case "info":
		Log.SetLevel(logs.LevelInfo)
	case "warn":
		Log.SetLevel(logs.LevelWarn)
	case "error":
		Log.SetLevel(logs.LevelError)
	default:
		//未处理的日志记录等级
		return errors.New("Unprocessed logging level")
	}

	return nil

}
