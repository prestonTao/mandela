package cloud_space

import (
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/virtual_node"
	"bytes"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type FileIndex struct {
	// Hash       *virtual_node.AddressNetExtend //文件hash
	Hash       *virtual_node.AddressNetExtend  //文件hash
	Name       string                          //真实文件名称
	Size       uint64                          //文件总大小
	Time       int64                           //文件上传时间
	FileChunk  []*FileChunk                    //文件块以及块共享者名单，用list保证文件块的顺序
	FileOwner  map[string]nodeStore.AddressNet //文件所有者 key 用户nid value FileOwner
	CryptUser  *nodeStore.AddressNet           //加密码者节点ID
	ChunkCount uint64                          //文件块总数
	//	Users *sync.Map        //共享的用户列表 key:string,value:*ShareUser
}

// type FileIndexTempVO struct {
// 	Hash       *virtual_node.AddressNetExtend //文件md5
// 	Name       string                         //真实文件名称
// 	Size       uint64                         //文件总大小
// 	Time       int64                          //文件上传时间
// 	FileChunk  []*FileChunkTempVO             //文件块以及块共享者名单
// 	FileOwner  map[string]FileOwner           //文件所有者 key 用户hash value FileUser
// 	CryptUser  *nodeStore.AddressNet          //加密码者节点ID
// 	ChunkCount uint64                         //文件块总数
// }
// type FileChunkTempVO struct {
// 	No    uint64                         //文件块编号，从0开始递增
// 	Hash  *virtual_node.AddressNetExtend //块hash值
// 	Size  uint64                         //块大小
// 	Users []*ShareUser                   //共享的用户列表 key:string,value:*ShareUser
// }

func (this *FileIndex) JSON() []byte {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil
	}
	return bs
	// fitvo := FileIndexTempVO{
	// 	Hash:       this.Hash,                   //文件md5
	// 	Name:       this.Name,                   //真实文件名称
	// 	Size:       this.Size,                   //文件大小
	// 	Time:       this.Time,                   //文件上传时间
	// 	FileChunk:  make([]*FileChunkTempVO, 0), //文件块以及块共享者名单
	// 	FileOwner:  make(map[string]FileOwner),  //文件所有者 key 用户hash value FileOwner
	// 	CryptUser:  this.CryptUser,              //加密码者节点ID
	// 	ChunkCount: this.ChunkCount,             //文件块总数
	// }
	// for _, v := range this.FileChunk.GetAll() {
	// 	one := v.(*FileChunk)
	// 	fctVO := FileChunkTempVO{
	// 		No:    one.No,                //文件块编号，从0开始递增
	// 		Hash:  one.Hash,              //块hash值
	// 		Size:  one.Size,              //块大小
	// 		Users: make([]*ShareUser, 0), //共享的用户列表 key:string,value:*ShareUser
	// 	}

	// 	one.Users.Range(func(key interface{}, valueItr interface{}) bool {
	// 		value := valueItr.(*ShareUser)
	// 		fctVO.Users = append(fctVO.Users, value)
	// 		return true
	// 	})
	// 	fitvo.FileChunk = append(fitvo.FileChunk, &fctVO)
	// }
	// fitvo.FileOwner = this.FileOwner
	// bs, err := json.Marshal(fitvo)
	// if err != nil {
	// 	return nil
	// }
	// m := make(map[string]interface{})
	// // err = json.Unmarshal(bs, &m)
	// decoder := json.NewDecoder(bytes.NewBuffer(bs))
	// decoder.UseNumber()
	// err = decoder.Decode(&m)
	// if err != nil {
	// 	return nil
	// }
	// return bs
}

/*
	添加一个共享用户
	@return  bool  是否添加成功，已经存在也是添加成功
*/
func (this *FileIndex) AddShareUser(no uint64, user *ShareUser) bool {
	for _, one := range this.FileChunk {
		if one.No == no {
			one.AddUpdateShareUser(user)
			return true
		}
	}
	return false
	// for i, v := range this.FileChunk.GetAll() {
	// 	one := v.(*FileChunk)
	// 	if one.No == no {
	// 		itr := this.FileChunk.Get(i)
	// 		two := itr.(*FileChunk)
	// 		two.AddUpdateShareUser(user)
	// 		return true
	// 	}
	// }
	// return false

}

/*
	添加一个文件块
*/
func (this *FileIndex) AddChunk(chunk *FileChunk) {

	if chunk.No >= this.ChunkCount { //块编号从0开始
		return
	}
	if uint64(len(this.FileChunk)) == chunk.No {
		this.FileChunk = append(this.FileChunk, chunk)
	}
	return

	// have := false
	// for _, v := range this.FileChunk.GetAll() {
	// 	one := v.(*FileChunk)
	// 	if one.No == chunk.No {
	// 		have = true
	// 		break
	// 	}
	// }
	// if !have {
	// 	this.FileChunk.Add(chunk)
	// }

	//	this.lock.Lock()
	//	have := false
	//	//检查块编号是否存在
	//	for _, one := range this.FileChunk {
	//		if one.No == chunk.No {
	//			have = true
	//			break
	//		}
	//	}
	//	if !have {
	//		this.FileChunk = append(this.FileChunk, chunk)
	//	}
	//	this.lock.Unlock()
}

/*
	查找本地是否有文件块
*/
func (this *FileIndex) FindChunk(hash virtual_node.AddressNetExtend) (have bool) {
	have = false
	for _, one := range this.FileChunk {
		if bytes.Equal(*one.Hash, hash) {
			have = true
			break
		}
	}
	return
	// for _, v := range this.FileChunk.GetAll() {
	// 	one := v.(*FileChunk)
	// 	if bytes.Equal(*one.Hash, hash) {
	// 		have = true
	// 		break
	// 	}
	// }
	// return

}

// /*
// 	查找本地是否有文件块
// */
// func (this *FileIndex) FindChunk(hash string) (have *FileChunk) {
// 	for _, v := range this.FileChunk.GetAll() {
// 		one := v.(*FileChunk)
// 		if one.Hash.B58String() == hash {
// 			have = one
// 			break
// 		}
// 	}
// 	return

// 	//	this.lock.RLock()
// 	//	for _, one := range this.FileChunk {
// 	//		if one.Hash.B58String() == hash {
// 	//			have = one
// 	//			break
// 	//		}
// 	//	}
// 	//	this.lock.RUnlock()
// 	//	return
// }

//增加文件所有者
/*

 */
func (this *FileIndex) AddFileOwner(vnode nodeStore.AddressNet) error {
	// user := FileOwner{Vnodeinfo: vnode, UpdateTime: time.Now().Unix()}
	this.FileOwner[vnode.B58String()] = vnode
	return nil
}

//删除超时用户(当文件所有者为空时，则停止文件更新并删除)
// func (this *FileIndex) DelFileUser() error {
// 	for k, v := range this.FileOwner {
// 		if time.Now().Unix()-v.UpdateTime > Time_fileuser {
// 			delete(this.FileOwner, k)
// 		}
// 	}
// 	return nil
// }

//合并所有者
// func (fia *FileIndex) MergeFileUser(fib *FileIndex) {
// 	ua := fia.FileOwner
// 	ub := fib.FileOwner
// 	if ua == nil || ub == nil {
// 		return
// 	}
// 	for k, v := range ub {
// 		_, ok := ua[k]
// 		//如果存在，则更新在线时间,不存在则直接合并
// 		if ok {
// 			if v.UpdateTime > ua[k].UpdateTime {
// 				ua[k] = v
// 			}
// 		} else {
// 			ua[k] = v
// 		}
// 	}
// 	//fmt.Printf("###合并前###ua:%+v ub:%+v", ua, ub)
// 	//fmt.Printf("###合并###%+v\n", ua)
// 	fia.FileOwner = ua
// 	fia.DelFileUser()
// 	return
// }
func ParseFileindex(bs []byte) (*FileIndex, error) {
	fi := new(FileIndex)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(fi)
	if err != nil {
		return nil, err
	}
	return fi, nil

	// fitVO := new(FileIndexTempVO)
	// // err := json.Unmarshal(bs, fitVO)
	// decoder := json.NewDecoder(bytes.NewBuffer(bs))
	// decoder.UseNumber()
	// err := decoder.Decode(fitVO)
	// if err != nil {
	// 	return nil, err
	// }

	// fi := new(FileIndex)
	// fi.FileChunk = utils.NewSyncList()
	// fi.ChunkCount = fitVO.ChunkCount
	// fi.Hash = fitVO.Hash
	// fi.Name = fitVO.Name
	// fi.Size = fitVO.Size
	// fi.Time = fitVO.Time
	// fi.CryptUser = fitVO.CryptUser
	// for _, one := range fitVO.FileChunk {
	// 	fc := new(FileChunk)
	// 	fc.Hash = one.Hash
	// 	fc.No = one.No
	// 	fc.Size = one.Size
	// 	fc.Users = new(sync.Map)
	// 	for _, u := range one.Users {
	// 		fc.Users.Store(u.Nid.B58String(), NewShareUser(u.Vnodeinfo))
	// 	}
	// 	fi.FileChunk.Add(fc)
	// 	//		 = append(fi.FileChunk, fc)
	// }
	// fi.FileOwner = fitVO.FileOwner
	// return fi, nil
}

/*
	创建一个文件信息
*/
func NewFileIndex(hash *virtual_node.AddressNetExtend, filename string, chunkCount uint64) *FileIndex {
	return &FileIndex{
		Hash:       hash,
		Name:       filename,
		FileChunk:  make([]*FileChunk, 0), // utils.NewSyncList(),
		FileOwner:  make(map[string]nodeStore.AddressNet),
		ChunkCount: chunkCount,
	}
}

type FileChunk struct {
	No         uint64                         //文件块编号，从0开始递增
	Size       uint64                         //块大小
	Hash       *virtual_node.AddressNetExtend //块hash值
	ShareUsers map[string]*ShareUser          //共享的用户列表 key:string,value:*ShareUser
}

/*
	添加用户，用户已经存在则更新
*/
func (this *FileChunk) AddUpdateShareUser(user *ShareUser) {

	value, ok := this.ShareUsers[user.Nid.B58String()]
	if ok {
		// u := value.(*ShareUser)
		atomic.StoreInt64(&value.UpdateTime, time.Now().Unix())
		return
	}
	u := NewShareUser(user.Vnodeinfo)
	this.ShareUsers[user.Nid.B58String()] = u
	// this.ShareUsers.Store(user.Nid.B58String(), u)
}

/*
	获取10分钟内在线的用户
*/
func (this *FileChunk) GetUserOnline() []*ShareUser {
	us := make([]*ShareUser, 0)
	for _, value := range this.ShareUsers {
		if !value.CheckOvertime(Time_sharefile * 2) {
			us = append(us, value)
		}
	}
	return us

	// this.ShareUsers.Range(func(key interface{}, valueItr interface{}) bool {
	// 	value := valueItr.(*ShareUser)
	// 	if !value.CheckOvertime(Time_sharefile * 2) {
	// 		us = append(us, value)
	// 	}
	// 	return true
	// })
	// return us
}

/*
	获取所有用户
*/
func (this *FileChunk) GetShareUserAll() []*ShareUser {
	us := make([]*ShareUser, 0)
	for i, _ := range this.ShareUsers {
		us = append(us, this.ShareUsers[i])
	}
	return us

	// this.ShareUsers.Range(func(key interface{}, valueItr interface{}) bool {
	// 	value := valueItr.(*ShareUser)
	// 	us = append(us, value)
	// 	return true
	// })
	// return us
}

//随机获取一个共享用户
func (this *FileChunk) RandUser() *ShareUser {
	us := this.GetUserOnline()
	if len(us) <= 0 {
		us = this.GetShareUserAll()
	}
	users := make([]*ShareUser, 0)
	for _, one := range us {
		users = append(users, one)
	}
	if len(users) <= 0 {
		return nil
	}
	r := utils.GetRandNum(int64(len(users)))

	//	rand.Seed(int64(time.Now().Nanosecond()))
	//	r := rand.Intn(len(names))
	return users[r]
}

/*
	清理60天都不在线的用户
*/
func (this *FileChunk) Clear() {
	us := this.GetShareUserAll()
	for _, one := range us {
		if one.CheckOvertime(Time_shareUserOfflineClear) {
			// fmt.Println("清理掉用户", one.Name.B58String(), time.Now().Unix()-one.UpdateTime)
			// this.ShareUsers.Delete(one.Nid.B58String())
			delete(this.ShareUsers, one.Nid.B58String())
		}
	}
}

func NewFileChunk(no uint64, hash *virtual_node.AddressNetExtend) *FileChunk {
	return &FileChunk{
		No:         no,
		Hash:       hash,
		ShareUsers: make(map[string]*ShareUser),
	}
}

/*
	文件块共享用户
*/
type ShareUser struct {
	virtual_node.Vnodeinfo
	//	Name       *nodeStore.AddressNet //用户名称
	UpdateTime int64 //最后在线时间，一个用户3个月不在线，则从块中删除
}

/*
	检查是否超时
*/
func (this *ShareUser) CheckOvertime(t int64) bool {
	if this.UpdateTime+t <= time.Now().Unix() {
		return true
	}
	return false
}

func NewShareUser(user virtual_node.Vnodeinfo) *ShareUser {
	return &ShareUser{
		Vnodeinfo:  user,
		UpdateTime: time.Now().Unix(),
	}
}
