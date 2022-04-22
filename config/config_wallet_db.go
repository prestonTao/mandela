/*
	数据库中key保存的格式
	1.保存第一个区块的hash:startblock
	2.保存区块头:[区块hash]
	3.交易历史纪录:1_[1/2]_[自己的地址]_[目标地址]
*/
package config

const (
	block_start_str              = "startblock"   //保存第一个区块hash
	History                      = "1_"           //key以“1_”前缀的含义：交易历史数据
	BlockHeight                  = "2_"           //key以“2_”前缀的含义：块高度和区块hash对应
	Name                         = "name_"        //key以“name_”前缀的含义：已经注册的域名
	Block_Highest                = "HighestBlock" //保存最高区块数量
	LEVELDB_Head_history_balance = "h_b_"         //交易历史记录 格式：h_b_[分叉链编号]_[GenerateId]例：h_b_0_100
	WitnessName                  = "3_"           //key以“3_”前缀的含义：见证人名称和社区节点名称对应的地址
	WitnessAddr                  = "4_"           //key以“4_”前缀的含义：见证人地址和社区节点地址对应的名称
	AlreadyUsed_tx               = "5_"           //key以“5_”前缀的含义：未花费的交易余额索引 格式：5_[交易hash]_[输出index]
	TokenInfo                    = "6_"           //key以“6_”前缀的含义：发布token的信息(单位，总量) 格式：6_[交易hash]
	TokenPublishTxid             = "7_"           //key以“7_”前缀的含义：发布token的合约txid 格式：7_[交易hash]_[输出index]
	DB_PRE_Tx_Not_Import         = "8_"           //key以“8_”前缀的含义：未导入区块中的交易txid，导入后删除这个key 格式：8_[交易hash]
	DB_spaces_mining_addr        = "9_"           //key以“9_”前缀的含义：质押地址对应的质押金额 格式：9_[质押地址]
	DB_community_addr            = "10_"          //key以“9_”前缀的含义：成为社区节点的交易区块hash 格式：9_[社区节点地址]
	DBKEY_tx_blockhash           = "11_"          //key以“11_”前缀的含义：保存交易所属的区块hash 格式：11_[交易hash]
)

var (
	Key_block_start = []byte(block_start_str) //保存第一个区块hash
	// DB_start_block_time = 0                       //创始区块创建时间
)

///*
//	构建交易历史转入key
//*/
//func BuildHistoryInKey(self, tag string) []byte {
//	return []byte(History + In + self + "_" + tag)
//}

///*
//	构建交易历史转出key
//*/
//func BuildHistoryOutKey(self, tag string) []byte {
//	return []byte(History + Out + self + "_" + tag)
//}

/*
	构建未导入区块的交易key标记
	将未导入的区块中的交易使用此key保存到数据库中作为标记，如果已经导入过区块，则删除此标记
	用作验证已存在的交易hash
*/
func BuildTxNotImport(txid []byte) []byte {
	return append([]byte(DB_PRE_Tx_Not_Import), txid...)
}

/*
	保存交易所属的区块hash
*/
func BuildTxToBlockHash(txid []byte) []byte {
	return append([]byte(DBKEY_tx_blockhash), txid...)
}
