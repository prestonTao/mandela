package main

import (
	"mandela/core/utils/crypto"
	"fmt"
	"time"

	"github.com/Jeiwan/eos-b58"
)

var alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func main() {
	example3()
}

func example3() {

	pre := "SELF"

	addrCoinStr := "SELFEzLXgGoCeb1tygrRt2qYTQ7MfTu7Sg8gr5"
	addrCoin := crypto.AddressFromB58String(addrCoinStr)

	fmt.Println(crypto.ValidAddr(pre, addrCoin))

	start := time.Now()
	for i := 0; i < 1000000; i++ {
		crypto.ValidAddr(pre, addrCoin)
	}
	fmt.Println("共耗时", time.Now().Sub(start))

}

func example2() {

	addrCoinStr := "SELFEzLXgGoCeb1tygrRt2qYTQ7MfTu7Sg8gr5"
	addrCoin := crypto.AddressFromB58String(addrCoinStr)

	start := time.Now()
	for i := 0; i < 10000; i++ {
		addrCoin.B58String()
	}
	fmt.Println("共耗时1", time.Now().Sub(start))

	// addrCoinStr = base64.NewEncoding(alphabet).EncodeToString(addrCoin)
	// fmt.Println("自定义转换方式", addrCoinStr)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		base58.Encode(addrCoin)
	}
	fmt.Println("共耗时2", time.Now().Sub(start))

}

func example1() {
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
