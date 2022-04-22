package fs

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"strconv"
	"sync"

	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/util"
	"github.com/go-xorm/xorm"

	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/logoove/sqlite"
)

const (
	prefix = "db_"
)

//初始占用空间
type Space struct {
	Db       *leveldb.DB
	sqldb    *sql.DB //sql数据库保存leveldb数据库中的key。
	DbPath   string  //数据库目录
	SpaceNum int     //默认占用空间大小 单位byte
	PerSpace int     //每块大小 单位byte
}

func NewSpace(dbpath string) *Space {
	db, err := leveldb.OpenFile(dbpath, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	spacenum := 1024 * 1024 * 1024
	perspace := 1024

	space := &Space{
		Db:       db,
		DbPath:   dbpath,
		SpaceNum: spacenum,
		PerSpace: perspace,
	}
	return space
}

//生成固定大小的字符集
func (s *Space) randStr() []byte {
	n := s.PerSpace //每块10MB
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}

//获取剩余空间 单位 pernum(10MB)
func (s *Space) FreeSpace() int {
	var free int
	iter := s.Db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		free += 1
	}
	iter.Release()
	return free
}

//新增数据
func (s *Space) Set(key, value []byte) bool {
	iter := s.Db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		err := s.Db.Delete(iter.Key(), nil)
		if err == nil {
			s.Db.Put(key, value, nil)
			return true
		}
	}
	iter.Release()
	return false
}

//获取数据
func (s *Space) Get(key []byte) []byte {
	val, err := s.Db.Get(key, nil)
	if err != nil {
		return nil
	}
	return val
}

//初始化填充数据库
func (s *Space) Init() error {
	var err error
	//	batch := new(leveldb.Batch)
	num := s.SpaceNum / s.PerSpace
	fmt.Println("init data space ...", num)
	for i := 0; i < num; i++ {
		key := prefix + strconv.Itoa(i)
		_, err := s.Db.Get([]byte(key), nil)
		fmt.Println(key, err)
		//如查已经占用，则跳过
		if err == nil {
			continue
		}
		s.Db.Put([]byte(key), []byte(s.randStr()), nil)
		//		batch.Put([]byte(key), []byte(s.randStr()))
		//		//批量写入，每5块写入一次
		//		if i%5 == 0 {
		//			err = s.Db.Write(batch, nil)
		//		}
	}
	//	err = s.Db.Write(batch, nil)
	return err
}
