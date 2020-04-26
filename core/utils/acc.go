package utils

import (
	"sync/atomic"
)

var acc uint64 = 0

/*
	获得累加器的值
*/
func GetAccNumber() uint64 {
	return atomic.AddUint64(&acc, 1)
}
