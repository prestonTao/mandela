package sqlite3_db

import (
	"errors"

	_ "github.com/go-xorm/xorm"
)

type MessageCache struct {
	Id         int64  `xorm:"pk autoincr unique 'id'"`  //id
	Hash       []byte `xorm:"Blob index unique 'hash'"` //消息Hash
	Head       []byte `xorm:"Blob 'head'"`              //消息头
	Body       []byte `xorm:"Blob 'body'"`              //消息体
	CreateTime int64  `xorm:"created 'createtime'"`     //创建时间，这个Field将在Insert时自动赋值为当前时间
}

/*
	添加一个消息记录
	@return    int64    数据库id
*/
func (this *MessageCache) Add(hash []byte, head, body []byte) error {
	dblock.Lock()
	mc := MessageCache{
		Hash: hash,
		Head: head,
		Body: body,
	}
	_, err := engineDB.Insert(&mc)
	dblock.Unlock()
	return err
}

/*
	查询一条记录
*/
func (this *MessageCache) FindByHash(hash []byte) (*MessageCache, error) {
	dblock.Lock()
	defer dblock.Unlock()
	mls := make([]MessageCache, 0)
	//err := table_msglog.Id(id).Find(&mls)
	err := engineDB.Where("hash=?", hash).Find(&mls)
	if err != nil {
		return nil, err
	}
	if len(mls) <= 0 {
		return nil, errors.New("not find")
	}
	return &mls[0], nil
}

/*
	查询过期的所有记录
*/
// func (this *MessageCache) GetOverTime(createtime int64) ([]MessageCache, error) {

// 	mc := make([]MessageCache, 0)
// 	err := engineDB.Cols("id", "hash").Where("createtime < ?", createtime).Find(&mc)
// 	// err := engineDB.Find(&mc)
// 	return mc, err
// }

/*
	删除多个消息记录
*/
func (this *MessageCache) Remove(createtime int64) error {
	if engineDB == nil {
		return nil
	}
	dblock.Lock()
	_, err := engineDB.Where("createtime < ?", createtime).Unscoped().Delete(this)
	dblock.Unlock()
	return err
}
