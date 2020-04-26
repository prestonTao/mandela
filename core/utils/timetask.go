package utils

import (
	"bytes"
	//	"fmt"
	"sort"
	"sync"
	"time"
)

type f func(class string, params []byte)

type Task struct {
	lock     *sync.RWMutex
	task     []*Ticks
	taskMap  map[int64]*Ticks
	function f
	isStart  bool
	nowTick  *Ticks
	update   chan bool
}

func (this *Task) Len() int {
	return len(this.task)
}

func (this *Task) Less(i, j int) bool {
	return this.task[i].tick < this.task[j].tick // 按值排序
}

func (this *Task) Swap(i, j int) {
	this.task[i], this.task[j] = this.task[j], this.task[i]
}

func (this *Task) start() {
	for {
		this.lock.RLock()
		if len(this.task) <= 0 {
			//			go fmt.Println("长度太小")
			this.lock.RUnlock()
			ticker := time.NewTicker(time.Hour * 24)
			select {
			case <-ticker.C:
			case <-this.update:
				ticker.Stop()
				//				go fmt.Println("更新1")
			}
			continue
		} else {
			this.lock.RUnlock()
		}

		now := time.Now().Unix()
		if this.task[0].tick <= now {
			//fmt.Println("超时待处理")
			for _, one := range this.task[0].ticks {
				this.function(one.class, one.params)
			}
			this.lock.Lock()
			delete(this.taskMap, this.task[0].tick)
			//			this.task = make([]Task)
			//			this.task = append([]*Ticks{}, (*this.task)[1:]...)
			temp := make([]*Ticks, 0)
			for i := 1; i < len(this.task); i++ {
				temp = append(temp, this.task[i])
			}
			this.task = temp
			this.lock.Unlock()
		} else {
			//			go fmt.Println("设置时间间隔 ", this.task[0].tick-now)
			ticker := time.NewTicker(time.Second * time.Duration(this.task[0].tick-now))
			//			<-this.update
			select {
			case <-ticker.C:
				//fmt.Println("时间到 ", this.task[0].ticks)
				for _, one := range this.task[0].ticks {
					this.function(one.class, one.params)
				}
				this.lock.Lock()
				delete(this.taskMap, this.task[0].tick)
				//			this.task = make([]Task)
				//				this.task = append([]Ticks{}, this.task[1:]...)
				temp := make([]*Ticks, 0)
				for i := 1; i < len(this.task); i++ {
					temp = append(temp, this.task[i])
				}
				this.task = temp
				this.lock.Unlock()
			case <-this.update:
				ticker.Stop()
				//fmt.Println("更新2")
			}
		}
		//		for _, one := range this.task {
		//			fmt.Println(one.tick)
		//		}
	}
}

/*
	@tick    int64    执行的时间（时间戳）
	@class   string   执行的命令
	@params  string   字符串参数
*/
func (this *Task) Add(tick int64, class string, params []byte) {
	//	go fmt.Println("add")
	newTick := NewTick(tick, class, params)
	this.lock.Lock()
	//	if len(this.task) == 0 {
	//		this.task = *newTick
	//		this.taskMap[tick] = newTick
	//	} else {
	ticks, ok := this.taskMap[tick]
	if ok {
		//fmt.Println("befor ", this.taskMap[tick])
		ticks.lock.Lock()
		ticks.ticks = append(ticks.ticks, *newTick)
		ticks.lock.Unlock()
		//fmt.Println("after ", this.taskMap[tick])
	} else {
		ticks := NewTicks(*newTick)
		this.task = append(this.task, ticks)
		this.taskMap[tick] = ticks
	}
	sort.Sort(this)
	//	}
	this.lock.Unlock()
	select {
	case this.update <- false:
		//		go fmt.Println("放进去了")
	default:
		//		go fmt.Println("放不进去就算了")
	}
}

func (this *Task) Remove(tick int64, class string, params []byte) {
	this.lock.Lock()
	ticks, ok := this.taskMap[tick]
	if ok {
		ticks.lock.Lock()
		//		ticks.ticks = append(ticks.ticks, *newTick)
		for i, one := range ticks.ticks {
			if one.class == class && bytes.Equal(one.params, params) {
				//这里可以优化内存
				temp := ticks.ticks[:i]
				ticks.ticks = append(temp, ticks.ticks[i+1:]...)
				break
			}
		}
		ticks.lock.Unlock()
	}
	sort.Sort(this)
	this.lock.Unlock()
	select {
	case this.update <- false:
	default:
	}
}

func NewTask(function f) *Task {
	task := Task{
		lock:     new(sync.RWMutex),
		task:     make([]*Ticks, 0),
		taskMap:  make(map[int64]*Ticks),
		function: function,
		update:   make(chan bool, 1),
	}
	go task.start()
	return &task
}

type Ticks struct {
	lock  *sync.RWMutex
	tick  int64
	ticks []Tick
}

func NewTicks(tick Tick) *Ticks {
	return &Ticks{
		lock:  new(sync.RWMutex),
		tick:  tick.tick,
		ticks: []Tick{tick},
	}
}

type Tick struct {
	tick   int64
	class  string
	params []byte
}

func NewTick(tick int64, class string, params []byte) *Tick {
	return &Tick{
		tick:   tick,
		class:  class,
		params: params,
	}
}
