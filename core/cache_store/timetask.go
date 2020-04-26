package cache_store

import (
	"encoding/hex"
	// "fmt"
	"mandela/core/config"
	"mandela/core/utils"
)

var task = utils.NewTask(prosess)

/*
	定时查询临时域名是否注册成功
*/
func addBuildTempName(name []byte, endTime int64) {
	task.Add(endTime, config.TSK_build_temp_name, name)
}

/*
	定时删除没注册成功的临时域名
*/
func addBuildTempNameRemove(name []byte, endTime int64) {
	task.Add(endTime, config.TSK_build_temp_name_remove, name)
}

/*
	定时广播需要同步的域名
*/
func AddSyncMulticastName(name []byte, endTime int64) {
	task.Add(endTime, config.TSK_name_sync_multicast, name)
}

/*
	定时广播需要同步的公钥
*/
func AddSyncMulticastKey(key []byte, endTime int64) {
	task.Add(endTime, config.TSK_key_sync_multicast, key)
}

func prosess(class string, params []byte) {
	switch class {
	case config.TSK_build_temp_name_remove: //删除没注册成功的临时域名
		// fmt.Println("开始删除临时域名", tempName)
		tempNameLock.Lock()
		delete(tempName, hex.EncodeToString(params))
		tempNameLock.Unlock()
		// fmt.Println("删除了这个临时域名", params, tempName)

	case config.TSK_build_temp_name: //定时查询临时域名是否注册成功
		//剩下是需要更新的域名
		flashName := FlashName{
			Name:  hex.EncodeToString(params),
			Class: class,
		}
		OutFlashTempName <- &flashName
	case config.TSK_name_sync_multicast: //定时广播需要同步的域名
		OutMulticastName <- params
	case config.TSK_key_sync_multicast: //定时广播需要同步的域名
		OutMulticastPKeyName <- params
	default:
		//剩下是需要更新的域名
		flashName := FlashName{
			Name:  hex.EncodeToString(params),
			Class: class,
		}
		OutFlashName <- &flashName
	}

}
