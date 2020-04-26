package utils

import (
	"math/big"
)

//通过一个域名和用户名得到节点的id
//@return 10进制字符串
func GetHashKey(account string) *big.Int {
	// hash := sha1.New()
	// hash.Write([]byte(account))
	// md := hash.Sum(nil)

	md := Hash_SHA3_256([]byte(account))
	// str16 := hex.EncodeToString(md)
	// resultInt, _ := new(big.Int).SetString(str16, 16)
	resultInt := new(big.Int).SetBytes(md)
	return resultInt
}

func Print(findInt *big.Int) {
	// fmt.Println("==================================\r\n")
	bi := ""

	// findInt := new(big.Int).SetBytes([]byte(nodeId))
	lenght := findInt.BitLen()
	for i := 0; i < lenght; i++ {
		tempInt := findInt
		findInt = new(big.Int).Div(tempInt, big.NewInt(2))
		mod := new(big.Int).Mod(tempInt, big.NewInt(2))
		bi = mod.String() + bi
	}
	// fmt.Println(bi, "\r\n")
	// fmt.Println("==================================\r\n")
}
