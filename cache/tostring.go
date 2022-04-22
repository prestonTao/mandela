package cache

import (
	"mandela/cache/db"

	"encoding/hex"

	"github.com/mr-tron/base58"
)

type Cache struct {
	DbPath string
}

func newCache() *Cache {
	c := Cache{DbPath: "data"}
	return &c
}
func (c Cache) init() {
	db.InitDB(c.DbPath)
}
func (c Cache) b58string(str []byte) (rs string) {
	//return base58.Encode(str)
	rsb, _ := db.Find(str)
	if rsb == nil {
		rs = base58.Encode(str)
		db.Save(str, []byte(rs))
		return
	}
	rs = string(rsb)
	return
}
func (c Cache) hex(str []byte) (rs string) {
	rs = hex.EncodeToString(str)
	return
}
