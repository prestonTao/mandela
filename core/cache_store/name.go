package cache_store

import (
	"mandela/core/config"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"bytes"
	"math/rand"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	tempNameLock     = new(sync.RWMutex)
	tempName         = make(map[string]*Name)     //保存刚构建的域名
	OutFlashTempName = make(chan *FlashName, 100) //存放需要确认构建的临时域名

	foreverNameLock = new(sync.RWMutex)
	foreverName     = make(map[string]*Name)     //保存一个永久域名
	OutFlashName    = make(chan *FlashName, 100) //存放需要更新的域名

	OutMulticastName = make(chan []byte, 100) //广播需要同步的域名

	agreeTempName = make(map[string]map[string]int)
	VoteAgree     = utils.NewPollManager() //投赞成票
	VoteDisagree  = utils.NewPollManager() //投反对票

	Root *Name = NewName(config.C_root_name, []*nodeStore.TempId{}, config.Root_publicKeyStr) //根节点域名

	NameSelf = "" //保存自己的域名，"" = 未查询;"." = 还未注册域名
)

/*
	查找一个临时域名
*/
func FindNameInTemp(name string) (ok bool) {
	tempNameLock.RLock()
	_, ok = tempName[name]
	tempNameLock.RUnlock()
	return
}

/*
	查找一个永久域名
*/
func FindNameInForever(name string) (ok bool, one *Name) {
	foreverNameLock.RLock()
	one, ok = foreverName[name]
	foreverNameLock.RUnlock()
	return
}

/*
	添加一个刚构建的域名
*/
func AddBuildName(name string) (ok bool) {
	now := time.Now().Unix()
	//	interval := []int64{now + 60, now + 60*2, now + 60*3, now + 60*4, now + 60*5,
	//		now + 60*6, now + 60*7, now + 60*8, now + 60*9, now + 60*10,
	//		now + 60*20, now + 60*30, now + 60*40, now + 60*50}
	interval := []int64{now + 1, now + 2, now + 3, now + 4, now + 5,
		now + 6, now + 7, now + 8, now + 9, now + 10,
		now + 20, now + 30, now + 40, now + 50}
	tName := Name{
		Name: name,
		//		interval: interval,
	}
	tempNameLock.Lock()
	_, ok = tempName[name]
	if ok {
		ok = false
	} else {
		ok = true
		tempName[name] = &tName
	}
	tempNameLock.Unlock()
	for _, one := range interval {
		addBuildTempName([]byte(name), one)
	}
	addBuildTempNameRemove([]byte(name), now+60)
	return
}

/*
	赞同构建一个临时域名,一共15个，有8个赞同就算通过
*/
func AgreeTempName(name []byte, replyMd5 string) {
	replyMd5 = time.Now().Format("2006-01-02 15:04:05.999999999")
	// fmt.Println("赞成", name, replyMd5)
	if VoteAgree.Vote(config.VOTE_agree_build_name, name, replyMd5) {
		// fmt.Println("投票成功构建域名" + name)
	}
}

/*
	转换一个注册的临时域名为一个永久域名
*/
func SwitchNameForForever(name string) {
	tempNameLock.Lock()
	foreverNameLock.Lock()
	nameOne, ok := tempName[name]
	if ok {
		foreverName[name] = nameOne
	}
	foreverNameLock.Unlock()
	tempNameLock.Unlock()
	//TODO 添加定时更新永久域名事件
}

/*
	root recv
	添加一个永久域名
*/
func AddForeverName(name *Name) {
	name.lock = new(sync.RWMutex)
	//	one := NewName(name, make([][]byte, 0), publicKey)
	foreverNameLock.Lock()
	foreverName[name.Name] = name
	foreverNameLock.Unlock()
	now := time.Now().Unix()
	AddSyncMulticastName([]byte(name.Name), now+config.Time_name_sync_multicast)
}

/*
	获得本机保存的所有域名
*/
func Debug_GetAllName() (names []Name) {
	names = make([]Name, 0)
	foreverNameLock.Lock()
	for _, value := range foreverName {
		names = append(names, *value)
	}
	foreverNameLock.Unlock()

	return
}

type Name struct {
	Name string        `json:"name"` //域名字符串
	lock *sync.RWMutex //
	//	Ids        [][]byte      `json:"ids"`        //所有临时id
	//	UpdateTime int64         `json:"updatetime"` //更新在线时间
	Ids       []*nodeStore.TempId `json:"ids"`       //所有临时id
	PublicKey []byte              `json:"publickey"` //公钥
	Exist     bool                `json:"exist"`     //这个域名是否存在，查找域名的时候返回不存在的域名
}

func (this *Name) JSON() []byte {
	bs, _ := json.Marshal(this)
	return bs
}

/*
	添加一个临时id
*/
func (this *Name) AddId(sid, pid *nodeStore.AddressNet) bool {
	//	sidStr := hex.EncodeToString(sid)
	//	pidStr := hex.EncodeToString(pid)
	this.lock.Lock()
	for _, one := range this.Ids {
		// if one.SuperPeerId.B58String() == sid.B58String() && one.PeerId.B58String() == pid.B58String() {
		if bytes.Equal(*one.SuperPeerId, *sid) && bytes.Equal(*one.PeerId, *pid) {
			this.lock.Unlock()
			return false
		}
		//		if hex.EncodeToString(one) == newId {
		//			this.lock.Unlock()
		//			return false
		//		}
	}
	id := nodeStore.NewTempId(sid, pid)

	this.Ids = append(this.Ids, id)
	this.lock.Unlock()
	return true
}

/*
	删除一个临时id
*/
func (this *Name) delId(sid, pid *utils.Multihash) {
	//	sidStr := hex.EncodeToString(sid)
	//	pidStr := hex.EncodeToString(pid)
	newIds := make([]*nodeStore.TempId, 0)
	this.lock.Lock()
	for i, one := range this.Ids {
		// if one.SuperPeerId.B58String() == sid.B58String() && one.PeerId.B58String() == pid.B58String() {
		if bytes.Equal(*one.SuperPeerId, *sid) && bytes.Equal(*one.PeerId, *pid) {
			copy(newIds, this.Ids[:i])
			newIds = append(newIds, this.Ids[i+1:]...)
			this.Ids = newIds
			break
		}

		//		if idStr == hex.EncodeToString(one) {
		//			copy(newIds, this.Ids[:i])
		//			newIds = append(newIds, this.Ids[i+1:]...)
		//			this.Ids = newIds
		//			break
		//		}
	}
	this.lock.Unlock()
}

/*
	随机获取一个地址
*/
func (this *Name) GetIdOne() *nodeStore.TempId {
	if len(this.Ids) == 0 {
		return nil
	}
	if len(this.Ids) == 1 {
		return this.Ids[0]
	}

	//	timens :=
	rand.Seed(int64(time.Now().Nanosecond()))
	r := rand.Intn(len(this.Ids))
	return this.Ids[r]
}

func NewName(name string, ids []*nodeStore.TempId, publicKey []byte) *Name {
	return &Name{
		Name: name,
		lock: new(sync.RWMutex), //
		Ids:  ids,               //所有临时id
		//		interval:  nil,               //间隔检查时间
		//		UpdateTime: time.Now().Unix(),
		PublicKey: publicKey, //
	}
}

//func ParseName(bs *[]byte) error {
//	nameVO := new(Name)
//	err := json.Unmarshal(bs, &nameVO)
//	if err != nil {
//		fmt.Println(err)
//		return err
//	}
//	for _, one := range nameVO.Ids {

//	}

//}

type FlashName struct {
	Name  string
	Class string
}
