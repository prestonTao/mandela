package utils

import (
	"sync"
)

type SyncList struct {
	lock *sync.RWMutex
	list []interface{}
}

func (this *SyncList) Add(obj interface{}) {
	this.lock.Lock()
	this.list = append(this.list, obj)
	this.lock.Unlock()
}

func (this *SyncList) Get(index int) (n interface{}) {
	this.lock.RLock()
	n = this.list[index]
	this.lock.RUnlock()
	return
}

func (this *SyncList) GetAll() (l []interface{}) {
	this.lock.RLock()
	l = this.list
	this.lock.RUnlock()
	return
}

//func (this *SyncList) Range(fn func(i int, v interface{}) bool) {
//	this.lock.Lock()
//	for i, one := range this.list {
//		if !fn(i, one) {
//			break
//		}
//	}
//	this.lock.Unlock()
//}

//func (this *SyncList) Find() {
//	for _,one := range this.list{
//		if
//	}
//}

func NewSyncList() *SyncList {
	return &SyncList{
		lock: new(sync.RWMutex),
		list: make([]interface{}, 0),
	}
}
