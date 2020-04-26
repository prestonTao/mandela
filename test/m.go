package main

import (
	"crypto/sha256"
	"fmt"
	gconfig "mandela/config"
	"mandela/core/utils"
)

func main() {
	var err error
	var h utils.Multihash
	t := sha256.Sum256([]byte("sdfasdfadsf"))
	mhBs, _ := utils.Encode(t[:], gconfig.HashCode)
	h = utils.Multihash(mhBs)
	fmt.Printf("%+v\n", h)
	fmt.Println(h.B58String())
	fmt.Println(h.B58String())
	fmt.Println(h.B58String())
	h, err = utils.FromB58String("W1hT14kpPGZi5oiyY3HPk5qsCUTfsNb4knrsZUDyTEbnib")
	fmt.Println(h.B58string, err)
	fmt.Println(h.HexString())
	h, err = utils.FromHexString("162076de7043d12ee831e023c8fff830edb2f6903ca5d690d2f650643ed646ff8520")
	fmt.Println(h.B58string, err)
}
