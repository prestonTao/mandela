package sharebox

import (
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var onece sync.Once
var watcher *Batcher

func InitWatcher() (err error) {
	onece.Do(func() {
		watcher, err = NewBatcher(time.Second)
		if err != nil {
			return
		}
		go func() {
			for {
				select {
				case events, ok := <-watcher.Events:
					if !ok {
						continue
					}
					for _, event := range events {
						//fmt.Printf("events: %+v", event)
						switch event.Op {
						case fsnotify.Create: //创建文件
							err1 := AddFold(event.Name)
							if err1 != nil {
								fmt.Println(err1)
							}
							//判断是否文件夹
							err := AddShareFile(event.Name)
							if err != nil {
								continue
							}
						case fsnotify.Write:
							err := AddFile(event.Name)
							if err != nil {
								fmt.Println(err)
							}
						case fsnotify.Remove:
							err := DelFold(event.Name)
							if err != nil {
								fmt.Println(err)
							}
							err = DelFile(event.Name)
							if err != nil {
								fmt.Println(err)
							}
						case fsnotify.Rename:
							err := DelFile(event.Name)
							if err != nil {
								fmt.Println(err)
							}
							err = DelFold(event.Name)
							if err != nil {
								fmt.Println(err)
							}
						case fsnotify.Chmod:
						}
						// if event.Op&fsnotify.Write == fsnotify.Write {
						// 	log.Println("modified file:", event.Name)
						// }
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}()
	})
	return
}

/*
	给文件夹添加监听
*/
func WatcherFolder(folder string) error {
	// fmt.Println("添加一个目录", folder)
	err := watcher.Add(folder)
	return err
}

/*
	删除对一个文件夹的监听
*/
func DelWatcherFolder(folder string) error {
	return watcher.Remove(folder)
}

/*
	添加一个共享文件，生成文件hash和索引
*/
func AddShareFile(path string) error {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileinfo.IsDir() {
		return nil
	}
	filehash, err := utils.FileSHA3_256(path)
	if err != nil {
		return err
	}
	// bs, err := utils.Encode(filehash, config.HashCode)
	// if err != nil {
	// 	return err
	// }
	// hashName := utils.Multihash(bs)
	hashName := nodeStore.AddressNet(filehash)
	fmt.Println("生成文件hash", hashName.B58String())

	_, fileName := filepath.Split(path)
	// fmt.Println(hashName, "fileName", fileName)

	fi := NewFileIndex(&hashName, fileName, uint64(fileinfo.Size()))
	err = UpNetFileindex(fi)
	if err != nil {
		return err
	}
	return nil
}

/*
	计算文件的hash值，封装成Multihash
*/
// func GetFileHash(path string) *utils.Multihash {
// 	fileinfo, err := os.Stat(path)
// 	if err != nil {
// 		return nil
// 	}
// 	if fileinfo.IsDir() {
// 		return nil
// 	}
// 	filehash, err := utils.FileSHA3_256(path)
// 	if err != nil {
// 		return nil
// 	}
// 	bs, err := utils.Encode(filehash, config.HashCode)
// 	if err != nil {
// 		return nil
// 	}
// 	hashName := utils.Multihash(bs)
// 	return &hashName
// }
