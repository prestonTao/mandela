/*
	数据库中保存共享文件夹目录列表
*/
package db

// import (
// 	"mandela/sharebox/config"
// 	"encoding/json"

// 	"github.com/syndtr/goleveldb/leveldb"
// )

// func GetFolders() ([]string, error) {
// 	bs, err := Find(config.DB_ShareFolders)
// 	if err != nil {
// 		if err == leveldb.ErrNotFound {
// 			return []string{}, nil
// 		}
// 		return nil, err
// 	}
// 	folders := make([]string, 0)
// 	err = json.Unmarshal(*bs, folders)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return folders, nil
// }

// func SaveFolders(folders []string) error {
// 	bs, err := json.Marshal(folders)
// 	if err != nil {
// 		return err
// 	}
// 	return Save(config.DB_ShareFolders, &bs)
// }
