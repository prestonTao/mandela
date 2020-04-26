package fs

import (
	"mandela/config"
	"mandela/core/utils"
	"mandela/core/virtual_node"
	"crypto/rand"
	"fmt"
	"os"
	"strconv"

	"github.com/go-xorm/xorm"
)

const (
	del  = "del"
	stop = "stop"
)

type Storage struct {
	VnodeId  virtual_node.AddressNetExtend //
	sqldb    *xorm.Engine                  //sql数据库保存leveldb数据库中的key。
	DbPath   string                        //数据库目录
	SpaceNum uint64                        //默认占用空间大小 单位byte
	PerSpace uint64                        //每块大小 单位byte
	TableNum uint64                        //分表数量
	UseSize  uint64                        //已经使用的空间大小，这里是自己上传的文件大小
	stopChan chan string                   //
}

func (this *Storage) StopAndDel() {
	select {
	case this.stopChan <- del:
	default:
	}
}

func (this *Storage) Stop() {
	select {
	case this.stopChan <- stop:
	default:
	}
}

/*
	获取一个未使用的块
*/
func (this *Storage) FindNotUseChunk() (uint64, []byte) {
	for i := uint64(0); i < this.TableNum; i++ {
		sqlStr := "select id from key" + strconv.Itoa(int(i)) + " where status = ? LIMIT 1"
		result, err := this.sqldb.Query(sqlStr, 1)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(result) <= 0 {
			continue
		}
		bs := result[0]["id"]
		return i, bs
	}
	return 0, nil
}

/*
	保存一个块
*/
func (this *Storage) Save(tnum uint64, id []byte, key virtual_node.AddressNetExtend, value *[]byte) {
	sqlStr := "UPDATE key" + strconv.Itoa(int(tnum)) + " SET key = ?,value = ?,status = ? WHERE id = ?"
	this.sqldb.Exec(sqlStr, key, *value, 2, id)
}

func (this *Storage) Find(key virtual_node.AddressNetExtend) *[]byte {
	for i := uint64(0); i < this.TableNum; i++ {
		sqlStr := "select value from key" + strconv.Itoa(int(i)) + " where key=?"
		result, err := this.sqldb.Query(sqlStr, key)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if len(result) <= 0 {
			continue
		}
		bs := result[0]["value"]
		return &bs
	}
	return nil
}

func (this *Storage) FullTable() {
	fmt.Println("FullTable")
	for i := uint64(0); i < this.TableNum; i++ {
		result := this.fullTableOne(i)
		switch result {
		case stop:
			return
		case del:
			this.sqldb.Close()
			os.Remove(this.DbPath)
			return
		default:
		}
	}
}

func (this *Storage) fullTableOne(n uint64) string {
	fmt.Println("fullTableOne")
	var table interface{} = nil
	switch n {
	case 0:
		table = Key0{}
	case 1:
		table = Key1{}
	case 2:
		table = Key2{}
	case 3:
		table = Key3{}
	case 4:
		table = Key4{}
	case 5:
		table = Key5{}
	case 6:
		table = Key6{}
	case 7:
		table = Key7{}
	case 8:
		table = Key8{}
	case 9:
		table = Key9{}
	}

	total, err := this.sqldb.Count(table)
	if err != nil {
		fmt.Println("查询错误", err)
		return ""
	}

	//	sqlStr := "select count(*) as count from key" + strconv.Itoa(n)
	//	result, err := this.sqldb.Query(sqlStr)
	//	if err != nil {
	//		fmt.Println("查询错误", err)
	//		return
	//	}

	//	if len(result) <= 0 {
	//		fmt.Println("没有记录")
	//		return
	//	}
	//	countBs := result[0]["count"]

	//	bytesBuffer := bytes.NewBuffer(countBs)
	//	var total uint64
	//	binary.Read(bytesBuffer, binary.LittleEndian, &total)

	//	//	total := utils.BytesToUint64(countBs)
	//	fmt.Println(result, total)

	//	return

	//	ssc := sqlite3_db.StoreSpaceConfig{}
	//	total, err = ssc.Count()
	//	if err != nil {
	//		return
	//	}
	//	count := uint64(total)

	num := int64(this.SpaceNum / this.TableNum / this.PerSpace)
	if total < num {
		num = num - total
	} else if total > num {
		//TODO 把多余的删除掉
	} else {
		return ""
	}

	sqlStr := "INSERT INTO `key" + strconv.Itoa(int(n)) + "` (`key`,`status`,`value`) VALUES (?,?,?)"

	fmt.Println("init data space ...", num, total)
	for i := int64(0); i < num; i++ {
		b := make([]byte, this.PerSpace)
		if _, err := rand.Read(b); err != nil {
			fmt.Println(err)
			return ""
		}
		keyBs := utils.Hash_SHA3_256(b)

		key := virtual_node.AddressNetExtend(keyBs)
		_, err := this.sqldb.Exec(sqlStr, key, 1, b)
		if err != nil {
			return ""
		}
		select {
		case result := <-this.stopChan:
			return result
		default:
		}
	}
	return ""
}

func NewStorage(vid virtual_node.AddressNetExtend, abspath string) (*Storage, error) {

	engineDB, err := xorm.NewEngine("sqlite3", "file:"+abspath+"?cache=shared")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	engineDB.ShowSQL(config.SQL_SHOW)

	s := &Storage{
		VnodeId:  vid,                  //
		sqldb:    engineDB,             //sql数据库保存leveldb数据库中的key。
		DbPath:   abspath,              //数据库目录
		SpaceNum: config.Spacenum,      // 1024 * 1024 * 1024 * 10, //默认占用空间大小 单位byte
		PerSpace: config.Chunk_size,    //每块大小 单位byte
		TableNum: 10,                   //分表数量
		stopChan: make(chan string, 1), //
	}

	//创建表格
	for i := uint64(0); i < s.TableNum; i++ {
		//
		sqlStr := "CREATE TABLE IF NOT EXISTS `key" + strconv.Itoa(int(i)) +
			"` (`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, `key` TEXT NULL, `status` INTEGER NULL, `value` BOLB NULL)"
		_, err := engineDB.Exec(sqlStr)
		if err != nil {
			return nil, err
		}
	}
	return s, nil

}
