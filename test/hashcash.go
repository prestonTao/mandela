package main

import (
	"mandela/core/utils"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

func main() {
	example1()
}

func example1() {
	start := time.Now()
	key := utils.Work("taopopoo@126.com", 20)

	tick := time.Now().Sub(start)

	ok := utils.Check("taopopoo@126.com", 20, key)
	if ok {
		fmt.Println("work done!")
	}

	fmt.Println("end", tick)
}

func example() {
	msg := "taopopoo"
	i := 0
	start := time.Now()
	for ; ; i++ {
		digest := utils.Hash_SHA3_256([]byte(msg + strconv.Itoa(i)))
		if CheckNonce(digest, 25) {
			break
		}
	}
	tick := time.Now().Sub(start)

	//	nonce := utils.Work(msg, zoroes)
	//	check := utils.Check(msg, zoroes, nonce)
	fmt.Println("end", i, tick)
}

/*
	检查hash值是否有足够多的前导零
	@code      string    Hash值
	@zeroes    uint64    前导零个数
*/
func CheckNonce(code []byte, zeroes uint64) bool {
	digestHex := new(big.Int).SetBytes(code)
	surplus := new(big.Int).Rsh(digestHex, uint(256-zeroes))
	zero := big.NewInt(0)
	result := zero.Cmp(surplus)
	if result == 0 {
		return true
	}
	return false
}
