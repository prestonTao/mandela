/**
占用空间测试
*/
package main

import (
	"mandela/store/fs"
	"fmt"
)

func main() {
	ch := make(chan int)
	s := fs.NewSpace()
	go s.Init()
	free := s.FreeSpace()
	fmt.Println(free)
	//s.Set([]byte("ok"), []byte("value"))
	<-ch
}
