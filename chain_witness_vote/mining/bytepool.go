package mining

import (
	"mandela/core/utils"
	"sync"
)

var bytePool = sync.Pool{
	New: func() interface{} {
		// b := make([]byte, 1024)
		buf := utils.NewBufferByte(0)
		return &buf
	},
}
