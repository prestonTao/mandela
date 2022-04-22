package pubstore

//种子导出
type Seed struct {
	Key       []byte //生成主密钥的随机数
	ChainCode []byte //主KDF链编码
	IV        []byte //aes加密向量
	CheckHash []byte //主私钥和链编码加密验证hash值
	Index     int    //地址索引
}

//输出公私钥
type PriPub struct {
	Prik string `json:"prik"`
	Pubk string `json:"pubk"`
}
