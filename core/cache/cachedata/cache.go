package cachedata

import (
	"sync"
	"time"
	"mandela/core/utils"
)

//存储结构
type Cache struct {
	Data *sync.Map //存储共享数据(key/value:hash/cachedata)
	Time time.Time //更新时间
}

func NewCache() *Cache {
	c := new(Cache)
	c.Data = new(sync.Map)
	c.Time = time.Now()
	return c
}

//格式化保存key,value
//uptime true 则更新时间
func (c *Cache) Save(d *CacheData) {
	k, v := d.buildKV()
	c.Data.Store(k, v)
}

//根据KEY获取原始数据DATA
func (c *Cache) Get(key []byte) []byte {
	khash := buildHash(key)
	k := khash.B58String()
	data, ok := c.Data.Load(k)
	if ok {
		b := data.([]byte)
		cd, err := Parse(b)
		if err != nil {
			return nil
		}
		return cd.Value
	}
	return nil
}

//根据KEY删除数据
func (c *Cache) Del(key []byte) {
	khash := buildHash(key)
	k := khash.B58String()
	c.Data.Delete(k)
}

//根据hash id 获取原始数据DATA
func (c *Cache) GetByHash(id *utils.Multihash) []byte {
	k := id.B58String()
	data, ok := c.Data.Load(k)
	if ok {
		b := data.([]byte)
		cd, err := Parse(b)
		if err != nil {
			return nil
		}
		return cd.Value
	}
	return nil
}

//根据原始key获取cachedata
func (c *Cache) GetCacheData(key []byte) *CacheData {
	khash := buildHash(key)
	k := khash.B58String()
	data, ok := c.Data.Load(k)
	if ok {
		b := data.([]byte)
		cd, err := Parse(b)
		if err != nil {
			return nil
		}
		return cd
	}
	return nil
}

//根据hash id 获取cachedata
func (c *Cache) GetCacheDataByHash(id *utils.Multihash) *CacheData {
	k := id.B58String()
	data, ok := c.Data.Load(k)
	if ok {
		b := data.([]byte)
		cd, err := Parse(b)
		if err != nil {
			return nil
		}
		return cd
	}
	return nil
}
func (c *Cache) UpTime() {
	c.Time = time.Now()
}
