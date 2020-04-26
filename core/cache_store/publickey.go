/*
	保存网络中各个节点公钥对应的域名
*/
package cache_store

import (
	"mandela/core/config"
	"encoding/hex"
	"sync"
	"time"
)

var (
	OutMulticastPKeyName = make(chan []byte, 100) //需要广播同步的公钥

	keyMapLock = new(sync.RWMutex)
	keyMap     = make(map[string]string) //key=公钥base64，value=域名
)

/*
	添加一个公钥和域名的对应
*/
func AddKeyName(key []byte, name string) {
	keyMapLock.Lock()
	keyMap[hex.EncodeToString(key)] = name
	keyMapLock.Unlock()
	AddSyncMulticastKey(key, time.Now().Unix()+config.Time_key_sync_multicast)
}

/*
	查找一个公钥base64值的
*/
func FindKeyName(key string) (name string, ok bool) {
	keyMapLock.RLock()
	name, ok = keyMap[key]
	keyMapLock.RUnlock()
	return
}
