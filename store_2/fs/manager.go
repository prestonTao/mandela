package fs

type FsManager struct {
}

/*
	添加空间
*/
func (this *FsManager) AddSpace(absPath string) {
	space := NewSpace(absPath)
	if space == nil {
		return
	}
	space.Init()

}

func NewFsManager() *FsManager {
	return &FsManager{}
}
