package utils

import (
	// "fmt"
	"sync"
)

var task *Task
var taskClass = new(sync.Map)

func init() {
	task = NewTask(taskFunc)
}

func taskFunc(class string, params []byte) {
	v, ok := taskClass.Load(class)
	if !ok {
		// fmt.Println("未注册的定时器类型", class)
		return
	}
	fn := v.(f)
	fn(class, params)
}

/*
	添加一个定时任务
	@tick    int64    未来的某个时刻,例如(未来10秒钟后执行)：time.Now().Unix()+int64(10)
*/
func AddTimetask(tick int64, fn f, class string, params []byte) {
	taskClass.Store(class, fn)
	task.Add(tick, class, params)
}
