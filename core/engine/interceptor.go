package engine

import (
	"sync"
)

type Interceptor interface {
	In(c Controller, msg Packet) bool
	Out(c Controller, msg Packet)
}

type InterceptorProvider struct {
	lock         *sync.RWMutex
	interceptors []Interceptor
}

func (this *InterceptorProvider) addInterceptor(itpr Interceptor) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.interceptors = append(this.interceptors, itpr)
}
func (this *InterceptorProvider) getInterceptors() []Interceptor {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.interceptors

}

func NewInterceptor() *InterceptorProvider {
	interceptor := new(InterceptorProvider)
	interceptor.lock = new(sync.RWMutex)
	interceptor.interceptors = make([]Interceptor, 0)
	return interceptor
}
