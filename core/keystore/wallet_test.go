/*
SELFKN1RzSSzNQ9KdC3rH255SKoVVtyASPtUk5
SELF9bnX2V2XzjBg7dVZmymM5yT8jpx3r6jxX5
SELFEpRPskfSE1cHvnmBzqbpTAZuAeXP3vaM15
SELFHBWJu2c7yWeKA56voG3fpVMK9T1uxGHfb5
SELFPYuJQuJi23i8X3pKUZTAFV9AUwxiZJdWz5
SELFByi6w6CHHQhLPaTwnc7cNfWWetA5xJcxp5
SELFPwpZbScVPvzddicLNJbazhKXXsHQJGudd5
SELF6yQRLwHDXqr33mVwLgvZtvLCaMas4aEax5
SELFCysC24ZGyvYChU5hecg8zd7FNQmY3J2um5
SELFANtFNxLNT5mLnPG5SxhY5QPvkjoV1ThjR5
SELFJWiDFsRjVZqQhUJH8qUzifmYaRorF73jk5
SELFEe3ZUZAVvESQLtGp8ee5kB2tLQXD1r2Ya5
SELFK6NrivUhhCcS26B7K18aFg8DqDa3KmDkh5
SELFJygoAVauBjZntAMzr6XFbroEpG4oypthS5
SELFHxrmjRueLarmeSwTDyf5QrJGQKpJ27bai5
SELF5BwbcxQiqRbxQQPg5992hyC3iWfV1fhAf5
SELF8poD28RMdZDM6NiVeAxR6siN7w1WbhjCw5
SELFL6LTHog9Ay4UPkWsY4vSVKyXeDbKUdrtk5
SELF39feXYVAVmoMTeF4GBYN1gG2DZJYE1Bpp5
*/
package keystore

import (
	"mandela/core/utils/crypto"
	"fmt"
	"testing"
)

var (
	path = "wallet.txt"
	pass = "1234566"
)

func init() {
	fmt.Println("start....")
	Load(path)
	//CreateKeystore(path, pass)
	fmt.Println(len(GetAddrAll()))
	// for i := 0; i < 100000; i++ {
	// 	addr, _ := GetNewAddr(pass)
	// 	fmt.Println("addr ", i, addr.B58String())
	// }
}
func Benchmark_Addaddr(t *testing.B) {
	for i := 0; i < t.N; i++ {
		addr := crypto.AddressFromB58String("SELF5BwbcxQiqRbxQQPg5992hyC3iWfV1fhAf5")
		FindAddress(addr)
	}
}
