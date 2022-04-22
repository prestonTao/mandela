package addr_manager

import (
	"mandela/core/engine"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
	检查一个地址的计算机是否在线
	@return idOnline    是否在线
*/
func CheckOnline(addr string) (isOnline bool) {
	//尝试连接节点，看是否在线
	engine.Log.Info("Try connecting nodes to see if they are online %s", addr)
	conn, err := net.DialTimeout("tcp", addr, time.Second*5)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

/*
	检查地址是否可用
*/
func CheckAddr() {
	/*
		先获得一个拷贝
	*/
	oldSuperPeerEntry := make(map[string]string)
	Sys_superNodeEntry.Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(string)
		oldSuperPeerEntry[key] = value
		return true
	})
	// for key, value := range Sys_superNodeEntry {
	// 	oldSuperPeerEntry[key] = value
	// }
	/*
		一个地址一个地址的判断是否可用
	*/
	for key, _ := range oldSuperPeerEntry {
		// engine.Log.Info("检查节点 1111111111111")
		//如果地址是本节点，则退出
		// if config.Init_LocalIP+":"+strconv.Itoa(int(config.Init_LocalPort)) == key {
		// 	continue
		// }
		if CheckOnline(key) {
			AddSuperPeerAddr(key)
		} else {
			// delete(Sys_superNodeEntry, key)
			Sys_superNodeEntry.Delete(key)
		}
	}
}

/*
	删除一个节点地址
*/
func RemoveIP(ip string, port uint16) {
	key := net.JoinHostPort(ip, strconv.Itoa(int(port)))
	// delete(Sys_superNodeEntry, key)
	Sys_superNodeEntry.Delete(key)
}

/*
	解析超级节点地址列表
*/
func parseSuperPeerEntry(fileBytes []byte) {
	var tempSuperPeerEntry map[string]string
	decoder := json.NewDecoder(bytes.NewBuffer(fileBytes))
	decoder.UseNumber()
	err := decoder.Decode(&tempSuperPeerEntry)
	// if err := json.Unmarshal(fileBytes, &tempSuperPeerEntry); err != nil {
	if err != nil {
		fmt.Println("解析超级节点地址列表失败", err)
		return
	}
	for key, _ := range tempSuperPeerEntry {
		AddSuperPeerAddr(key)
	}
	// AddSuperPeerAddr(Path_SuperPeerdomain)
}
