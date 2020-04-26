package sqlite

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var once sync.Once

var db *sql.DB

var engineDB *xorm.Engine

func Init() {
	once.Do(connect)

}

func connect() {
	var err error
	engineDB, err = xorm.NewEngine("sqlite3", "file:address_balance.db?cache=shared")
	if err != nil {
		fmt.Println(err)
	}
	engineDB.ShowSQL(false)

	ok, err := engineDB.IsTableExist(Balance{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(Balance{}).CreateTable(Balance{}) //创建表格
		table_friends := engineDB.Table(Balance{})       //切换表格
		table_friends.CreateIndexes(Balance{})           //创建索引
		table_friends.CreateUniques(Balance{})           //创建唯一性约束
	} else {
		// table_friends = engineDB.Table(Balance{})         //切换表格
		// table_sharefolder = engineDB.Table(ShareFolder{}) //
	}

}
