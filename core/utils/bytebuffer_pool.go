package utils

import (
	"sync"
)

type BufferByte struct {
	lock      *sync.RWMutex
	useLength int
	bs        []byte
}

/*
	清空缓冲区
*/
func (this *BufferByte) Clean() {
	this.lock.Lock()
	this.useLength = 0
	this.lock.Unlock()
}

/*
	设置缓冲大小，只能往大了设置，不能缩小
*/
func (this *BufferByte) SetLength(maxLength int) {
	this.lock.Lock()
	this.setLength(maxLength)
	this.lock.Unlock()
}

/*
	设置缓冲大小，只能往大了设置，不能缩小
*/
func (this *BufferByte) setLength(maxLength int) {
	if maxLength <= 0 {
		return
	}
	if maxLength <= cap(this.bs) {
		return
	}
	bs := make([]byte, 0, maxLength)
	bs = append(bs, this.bs[:this.useLength]...)
	this.bs = bs
}

func (this *BufferByte) Write(bs *[]byte) {

	this.lock.Lock()
	//检查缓冲区大小是否够用
	length := cap(this.bs)
	useLength := this.useLength

	//缓冲器大小不够用，则扩容
	if length-useLength < len(*bs) {
		this.setLength(len(*bs) - (length - useLength))
	}

	n := copy(this.bs[this.useLength:], *bs)
	if n < len(*bs) {
		this.bs = append(this.bs, (*bs)[n:]...)
	}
	this.useLength += len(*bs)

	this.lock.Unlock()

}

func (this *BufferByte) Bytes() (bs *[]byte) {
	this.lock.Lock()
	temp := this.bs[:this.useLength]
	bs = &temp
	this.lock.Unlock()
	return
}

func (this *BufferByte) Length() (n int) {
	this.lock.Lock()
	n = this.useLength
	this.lock.Unlock()
	return
}

func NewBufferByte(length int) *BufferByte {
	if length < 0 {
		length = 0
	}
	return &BufferByte{
		lock:      new(sync.RWMutex),
		useLength: 0,
		bs:        make([]byte, 0, length),
	}
}
