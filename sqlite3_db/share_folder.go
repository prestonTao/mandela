package sqlite3_db

const (
	split = "*|*"
)

/*
	共享目录
*/
type ShareFolder struct {
	Path string `xorm:"varchar(25) pk notnull unique 'path'"`
}

/*
	添加一个共享目录
*/
func (this *ShareFolder) Add(path string) error {
	// dirs := utils.FilePathSplit(path)
	// str := strings.Join(dirs, split)
	// fmt.Println("处理后的路径", str)
	this.Path = path
	_, err := engineDB.Insert(this)
	return err
}

func (this *ShareFolder) Del(path string) {

	// dirs := utils.FilePathSplit(path)
	// str := strings.Join(dirs, split)
	this.Path = path
	engineDB.Where("path=?", path).Unscoped().Delete(this)
}

func (this *ShareFolder) GetAll() ([]ShareFolder, error) {
	sf := make([]ShareFolder, 0)
	err := engineDB.Find(&sf)
	return sf, err
}
