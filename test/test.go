package main

import (
	"mandela/core/utils"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antlabs/timer"
)

var txStr = "0700000000000000f4d6dc45153d70861e441e207fc4e507f886a3f8170745a1668d5f04f8c97422"

func main() {
	str := "vhuFoOQQ27QkiyrkJcHVUbB9ajmOp+m7W/DHT69Zzrk="
	bs, _ := base64.StdEncoding.DecodeString(str)
	fmt.Println(hex.EncodeToString(bs))

	zore := uint64(0)
	fmt.Println(int64(zore))

	//25%直接到账
	first25 := new(big.Int).Div(big.NewInt(int64(zore)), big.NewInt(int64(4)))
	//剩下的75%
	surplus := new(big.Int).Sub(big.NewInt(int64(zore)), first25)

	fmt.Println(first25.Uint64(), surplus.Uint64())

	ipStr := "127.0.0.2"
	ip := net.ParseIP("127.0.0.1")
	fmt.Println(ip)

	ipv4, err := IPString2Long(ipStr)
	ipstr, err := Long2IPString(ipv4)
	fmt.Println(ipstr, err)

	// ipv4 := net.IPv4(ip[0], ip[1], ip[2], ip[3])
	// fmt.Println(ipv4.String())

	txid, err := hex.DecodeString(txStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	key := utils.Bytes2string(txid) + "_" + strconv.Itoa(0)
	fmt.Println(key)
	strs := strings.SplitN(key, "_", 2)
	fmt.Println(len(strs), strs[0])

	// e1()

	// e2()

	// timertt()
	group()
}

// IPString2Long 把ip字符串转为数值
func IPString2Long(ip string) (uint32, error) {
	b := net.ParseIP(ip).To4()
	if b == nil {
		return 0, errors.New("invalid ipv4 format")
	}

	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24, nil
}

// Long2IPString 把数值转为ip字符串
func Long2IPString(i uint32) (string, error) {
	if i > math.MaxUint32 {
		return "", errors.New("beyond the scope of ipv4")
	}

	ip := make(net.IP, net.IPv4len)
	ip[0] = byte(i >> 24)
	ip[1] = byte(i >> 16)
	ip[2] = byte(i >> 8)
	ip[3] = byte(i)

	return ip.String(), nil
}

func group() {
	wg := new(sync.WaitGroup)
	wg.Add(0)
	wg.Wait()
}

func timertt() {
	tm := timer.NewTimer()

	tm.AfterFunc(1*time.Second, func() {
		log.Printf("after 1s\n")
	})

	tm.AfterFunc(10*time.Second, func() {
		log.Printf("after 10s\n")
	})
	// tm.Run()
	go func() {
		tm.Run()
	}()

	tm.AfterFunc(1*time.Second, func() {
		log.Printf("haha\n")
	})
	time.Sleep(time.Second * 20)
}
