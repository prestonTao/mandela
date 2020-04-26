package main

import (
	"time"
	"mandela/core/engine"
)

func main() {
	engine.NLog.Debug(engine.LOG_file, "%s", "nihao")
	engine.NLog.Error(engine.LOG_file, "%s", "我不好")
	time.Sleep(time.Second)
}
