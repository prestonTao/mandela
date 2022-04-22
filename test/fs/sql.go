package main

import (
	"mandela/core/utils"
	"mandela/store/fs"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/go-xorm/xorm"
	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/logoove/sqlite"
)

const dbpath = "D:/test/hzzfiles/sqlkey10table.db"

var engineDB *xorm.Engine

func main() {
	fmt.Println("start")
	s := fs.NewStorage(dbpath)

	time.Sleep(time.Second * 10)

	//	s := NewStorage(dbpath)
	//	s.FullTable()
	start := time.Now()
	//	s.Find("fb645d014be8572d1df4d3e5c4e9dd0a80c85fa0fe8e6d87fdf0bb510f097190")
	s.Find("ff226c33120b3ea63659303ec89055fc332700b09d33d72d70977a194e61d773")
	fmt.Println(time.Now().Sub(start))
	fmt.Println("end")
	return

	//	connect()
	//	//	initdb()

	//	fmt.Println("开始创建索引")
	//	//	table_keys.CreateIndexes(Key{}) //创建索引
	//	fmt.Println("创建索引完成")

	//	fmt.Println("开始查询")
	//	startTime := time.Now()

	//	//查询一个没有的
	//	bs := getRandBytes(512)
	//	ks := make([]Key, 0)
	//	err := engineDB.Where("value = ?", bs).Find(&ks)
	//	if err != nil {
	//		fmt.Println("error:", err)
	//		return
	//	}
	//	fmt.Println(time.Now().Sub(startTime))
	//	if len(ks) <= 0 {
	//		fmt.Println("没有找到记录")
	//		return
	//	}
	//	fmt.Println("查询结果", ks[0].Id)

	fmt.Println("end")
}

func getRandBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}

//func initdb() {
//	fullTable(table_keys1)
//	fullTable(table_keys2)
//	fullTable(table_keys3)
//	fullTable(table_keys4)
//	fullTable(table_keys5)
//}

//func fullTable(table_keys interface{}) {
//	n := 512 * 1024 //每块 512KB
//	num := 1024 * 1024 * 1024 * 10 / n / 5
//	fmt.Println("init data space ...", num)
//	for i := 0; i < num; i++ {
//		b := make([]byte, n)
//		if _, err := rand.Read(b); err != nil {
//			fmt.Println(err)
//			return
//		}
//		k := Key{Value: b}
//		_, err := table_keys.Insert(&k)
//		if err != nil {
//			fmt.Println("end", i)
//			break
//		}
//	}
//}

type Key1 struct {
	Key
}
type Key2 struct {
	Key
}
type Key3 struct {
	Key
}
type Key4 struct {
	Key
}
type Key5 struct {
	Key
}
type Key6 struct {
	Key
}
type Key7 struct {
	Key
}
type Key8 struct {
	Key
}
type Key9 struct {
	Key
}
type Key10 struct {
	Key
}

type KeyInterface interface {
	Set(key string, value []byte)
}

type Key struct {
	Id     uint64 `xorm:"pk autoincr unique 'id'"` //id
	Key    string `xorm:"varchar(25) 'key'"`       //
	Value  []byte `xorm:"Blob 'value'"`            //
	Status int    `xorm:"int 'status'"`            //状态.1=添加;2=删除;3=;4=;5=;6=;
}

func (this *Key) Set(key string, value []byte) {
	this.Key = key
	this.Value = value
	this.Status = 1
}

//func connect() error {
//	fmt.Println("开始连接网络")
//	var err error
//	engineDB, err = xorm.NewEngine("sqlite3", "file:"+dbpath+"?cache=shared")
//	if err != nil {
//		fmt.Println(err)
//		return err
//	}
//	engineDB.ShowSQL(false)

//	ok, err := engineDB.IsTableExist(Key{})
//	if err != nil {
//		fmt.Println(err)
//	}
//	if !ok {
//		engineDB.Table(Key{}).CreateTable(Key{}) //创建表格
//		table_keys = engineDB.Table(Key{})       //切换表格
//		table_keys.CreateIndexes(Key{})          //创建索引
//		table_keys.CreateUniques(Key{})          //创建唯一性约束
//	} else {
//		table_keys = engineDB.Table(Key{})
//	}
//	return nil
//}

type Storage struct {
	sqldb    *xorm.Engine //sql数据库保存leveldb数据库中的key。
	DbPath   string       //数据库目录
	SpaceNum int          //默认占用空间大小 单位byte
	PerSpace int          //每块大小 单位byte
	TableNum int          //分表数量
}

func (this *Storage) Find(key string) *[]byte {
	for i := 0; i < this.TableNum; i++ {
		sqlStr := "select value from key" + strconv.Itoa(i) + " where key=?"
		result, err := this.sqldb.Query(sqlStr, key)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if len(result) <= 0 {
			continue
		}
		//		for k, _ := range result[0] {
		//			fmt.Println(k)
		//		}
		bs := result[0]["value"]
		return &bs
	}
	return nil
}

func (this *Storage) FullTable() {

	for i := 0; i < this.TableNum; i++ {
		this.fullTableOne(i)
	}

	//	this.fullTableOne(&Key1{})
	//	this.fullTableOne(&Key2{})
	//	this.fullTableOne(&Key3{})
	//	this.fullTableOne(&Key4{})
	//	this.fullTableOne(&Key5{})
	//	this.fullTableOne(&Key6{})
	//	this.fullTableOne(&Key7{})
	//	this.fullTableOne(&Key8{})
	//	this.fullTableOne(&Key9{})
	//	this.fullTableOne(&Key10{})
}

func (this *Storage) fullTableOne(n int) {

	//	sqlStr := "INSERT INTO `key0` (`key`,`value`) VALUES (?,?)"
	//	this.sqldb.SQL(sqlStr)

	//	table := this.sqldb.Table(obj)
	sqlStr := "INSERT INTO `key" + strconv.Itoa(n) + "` (`key`,`value`) VALUES (?,?)"

	//	n := this.PerSpace //每块 512KB
	num := this.SpaceNum / this.TableNum / this.PerSpace
	fmt.Println("init data space ...", num)
	for i := 0; i < num; i++ {
		b := make([]byte, this.PerSpace)
		if _, err := rand.Read(b); err != nil {
			fmt.Println(err)
			return
		}

		//		fmt.Println(len(b), hex.EncodeToString(b))

		key := utils.Hash_SHA3_256(b)
		keyStr := hex.EncodeToString(key)

		this.sqldb.Exec(sqlStr, keyStr, b)

		//		k := Key{
		//			Key:   keyStr,
		//			Value: b,
		//		}
		//		k1 := Key1{Key: k}

		//		//		obj.Set(keyStr, b)
		//		fmt.Println(keyStr)
		//		//		k := Key{Value: b}
		//		_, err := this.sqldb.Insert(k1)
		//		if err != nil {
		//			fmt.Println("end", i)
		//			break
		//		}
	}
}

func NewStorage(abspath string, index uint64) *Storage {
	var err error
	engineDB, err = xorm.NewEngine("sqlite3", "file:"+abspath+"?cache=shared")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	engineDB.ShowSQL(config.ShowSQL)

	//创建表格
	//	var k KeyInterface = &Key1{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}
	//	k = &Key2{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}
	//	k = &Key3{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}
	//	k = &Key4{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}
	//	k = &Key5{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}
	//	k = &Key6{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}
	//	k = &Key7{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}
	//	k = &Key8{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}
	//	k = &Key9{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}
	//	k = &Key10{}
	//	if ok, _ := engineDB.IsTableExist(k); !ok {
	//		engineDB.Table(k).CreateTable(k)
	//	}

	s := &Storage{
		sqldb:    engineDB,                //sql数据库保存leveldb数据库中的key。
		DbPath:   abspath,                 //数据库目录
		Index:    index,                   //
		SpaceNum: 1024 * 1024 * 1024 * 10, //默认占用空间大小 单位byte
		PerSpace: 512 * 1024,              //每块大小 单位byte
		TableNum: 10,                      //分表数量
	}

	for i := 0; i < s.TableNum; i++ {
		//
		sqlStr := "CREATE TABLE IF NOT EXISTS `key" + strconv.Itoa(i) + "` (`key` TEXT NULL, `value` BOLB NULL)"
		//	engineDB.SQL("sqlStr")
		engineDB.Exec(sqlStr)
	}
	return s

}
