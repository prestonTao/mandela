/*
	Package log provides 全局日志记录，内部使用beego/log

	使用方法
	utils.GlobalInit("console", "", "debug", 1)
	utils.GlobalInit("file", `{"filename":"/var/log/gd/gd.log"}`, "", 1000)
	engine.Log.Debug("session handle receive, %d, %v", msg.Code(), msg.Content())
	engine.Log.Debug("test debug")
	engine.Log.Warn("test warn")
	engine.Log.Error("test error")
*/

package utils

import (
	"errors"
	"github.com/astaxie/beego/logs"
)

var (
	logIsOpen = false
	Log       *BeegoLog
)

func GlobalInit(kind, path, level string, length int) error {
	logIsOpen = true
	Log = new(BeegoLog)
	if Log.log == nil {
		Log.log = logs.NewLogger(int64(length))
	}

	// if Log == nil {
	// 	Log = logs.NewLogger(int64(length))
	// }

	err := Log.log.SetLogger(kind, path)
	if err != nil {
		return err
	}

	switch level {
	case "debug":
		Log.log.SetLevel(logs.LevelDebug)
	case "info":
		Log.log.SetLevel(logs.LevelInfo)
	case "warn":
		Log.log.SetLevel(logs.LevelWarn)
	case "error":
		Log.log.SetLevel(logs.LevelError)
	default:
		return errors.New("未处理的日志记录等级")
	}

	return nil

}

type BeegoLog struct {
	log *logs.BeeLogger
}

func (this *BeegoLog) Info(format string, v ...interface{}) {
	if logIsOpen {
		this.log.Info(format, v...)
	}
}
func (this *BeegoLog) Debug(format string, v ...interface{}) {
	if logIsOpen {
		this.log.Debug(format, v...)
	}
}
func (this *BeegoLog) Warn(format string, v ...interface{}) {
	if logIsOpen {
		this.log.Warn(format, v...)
	}
}
func (this *BeegoLog) Error(format string, v ...interface{}) {
	if logIsOpen {
		this.log.Error(format, v...)
	}
}
