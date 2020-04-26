package name

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"bytes"
	"encoding/json"
)

/*
	在网络中查找域名
*/
func FindNameToNet(name string) *Nameinfo {
	// engine.Log.Info("本次查找域名 %s", name)
	dbKey := append([]byte(config.Name), []byte(name)...)
	// db.Remove(dbKey)
	// engine.Log.Info("查找的域名key %s %s", string(dbKey), hex.EncodeToString(dbKey))
	txbs, err := db.Find(dbKey)
	if err != nil {
		// engine.Log.Info()
		return nil
	}
	// engine.Log.Info("查找到域名")
	nameinfo := new(Nameinfo)
	decoder := json.NewDecoder(bytes.NewBuffer(*txbs))
	decoder.UseNumber()
	err = decoder.Decode(nameinfo)
	// err = json.Unmarshal(*txbs, nameinfo)
	if err != nil {
		// engine.Log.Info("域名解析错误 %s", err.Error())
		return nil
	}
	// engine.Log.Info("查找到域名2")
	return nameinfo

	// //判断域名是否过期，过期了就删除掉
	// if (nameinfo.Height + NameOfValidity) > height {
	// 	return nameinfo
	// }
	// //已经过期
	// db.Remove(dbKey)
	// return nil
}

/*
	在网络中查找域名，从域名中随机选择一个地址返回
*/
func FindNameToNetRandOne(name string, height uint64) *nodeStore.AddressNet {
	nameinfo := FindNameToNet(name)
	if nameinfo == nil {
		return nil
	}
	if nameinfo.CheckIsOvertime(height) {
		return nil
	}
	addr := nameinfo.NetIds[utils.GetRandNum(int64(len(nameinfo.NetIds)))]
	return &addr
}

// /*
// 	解析地址
// */
// func ParseName(txid []byte) *Nameinfo {
// 	txbs, err := db.Find(txid)
// 	if err != nil {
// 		return nil
// 	}
// 	nameinfo := new(Nameinfo)
// 	err = json.Unmarshal(*txbs, nameinfo)
// 	if err != nil {
// 		return nil
// 	}
// 	return nameinfo
// }
