package utils

import (
	"encoding/hex"
	"math/big"
)

/*
	获得域名的hash值
*/
func GetHashForDomain(domain string) []byte {
	return Hash_SHA3_256([]byte(domain))
	// hash := sha1.New()
	// hash.Write([]byte(domain))
	// //	md := hash.Sum(nil)
	// //	return FormatIdUtil(new(big.Int).SetBytes(md))
	// return hash.Sum(nil)
}

//func EncodeToString(bs []byte) string {
//	hex.EncodeToString(bs)
//}

func GetHashByByte(bs []byte) string {
	// hash := sha1.New()
	// hash.Write(bs)
	// md := hash.Sum(nil)
	md := Hash_SHA3_256(bs)
	return FormatIdUtil(new(big.Int).SetBytes(md))
}

/*
	格式化id为十进制或十六进制字符串
*/
func FormatIdUtil(idInt *big.Int) string {
	if idInt.Cmp(big.NewInt(0)) == 0 {
		return "0"
	}
	return hex.EncodeToString(idInt.Bytes())
}
