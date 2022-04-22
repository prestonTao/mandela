/*
	预备矿工管理
	管理已经投票成功的预备矿工
*/
package mining

import (
	"mandela/core/utils"
	"bytes"
	"sync"
)

var groupMinersLock = new(sync.RWMutex)
var groupMiners = make([]utils.Multihash, 0)

/*
	备用矿工
*/
type BackupMiners struct {
	Time   int64         //统计时间
	Miners []BackupMiner //预备矿工最多保存两组矿工最大数量(14个)
}

/*
	备用矿工选票计数器
*/
type BackupMiner struct {
	Miner utils.Multihash //矿工地址
	Count uint64          //票数
}

// func (this *BackupMiners) JSON() *[]byte {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil
// 	}
// 	return &bs
// }

/*
	解析预备矿工
*/
// func ParseBackupMiners(bs *[]byte) (*BackupMiners, error) {
// 	bh := new(BackupMiners)
// 	// var jso = jsoniter.ConfigCompatibleWithStandardLibrary
// 	// err := json.Unmarshal(*bs, bh)
// 	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
// 	decoder.UseNumber()
// 	err := decoder.Decode(bh)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return bh, nil
// }

/*
	预备矿工组
*/
//type BackupGroupMiner struct {
//	Height uint64            `json:"height"` //组高度
//	Miners []utils.Multihash `json:"miners"` //预备矿工列表
//}

func AddGroupBackupMiner(miners ...*utils.Multihash) {
	groupMinersLock.Lock()
	//去掉重复的
	for _, one := range miners {
		find := false
		for _, two := range groupMiners {
			if bytes.Equal(*one, two) {
				find = true
				break
			}
		}
		if find {
			continue
		}
		groupMiners = append(groupMiners, *one)
	}
	groupMinersLock.Unlock()
}

/*
	获取预备矿工数量
*/
func TotalBackupMiner() (n uint64) {
	groupMinersLock.RLock()
	n = uint64(len(groupMiners))
	groupMinersLock.RUnlock()
	return
}

//func Get

func RemoveGroupBackupMiner(miners ...utils.Multihash) {
	newMiners := make([]utils.Multihash, 0)
	groupMinersLock.Lock()
	for _, one := range groupMiners {
		find := false
		for _, two := range miners {
			if bytes.Equal(one, two) {
				find = true
				break
			}
		}
		if find {
		} else {
			newMiners = append(newMiners, one)
		}
	}
	groupMiners = newMiners
	groupMinersLock.Unlock()
}
