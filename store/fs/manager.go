package fs

import (
	"mandela/config"
	"mandela/core/virtual_node"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type FsManager struct {
	lock     *sync.RWMutex //
	storeges []*Storage    //保存所有存储空间
}

/*
	获取总空间大小
*/
func (this *FsManager) GetSpaceNum() (size uint64) {
	this.lock.Lock()
	n := len(this.storeges)
	this.lock.Unlock()
	size = uint64(n) * config.Spacenum
	return
}

/*
	获取已经使用空间大小
*/
func (this *FsManager) GetUseSpaceSize() (size uint64) {
	this.lock.Lock()
	for _, one := range this.storeges {
		size = size + one.UseSize
	}
	this.lock.Unlock()
	return
}

/*
	获取能装下指定大小文件的空间
*/
func (this *FsManager) GetNotUseSpace(size uint64) (vnodeinfo *virtual_node.Vnodeinfo) {
	//	fmt.Println("1111111111111")
	this.lock.Lock()
	//	fmt.Println("22222222222")
	for _, one := range this.storeges {
		//		fmt.Println("333333333", one.SpaceNum, one.UseSize, size)
		if one.SpaceNum >= one.UseSize+size {
			//			fmt.Println("4444444444444", one.VnodeId.B58String())
			vnodeinfo = virtual_node.FindInVnodeinfoSelf(one.VnodeId)

			break
		}
	}
	this.lock.Unlock()
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

	for _, one := range configs {

		vnode := virtual_node.AddVnode()

		newStore, err := NewStorage(vnode.Vid, one.DBAbsPath)
		if err != nil {
			//删除不可用空间
			ssc := StoreSpaceConfig{}
			ssc.Del(one.VNodeId)
			os.Remove(one.DBAbsPath)
			continue
		}
		go newStore.FullTable()

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
	this.lock.Lock()
	for _, one := range this.storeges {
		if bytes.Equal(one.VnodeId, vid) {
			tnum, id := one.FindNotUseChunk()
			fmt.Println("查询到的id", id)
			one.Save(tnum, id, key, value)
			break
		}
	}
	this.lock.Unlock()
}

/*
	添加空间
*/
func (this *FsManager) AddSpace(absPath string, n uint64) (err error) {
	if absPath == "" {
		absPath = config.Store_path_dir
	}
	this.lock.Lock()
	count := len(this.storeges)
	for i := uint64(0); i < n; i++ {

		count++
		vnode := virtual_node.AddVnode()

		//		virtual_node.SetupVnodeNumber(uint64(count))
		fileName := StoreSqlNamePre + strconv.Itoa(count) + StoreSqlNameTail
		filePath := filepath.Join(absPath, fileName)
		var newStore *Storage
		newStore, err = NewStorage(vnode.Vid, filePath)
		if err != nil {
			break
		}
		go newStore.FullTable()

		this.storeges = append(this.storeges, newStore)
		ssc := StoreSpaceConfig{}
		ssc.Add(vnode.Index, vnode.Vid, filePath)

		//		space := NewSpace(absPath)
		//		if space == nil {
		//			return
		//		}
		//		space.Init()
	}
	this.lock.Unlock()
	return
}

/*
	删除空间
*/
func (this *FsManager) DelSpace(n uint64) {
	fmt.Println("DelSpace")

	this.lock.Lock()
	for i := uint64(0); i < n; i++ {
		vnode := virtual_node.DelVnode()
		vid := vnode.Vid.B58String()
		ssc := StoreSpaceConfig{}
		ssc.Del(vnode.Vid)

		for j, one := range this.storeges {
			fmt.Println("DelSpace 1", one.VnodeId)
			if bytes.Equal(one.VnodeId, vnode.Vid) {
				fmt.Println("DelSpace 2", vid)
				temp := this.storeges[:j]
				this.storeges = append(temp, this.storeges[j+1:]...)
				fmt.Println("DelSpace 3", vid)
				//暂停数据库
				one.StopAndDel()
				//删除数据库
				one.sqldb.Close()
				os.Remove(one.DbPath)
				fmt.Println("DelSpace 4", vid)
				break
			}
		}

	}
	this.lock.Unlock()

}

func NewFsManager() *FsManager {
	return &FsManager{
		lock:     new(sync.RWMutex),
		storeges: make([]*Storage, 0),
	}
}
