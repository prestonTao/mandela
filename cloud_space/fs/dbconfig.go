package fs

const (
	StoreSqlNamePre  = "store_"  //数据库文件名称前缀
	StoreSqlNameTail = "_sql.db" //

	StoreFileindex = "fileindex_sql.db" //
)

//import (
//	"mandela/core/utils"
//	"crypto/rand"
//	"encoding/hex"
//	"fmt"
//	"strconv"

//	"github.com/go-xorm/xorm"
//)

//type Storage struct {
//	sqldb    *xorm.Engine //sql数据库保存leveldb数据库中的key。
//	DbPath   string       //数据库目录
//	SpaceNum int          //默认占用空间大小 单位byte
//	PerSpace int          //每块大小 单位byte
//	TableNum int          //分表数量
//}

//func (this *Storage) Find(key string) *[]byte {
//	for i := 0; i < this.TableNum; i++ {
//		sqlStr := "select value from key" + strconv.Itoa(i) + " where key=?"
//		result, err := this.sqldb.Query(sqlStr, key)
//		if err != nil {
//			fmt.Println(err)
//			return nil
//		}
//		if len(result) <= 0 {
//			continue
//		}
//		bs := result[0]["value"]
//		return &bs
//	}
//	return nil
//}

//func (this *Storage) FullTable() {
//	for i := 0; i < this.TableNum; i++ {
//		this.fullTableOne(i)
//	}
//}

//func (this *Storage) fullTableOne(n int) {

//	sqlStr := "INSERT INTO `key" + strconv.Itoa(n) + "` (`key`,`value`) VALUES (?,?)"

//	num := this.SpaceNum / this.TableNum / this.PerSpace
//	fmt.Println("init data space ...", num)
//	for i := 0; i < num; i++ {
//		b := make([]byte, this.PerSpace)
//		if _, err := rand.Read(b); err != nil {
//			fmt.Println(err)
//			return
//		}
//		key := utils.Hash_SHA3_256(b)
//		keyStr := hex.EncodeToString(key)

//		this.sqldb.Exec(sqlStr, keyStr, b)
//	}
//}

//func NewStorage(abspath string) *Storage {

//	engineDB, err := xorm.NewEngine("sqlite3", "file:"+abspath+"?cache=shared")
//	if err != nil {
//		fmt.Println(err)
//		return nil
//	}
//	engineDB.ShowSQL(false)

//	s := &Storage{
//		sqldb:    engineDB,                //sql数据库保存leveldb数据库中的key。
//		DbPath:   abspath,                 //数据库目录
//		SpaceNum: 1024 * 1024 * 1024 * 10, //默认占用空间大小 单位byte
//		PerSpace: 512 * 1024,              //每块大小 单位byte
//		TableNum: 10,                      //分表数量
//	}

//	//创建表格
//	for i := 0; i < s.TableNum; i++ {
//		//
//		sqlStr := "CREATE TABLE IF NOT EXISTS `key" + strconv.Itoa(i) + "` (`key` TEXT NULL, `value` BOLB NULL)"
//		engineDB.Exec(sqlStr)
//	}
//	return s

//}
