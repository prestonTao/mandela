package cachedata

import (
	"encoding/json"
	// "fmt"
	"mandela/core/utils"
	"time"
)

//数据结构
type CacheData struct {
	Kbyte []byte             //原始KEY
	Key   *utils.Multihash   //数据key hash
	Value []byte             //数据内容
	Ownid []*utils.Multihash //所属用户ID
	Time  time.Time          //更新时间,用于清理过期数据
	Del   bool               //删除用
}

//根据key value生成要存储的数据结构
func newCacheData(key, value []byte) *CacheData {
	khash := buildHash(key)
	data := CacheData{Kbyte: key, Key: khash, Value: value, Time: time.Now(), Del: false}
	return &data
}

//增加节点为共享用户
func (cd *CacheData) AddOwnId(id *utils.Multihash) error {
	if !checkOwn(id, cd.Ownid) {
		cd.Ownid = append(cd.Ownid, id)
	}
	return nil
}
func (cd *CacheData) SetTime() error {
	cd.Time = time.Now()
	return nil
}
func (cd *CacheData) Json() []byte {
	res, err := json.Marshal(cd)
	if err != nil {
		// fmt.Println("CacheData:", err)
	}
	return res
}

//生成需要保存的K/V
func (cd *CacheData) buildKV() (key string, value []byte) {
	key = cd.Key.B58String()
	value = cd.Json()
	return
}

//解析字节为cachedata
func Parse(data []byte) (*CacheData, error) {
	cd := new(CacheData)
	// err := json.Unmarshal(data, cd)
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	err = decoder.Decode(cd)
	if err != nil {
		// fmt.Println("CacheData:", err)
		return cd, err
	}
	return cd, nil
}
