package fs

// "mandela/core/virtual_node"

type FileindexNet struct {
	Id     uint64 `xorm:"pk autoincr unique 'id'"` //id
	Vid    string `xorm:"TEXT 'vid'"`              //虚拟节点id
	FileId string `xorm:"TEXT 'fileid'"`           //索引哈市值
	Value  []byte `xorm:"Blob 'value'"`            //内容
	//	Status int    `xorm:"int 'status'"`            //好友状态.1=添加好友时，用户不在线;2=申请添加好友状态;3=同意添加;4=;5=;6=;
}

func (this *FileindexNet) TableName() string {
	return "fileindex_net"
}

func (this *FileindexNet) Add(f *FileindexNet) error {
	_, err := engineDB.Insert(f)
	return err
}

func (this *FileindexNet) Del(fid string) error {
	_, err := engineDB.Where("fileid = ?", fid).Unscoped().Delete(this)
	return err
}

func (this *FileindexNet) Update() error {
	_, err := engineDB.Where("nodeid = ?", this.FileId).Update(this)
	return err
}

//修改
func (this *FileindexNet) UpdateValue(fileid string, value []byte) error {
	this.Value = value
	_, err := engineDB.Where("fileid = ?", fileid).Update(this)
	return err
}
func (this *FileindexNet) Getall() ([]FileindexNet, error) {

	fs := make([]FileindexNet, 0)
	err := engineDB.Find(&fs)
	return fs, err
}

/*
	检查用户id是否存在
*/
func (this *FileindexNet) FindByVid(vid string) (*FileindexNet, error) {
	fs := make([]FileindexNet, 0)
	err := engineDB.Where("vid = ?", vid).Find(&fs)
	if err != nil {
		return nil, err
	}
	if len(fs) <= 0 {
		return nil, nil
	}
	return &fs[0], nil
}

/*
	检查文件是否存在
*/
func (this *FileindexNet) FindByFileid(fid string) (*FileindexNet, error) {
	fs := make([]FileindexNet, 0)
	err := engineDB.Where("fileid = ?", fid).Find(&fs)
	if err != nil {
		return nil, err
	}
	if len(fs) <= 0 {
		return nil, nil
	}
	return &fs[0], nil
}
