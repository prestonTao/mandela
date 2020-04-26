package sharebox

import (
	"mandela/core/utils"
	"mandela/sharebox/config"
	sqldb "mandela/sqlite3_db"
	"os"
)

func RegsterStore() error {
	// sqldb.Init()

	//删除本地所有文件分片及分片索引
	if config.IsRemoveStore {
		err := os.RemoveAll(config.Store_dir)
		if err != nil {
			return err
		}
	}

	//创建保存文件的文件夹
	utils.CheckCreateDir(config.Store_dir)
	//创建网络文件缓存文件夹
	utils.CheckCreateDir(config.Store_cache)
	//创建保存文件索引的文件夹
	// utils.CheckCreateDir(filepath.Join(config.Store_fileinfo_self))
	//创建保存文件索引的文件夹
	// utils.CheckCreateDir(config.Store_fileinfo_local)
	//创建保存文件索引的文件夹
	// utils.CheckCreateDir(config.Store_fileinfo_net)
	//创建保存文件索引的文件夹
	// utils.CheckCreateDir(config.Store_fileinfo_cache)
	//创建临时文件夹
	// utils.CheckCreateDir(config.Store_temp)

	//链接数据库
	sqldb.Init()

	// err := db.InitDB(config.DB_Path)
	// if err != nil {
	// 	return err
	// }

	initTask()

	err := InitWatcher()
	if err != nil {
		return err
	}

	err = loadShareFolder()
	if err != nil {
		return err
	}

	// folders, err := db.GetFolders()
	// if err != nil {
	// 	return err
	// }
	// err = AddLocalShareFolders(folders...)
	// if err != nil {
	// 	return err
	// }

	//加载自己共享的文件
	// err = LoadFileInfoSelf()
	// if err != nil {
	// 	return err
	// }
	//加载本地文件索引
	// err = LoadFileInfoLocal()
	// if err != nil {
	// 	return err
	// }
	//加载网络文件索引
	err = LoadFileInfoNet()
	if err != nil {
		return err
	}
	RegisterMsgid()

	return nil
}
