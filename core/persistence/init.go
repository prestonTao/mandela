package persistence

import (
	"database/sql"
	"errors"

	// "fmt"
	gconfig "mandela/config"
	"os"
	"path/filepath"

	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/logoove/sqlite"
)

const (
	folderName_db  = "db"
	folderName_msg = "message"

	friendsFileName = "friends"
)

var db *sql.DB

func Init() {
	//TODO 开发阶段删除目录，正式发布时去掉
	os.RemoveAll(filepath.Join(gconfig.Path_configDir, folderName_db))
	os.RemoveAll(filepath.Join(gconfig.Path_configDir, folderName_msg))

	//检查db文件夹是否存在，不存在则创建
	folderPath := filepath.Join(gconfig.Path_configDir, folderName_db)
	err := mkDir(folderPath)
	if err != nil {
		panic(err)
	}

	//检查message文件夹是否存在，不存在则创建
	folderPath = filepath.Join(gconfig.Path_configDir, folderName_msg)
	err = mkDir(folderPath)
	if err != nil {
		panic(err)
	}

	loadFriends()

	if err = initLogFile(); err != nil {
		panic(err)
	}

	if err = loadMsgTableKey(); err != nil {
		panic(err)
	}
}

/*
	判断文件夹是否存在，不存在则创建
*/
func mkDir(fpath string) error {
	_, err := os.Stat(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			//创建一个文件夹
			err = os.MkdirAll(fpath, 0777)
			if err != nil {
				//创建" + fpath + "文件夹失败
				return errors.New("create " + fpath + " dir fail")
			}
		} else {
			//读取" + fpath + "文件夹失败
			return errors.New("read " + fpath + " dir fail")
		}
	}
	return nil
}

/*
	初始化日志数据库
*/
func initLogFile() error {
	var err error
	db, err = sql.Open("sqlite3", filepath.Join(gconfig.Path_configDir, folderName_db, "message.db"))
	if err != nil {
		return err
	}
	//	defer db.Close()

	sqlStmt := `
CREATE TABLE message (
  id int(11) NOT NULL,
  sender varchar(255) DEFAULT NULL,
  recver varchar(255) DEFAULT NULL,
  content varchar(255) DEFAULT NULL,
  updatetime int(11) DEFAULT NULL,
  PRIMARY KEY (id)
);
	`

	//	sqlStmt := `
	//	create table friends (id integer not null primary key, name text);
	//	delete from friends;
	//	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		// fmt.Println(err)
		return err
	}
	return nil
}
