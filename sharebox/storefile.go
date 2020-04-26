//调用存储模块的本地索引
package sharebox

import (
	"mandela/config"
	"mandela/core/nodeStore"
	sconfig "mandela/sharebox/config"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type FileInfo struct {
	Hash *nodeStore.AddressNet `json:"Hash"` //文件hash
	Name string                `json:"Name"` //真实文件名称
	Size uint64                `json:"Size"` //文件总大小
	Time int64                 `json:"Time"` //文件上传时间
}

func parseFileInfo(d []byte) (*FileInfo, error) {
	fi := new(FileInfo)
	// err := json.Unmarshal(d, &fi)
	decoder := json.NewDecoder(bytes.NewBuffer(d))
	decoder.UseNumber()
	err := decoder.Decode(&fi)
	if err != nil {
		return nil, err
	}
	return fi, nil
}

//列出本地索引文件
func getListFileFromSelf() *DirVO {
	//增加本地索引文件
	selffile := config.Store_fileinfo_self
	dir := NewDir(selffile)
	listFileFromSelf(selffile, dir)
	return dir.conversionVO()
}

//列出目录下的所有文件
func listFileFromSelf(myfolder string, dir *Dir) {
	fileInfos, _ := ioutil.ReadDir(myfolder)
	for _, fileinfo := range fileInfos {
		filePath := filepath.Join(myfolder, fileinfo.Name())
		if fileinfo.IsDir() {
			newDir := NewDir(filePath)
			dir.AddDir(newDir)
			listFile(filePath, newDir)
		} else {
			fi := ParseFile(filePath)
			hashName := FileAddressFromB58String(fi.Hash.B58String())
			if hashName == nil {
				continue
			}
			if path.Ext(fi.Name) == sconfig.Suffix {
				continue
			}
			newFile := NewFile(filePath, hashName)
			newFile.Name = fi.Name
			dir.AddFile(newFile)
		}
	}
}

//解析索引文件内容
func ParseFile(name string) (fi *FileInfo) {
	f, err := os.Open(name)
	if err != nil {
		return
	}
	res, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	fi, err = parseFileInfo(res)
	if err != nil {
		return
	}
	return
}
