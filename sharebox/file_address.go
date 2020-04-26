package sharebox

import (
	"mandela/core/utils"
	"mandela/core/utils/base58"
	"os"
)

type FileAddr []byte

func (this *FileAddr) B58String() string {
	return string(base58.Encode(*this))
}

func FileAddressFromB58String(str string) FileAddr {
	return FileAddr(base58.Decode(str))
}

/*
	将文件计算hash值，构建文件地址
*/
func BuildFileAddr(path string) FileAddr {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if fileinfo.IsDir() {
		return nil
	}
	filehash, err := utils.FileSHA3_256(path)
	if err != nil {
		return nil
	}
	return FileAddr(filehash)

	// bs, err := utils.Encode(filehash, config.HashCode)
	// if err != nil {
	// 	return nil
	// }
	// hashName := utils.Multihash(bs)
	// return &hashName
}
