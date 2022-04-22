package utils

import (
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"time"
)

/*
	字符串中随机选择一个
*/
func RandString(strs ...string) string {
	//	timens := int64(time.Now().Nanosecond())
	rand.Seed(int64(time.Now().Nanosecond()))
	r := rand.Intn(len(strs))
	return strs[r]
}

/*
	获得一个随机数(0 - n]，包含0，不包含n
*/
func GetRandNum(n int64) int64 {
	if n <= 0 {
		return 0
	}
	result, _ := crand.Int(crand.Reader, big.NewInt(int64(n)))
	return result.Int64()
}

/*
	随机获取一个域名
*/
func GetRandomDomain() string {
	str := "abcdefghijklmnopqrstuvwxyz"
	// rand.Seed(int64(time.Now().Nanosecond()))
	result := ""
	r := int64(0)
	for i := 0; i < 8; i++ {
		r = GetRandNum(int64(25))
		result = result + str[r:r+1]
	}
	return result
}

/*
	随机获取一个int64类型的随机数
*/
func GetRandomOneInt64() int64 {
	max := BytesToUint64([]byte{255, 255, 255, 255, 255, 255, 255, 255 - 128})
	return GetRandNum(int64(max))
}
