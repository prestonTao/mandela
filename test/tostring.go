package main

import (
	"mandela/cache"
	"fmt"
)

func main() {
	b58 := cache.To58String([]byte("ok"))
	fmt.Println(b58)
}
