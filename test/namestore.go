package main

import (
	"fmt"
	"mandela/core/cache_store"
	"time"
)

func main() {
	go func() {
		id := cache_store.GetAddressInName("tao")
		fmt.Println(id)

		id = cache_store.GetAddressInName("tao")
		fmt.Println(id)
	}()

	go func() {
		time.Sleep(time.Second * 5)
		name := cache_store.NewName("tao", [][]byte{[]byte{1, 2, 3}}, []byte{4, 5, 6})
		cache_store.AddAddressInName("tao", name)
	}()

	time.Sleep(time.Second * 15)

}
