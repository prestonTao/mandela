package main

import (
	tmn "mandela/core/transmission"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

func main() {

	seed := "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().Unix())
	var b []byte
	fmt.Println("start")
	for i := 0; i < 1024*1024; i++ {
		b = append(b, seed[rand.Intn(len(seed))])
	}
	str := hex.EncodeToString(b)
	fmt.Println("end")
	s := time.Now()
	fmt.Println(len(str))
	prk1, puk2 := tmn.CreatePrkPuk()

	tmn.InitDh()
	fmt.Println(tmn.PPJSON.Prik, tmn.PPJSON.Pubk)

	key1, err := tmn.GetKey(tmn.PPJSON.Prik, puk2)
	fmt.Println("协商密钥A", key1, err)

	key2, err := tmn.GetKey(prk1, tmn.PPJSON.Pubk)
	fmt.Println("协商密钥B", key2, err)

	rs, err := tmn.Encrypt(tmn.Stob(key1), []byte(str))
	//fmt.Println(rs, err)
	rss, err := tmn.Decrypt(tmn.Stob(key2), rs)
	fmt.Println(len(string(rss)), err)
	fmt.Println(time.Now().Sub(s))
}
