package main

import (
	"mandela/core/utils/crypto"
	"fmt"
)

func main() {
	pre := "你好"

	addr := crypto.BuildAddr(pre, []byte("fdafjkdlfajkldajf"))

	fmt.Println(crypto.ValidAddr(pre, addr))

	addrStr := addr.B58String()
	fmt.Println(addrStr)

	addr = crypto.AddressFromB58String(addrStr)
	fmt.Println(addr.B58String())

	ok := crypto.ValidAddr(pre, addr)
	fmt.Println(ok)

}
