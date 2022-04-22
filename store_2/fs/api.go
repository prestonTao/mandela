package fs

var fsm = NewFsManager()

/*
	添加一个空间
	@filepath    string    空间文件的保存路径
*/
func AddSpace(absPath string) {
	fsm.AddSpace(absPath)
}

/*
	检查空间是否够用
*/
func CheckSpace(newSize uint64) bool {
	//TODO
}
