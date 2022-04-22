package im

import (
	"bytes"
	"fmt"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

//消息文件
type FileInfo struct {
	Name  string //原始文件名
	Size  int64
	Path  string
	Index int64
	Data  []byte
}

func (fi *FileInfo) Json() []byte {
	d, err := json.Marshal(fi)
	if err != nil {
		fmt.Println(err)
	}
	return d
}
func ParseFileInfo(d []byte) *FileInfo {
	fi := FileInfo{}
	// err := json.Unmarshal(d, &fi)
	decoder := json.NewDecoder(bytes.NewBuffer(d))
	decoder.UseNumber()
	err := decoder.Decode(&fi)
	if err != nil {
		fmt.Println(err)
	}
	return &fi
}
