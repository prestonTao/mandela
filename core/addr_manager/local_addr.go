package addr_manager

import (
	"mandela/config"
	"fmt"

	// "fmt"
	"io/ioutil"
)

func init() {
	registerFunc(LoadSuperPeerEntry)
}

/*
	读取并解析本地的超级节点列表文件
*/
func LoadSuperPeerEntry() {
	if len(config.Entry) > 0 {
		for _, value := range config.Entry {
			AddSuperPeerAddr(value)
		}
		// AddSuperPeerAddr(Path_SuperPeerdomain)
	} else {

		fileBytes, err := ioutil.ReadFile(Path_SuperPeerAddress)
		if err != nil {
			fmt.Println("读取超级节点地址列表失败", err)
			return
		}
		parseSuperPeerEntry(fileBytes)
	}
}
