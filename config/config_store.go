package config

import (
	"sync/atomic"
)

const (
	Chunk_size = 10 * 1024 * 1024  //每块10M 单位byte
	Spacenum   = 1024 * Chunk_size //总占用空间 10G  单位byte
)

var (
	spaceTotal     = uint64(0) //保存全网提供空间总量
	spaceTotalAddr = uint64(0) //保存全网提供空间地址数量
)

func GetSpaceTotal() uint64 {
	return atomic.LoadUint64(&spaceTotal)
}

func SetSpaceTotal(st uint64) {
	atomic.StoreUint64(&spaceTotal, st)
}

func GetSpaceTotalAddr() uint64 {
	return atomic.LoadUint64(&spaceTotalAddr)
}

func SetSpaceTotalAddr(sta uint64) {
	atomic.StoreUint64(&spaceTotalAddr, sta)
}
