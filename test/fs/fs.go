package main

import (
	"mandela/core/virtual_node"
	"mandela/store/fs"
	"fmt"
)

func main() {
	// fs.AddSpace("D:/test/hzzfiles/data1")
	fiTable := fs.FileindexSelf{
		Name:   "123",                                            //真实文件名称
		Vid:    virtual_node.AddressNetExtend([]byte("haoayou")), //虚拟节点id
		FileId: virtual_node.AddressNetExtend([]byte("nihao")),   //索引哈市值
		Value:  []byte("buhao"),                                  //内容
	}
	err := fiTable.Add(&fiTable)

	fmt.Println("end", err)
}
