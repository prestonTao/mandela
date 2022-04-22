package fs

import (
	"mandela/config"
	"mandela/core/virtual_node"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

type FsManager struct {
	lock     *sync.RWMutex //
	storeges []*Storage    //保存所有存储空间
}

/*
	获取存储空间信息列表
*/
func (this *FsManager) GetSpaceList() []StorageVO {
	storeages := make([]StorageVO, 0)
	this.lock.RLock()
	for i := 0; i < len(this.storeges); i++ {
		svo := ConvertStorageVO(this.storeges[i])
		storeages = append(storeages, svo)
	}
	this.lock.RUnlock()
	return storeages
}

/*
	获取总空间大小
*/
func (this *FsManager) GetSpaceNum() (size uint64) {
	this.lock.RLock()
	n := len(this.storeges)
	this.lock.RUnlock()
	size = uint64(n) * config.Spacenum
	return
}

/*
	获取已经使用空间大小
*/
func (this *FsManager) GetUseSpaceSize() (size uint64) {
	this.lock.RLock()
	for _, one := range this.storeges {
		size = size + one.UseSize
	}
	this.lock.RUnlock()
	return
}

/*
	获取能装下指定大小文件的空间
*/
func (this *FsManager) GetNotUseSpace(size uint64) (vnodeinfo *virtual_node.Vnodeinfo) {
	//	fmt.Println("1111111111111")
	this.lock.RLock()
	//	fmt.Println("22222222222")
	for _, one := range this.storeges {
		//		fmt.Println("333333333", one.SpaceNum, one.UseSize, size)
		if one.SpaceNum >= one.UseSize+size {
			//			fmt.Println("4444444444444", one.VnodeId.B58String())
			vnodeinfo = virtual_node.FindInVnodeinfoSelf(one.VnodeId)

			break
		}
	}
	this.lock.RUnlock()
	return
}

/*
	加载空间
*/
func (this *FsManager) LoadAllSpace() error {
	ssc := StoreSpaceConfig{}
	configs, err := ssc.GetAll()
	if err != nil {
		return err
	}
	this.lock.Lock()

	tags := make(chan bool, runtime.NumCPU())
	for _, one := range configs {
		tags <- false
		vnode := virtual_node.AddVnode()

		newStore, err := NewStorage(vnode.Vid, one.DBAbsPath, one.Index)
		if err != nil {
			//删除不可用空间
			ssc := StoreSpaceConfig{}
			ssc.Del(one.VNodeId)
			os.Remove(one.DBAbsPath)
			<-tags
			continue
		}
		go newStore.FullTable(tags)

		this.storeges = append(this.storeges, newStore)
	}

	this.lock.Unlock()
	return nil
}

/*
	保存文件块
*/
func (this *FsManager) SaveFileChunk(vid virtual_node.AddressNetExtend,
	key virtual_node.AddressNetExtend, value *[]byte) {
	this.lock.RLock()
	for _, one := range this.storeges {
		if bytes.Equal(one.VnodeId, vid) {
			tnum, id := one.FindNotUseChunk()
			fmt.Println("查询到的id", id)
			one.Save(tnum, id, key, value)
			break
		}
	}
	this.lock.RUnlock()
}

/*
	添加空间
*/
func (this *FsManager) AddSpace(absPath string, n uint64) (err error) {
	if absPath == "" {
		absPath = config.Store_path_dir
	}
	this.lock.Lock()

	maxIndex := uint64(0)
	//查找数据库名称索引最大值
	for i := 0; i < len(this.storeges); i++ {
		if this.storeges[i].Index > maxIndex {
			maxIndex = this.storeges[i].Index
		}
	}

	tags := make(chan bool, runtime.NumCPU())
	// count := len(this.storeges)
	for i := uint64(0); i < n; i++ {
		tags <- false

		maxIndex++
		// count++
		vnode := virtual_node.AddVnode()

		//		virtual_node.SetupVnodeNumber(uint64(count))
		fileName := StoreSqlNamePre + strconv.Itoa(int(maxIndex)) + StoreSqlNameTail
		filePath := filepath.Join(absPath, fileName)
		var newStore *Storage
		newStore, err = NewStorage(vnode.Vid, filePath, maxIndex)
		if err != nil {
			break
		}
		go newStore.FullTable(tags)

		this.storeges = append(this.storeges, newStore)
		ssc := StoreSpaceConfig{}
		ssc.Add(vnode.Index, vnode.Vid, filePath, maxIndex)

		//		space := NewSpace(absPath)
		//		if space == nil {
		//			return
		//		}
		//		space.Init()
		// <-tags
	}
	this.lock.Unlock()
	return
}

/*
	删除空间
*/
func (this *FsManager) DelSpace(n uint64) {
	// fmt.Println("DelSpace")

	this.lock.Lock()
	for i := uint64(0); i < n; i++ {
		vnode := virtual_node.DelVnode()
		// vid := vnode.Vid.B58String()
		ssc := StoreSpaceConfig{}
		ssc.Del(vnode.Vid)
		for j, one := range this.storeges {
			// fmt.Println("DelSpace 1", one.VnodeId)
			if bytes.Equal(one.VnodeId, vnode.Vid) {
				// fmt.Println("DelSpace 2", vid)
				temp := this.storeges[:j]
				this.storeges = append(temp, this.storeges[j+1:]...)
				// fmt.Println("DelSpace 3", vid)
				//暂停数据库
				one.StopAndDel()
				//删除数据库
				one.sqldb.Close()
				os.Remove(one.DbPath)
				// fmt.Println("DelSpace 4", vid)
				break
			}
		}
	}
	this.lock.Unlock()
}

/*
	删除一个指定的空间
*/
func (this *FsManager) DelSpaceOne(dbpath string) {
	fmt.Println("DelSpaceOne")

	this.lock.Lock()
	//找到要删除的空间
	findIndex := -1
	for i, one := range this.storeges {
		if one.DbPath == dbpath {
			findIndex = i
		}
	}
	if findIndex < 0 {
		//未找到要删除的空间
		return
	}

	//找到要删除的空间后，转移要删除空间中的文件

	// MoveFileChunk()

	// for i := uint64(0); i < n; i++ {
	// 	vnode := virtual_node.DelVnode()
	// 	vid := vnode.Vid.B58String()
	// 	ssc := StoreSpaceConfig{}
	// 	ssc.Del(vnode.Vid)
	// 	for j, one := range this.storeges {
	// 		fmt.Println("DelSpace 1", one.VnodeId)
	// 		if bytes.Equal(one.VnodeId, vnode.Vid) {
	// 			fmt.Println("DelSpace 2", vid)
	// 			temp := this.storeges[:j]
	// 			this.storeges = append(temp, this.storeges[j+1:]...)
	// 			fmt.Println("DelSpace 3", vid)
	// 			//暂停数据库
	// 			one.StopAndDel()
	// 			//删除数据库
	// 			one.sqldb.Close()
	// 			os.Remove(one.DbPath)
	// 			fmt.Println("DelSpace 4", vid)
	// 			break
	// 		}
	// 	}
	// }
	this.lock.Unlock()
}

func NewFsManager() *FsManager {
	return &FsManager{
		lock:     new(sync.RWMutex),
		storeges: make([]*Storage, 0),
	}
}

/*
	移动数据库中的文件片
*/
func MoveFileChunk(src, dsc *Storage) {
	// src
}
