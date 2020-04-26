package config

import (
	"path/filepath"
)

const (
	SQLITE3DB_name               = "sqlite3.db"          //sqlite3数据库文件名称
	Table_name_friend            = "friend"              //
	Table_name_shareFolderRemote = "share_folder_remote" //远程共享目录
	SQL_SHOW                     = false                 //是否打印sql语句
)

var (
	SQLITE3DB_path = filepath.Join(Path_configDir, SQLITE3DB_name) //sqlite3数据库文件保存路径
)
