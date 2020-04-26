package crypto

import (
	"mandela/core/utils/base58"
	"bytes"
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

type AddressCoin []byte

func (this *AddressCoin) B58String() string {
	if len(*this) <= 0 {
		return ""
	}
	lastByte := (*this)[len(*this)-1:]
	lastStr := string(base58.Encode(lastByte))
	preLen := int(lastByte[0])
	preStr := string((*this)[:preLen])
	centerStr := string(base58.Encode((*this)[preLen : len(*this)-1]))
	return preStr + centerStr + lastStr

	//	return string(base58.Encode(*this))
}

func AddressFromB58String(str string) AddressCoin {
	if str == "" {
		return nil
	}
	lastStr := str[len(str)-1:]
	lastByte := base58.Decode(lastStr)
	preLen := int(lastByte[0])
	preStr := str[:preLen]
	preByte := []byte(preStr)
	centerByte := base58.Decode(str[preLen : len(str)-1])
	buf := bytes.NewBuffer(preByte)
	buf.Write(centerByte)
	buf.Write(lastByte)
	return AddressCoin(buf.Bytes())

	//	return AddressCoin(base58.Decode(str))
}

/*
	通过公钥生成地址
	@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr(pre string, pubKey []byte) AddressCoin {
	//第一步，计算SHA-256哈希值
	publicSHA256 := sha256.Sum256(pubKey)
	//第二步，计算RIPEMD-160哈希值
	RIPEMD160Hasher := ripemd160.New()
	RIPEMD160Hasher.Write(publicSHA256[:])

	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	//第三步，在上一步结果之间加入地址版本号（如比特币主网版本号“0x00"）
	buf := bytes.NewBuffer([]byte(pre))
	buf.Write(publicRIPEMD160)

	//第四步，计算上一步结果的SHA-256哈希值
	temp := sha256.Sum256(buf.Bytes())
	//第五步，再次计算上一步结果的SHA-256哈希值
	temp = sha256.Sum256(temp[:])
	//第六步，取上一步结果的前4个字节（8位十六进制数）D61967F6，把这4个字节加在第三步结果的后面，作为校验

	buf = bytes.NewBuffer([]byte(pre))
	buf.Write(publicRIPEMD160)
	buf.Write(temp[:4])
	preLen := len([]byte(pre))
	buf.WriteByte(byte(preLen))

	return buf.Bytes()

}

/*
	判断有效地址
	@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func ValidAddr(pre string, addr AddressCoin) bool {
	//判断版本是否正确
	ok := bytes.HasPrefix(addr, []byte(pre))
	if !ok {
		return false
	}
	length := len(addr)
	//
	//	lastBytes := base58.Decode(string(addr[length-1:]))
	//	preLen := int(lastBytes[0])
	preLen := int(addr[length-1])
	preStr := string(addr[:preLen])

	//	fmt.Println(len(pre), pre, len(preStr), preStr)
	if pre != preStr {
		return false
	}
	//
	temp := sha256.Sum256(addr[:length-4-1])
	temp = sha256.Sum256(temp[:])

	//	fmt.Println(addr[:len(addr)-1], temp[:4])

	ok = bytes.HasSuffix(addr[:len(addr)-1], temp[:4])
	if !ok {
		//		fmt.Println("false false")
		return false
	}
	return true

	//
	//	temp := sha256.Sum256(addr[:length-4])
	//	temp = sha256.Sum256(temp[:])
	//	ok = bytes.HasSuffix(addr, temp[:4])
	//	if !ok {
	//		return false
	//	}
	//	return true
}

/*
	检查公钥生成的地址是否一样
	@return    bool    是否一样 true=相同;false=不相同;
*/
func CheckPukAddr(pre string, pubKey []byte, addr AddressCoin) bool {
	tagAddr := BuildAddr(pre, pubKey)
	return bytes.Equal(tagAddr, addr)
}
