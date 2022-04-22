package sqlite3_db

import (
	"bytes"
	"fmt"

	_ "github.com/go-xorm/xorm"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Property struct {
	Hash       string `xorm:"varchar(25) pk notnull unique 'hash'"` //网络节点hash
	Nickname   string `xorm:"varchar(25) 'nick_name'"`              //昵称
	CreateTime uint64 `xorm:"created 'createtime'"`                 //创建时间，这个Field将在Insert时自动赋值为当前时间
	UpdateTime uint64 `xorm:"updated 'updated'"`                    //修改后自动更新时间
}

func (this *Property) Json() []byte {
	rs, err := json.Marshal(this)
	if err != nil {
		fmt.Println(err)
	}
	return rs
}

//解析用户属性
func ParseProperty(bs []byte) Property {
	p := Property{}
	// err := json.Unmarshal(bs, &p)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(&p)
	if err != nil {
		fmt.Println(err)
	}
	return p
}
func (this *Property) Update() error {
	ok, err := engineDB.Where("hash=?", this.Hash).Exist(&Property{})
	if err != nil {
		return err
	}
	if ok {
		_, err = engineDB.Where("hash=?", this.Hash).Update(this)
	} else {
		_, err = engineDB.Insert(this)
	}

	return err
}
func (this *Property) Get(hash string) (dp Property) {
	_, err := engineDB.Where("hash = ?", hash).Get(&dp)
	if err != nil {
		fmt.Println(err)
	}
	return
}
