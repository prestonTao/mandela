package fs

type StorageVO struct {
	// VnodeId  virtual_node.AddressNetExtend //
	DbPath   string //数据库目录
	SpaceNum uint64 //默认占用空间大小 单位byte
	PerSpace uint64 //每块大小 单位byte
	TableNum uint64 //分表数量
	UseSize  uint64 //已经使用的空间大小，这里是自己上传的文件大小
}

func ConvertStorageVO(s *Storage) StorageVO {
	return StorageVO{
		DbPath:   s.DbPath,   //数据库目录
		SpaceNum: s.SpaceNum, //默认占用空间大小 单位byte
		PerSpace: s.PerSpace, //每块大小 单位byte
		TableNum: s.TableNum, //分表数量
		UseSize:  s.UseSize,  //已经使用的空间大小，这里是自己上传的文件大小
	}
}
