/*
	管理根目录列表
*/
package sharebox

// import (
// 	"sync"
// )

// //所有共享文件夹的根目录，具有排序功能
// var shareFolderRoots = make([]string, 0)
// var shareFolderRootsLock = new(sync.RWMutex)

// /*
// 	添加一个文件夹
// */
// func AddShareFolderRoots(folder string) {
// 	shareFolderRootsLock.Lock()
// 	shareFolderRoots = append(shareFolderRoots, folder)
// 	shareFolderRootsLock.Unlock()
// }

// /*
// 	删除一个文件夹
// */
// func RemoveShareFolderRoots(folder string) {
// 	shareFolderRootsLock.Lock()
// 	for i, one := range shareFolderRoots {
// 		if one == folder {
// 			temp := shareFolderRoots[:i]
// 			temp = append(temp, shareFolderRoots[i+1:]...)
// 			shareFolderRoots = temp
// 			break
// 		}
// 	}
// 	shareFolderRootsLock.Unlock()
// }

// /*
// 	设置共享文件夹根目录列表，用于排序
// */
// func SetShareFolderRoots(folders []string) {
// 	shareFolderRootsLock.Lock()
// 	shareFolderRoots = folders
// 	shareFolderRootsLock.Unlock()
// }

// /*
// 	获取根目录列表
// */
// func GetShareFolderRoots() []string {
// 	shareFolderRootsLock.RLock()
// 	temp := shareFolderRoots
// 	shareFolderRootsLock.RUnlock()
// 	return temp
// }
