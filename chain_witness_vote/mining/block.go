package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"crypto/ed25519"
	"encoding/json"
)

const (
	BlockHead_Hash              = "Hash"
	BlockHead_Height            = "Height"
	BlockHead_MerkleRoot        = "MerkleRoot"
	BlockHead_Previousblockhash = "Previousblockhash"
	BlockHead_Nextblockhash     = "Nextblockhash"
	BlockHead_Sign              = "sign"
	BlockHead_Tx                = "Tx"
	BlockHead_Time              = "Time"
)

//var (

//	//	headBlock     = new(sync.Map)              //保存对应区块高度的区块头hash。key:uint64=区块高度;value:*[]byte=区块高度对应区块头hash;
//	lastBlockHead *BlockHead                   //最高区块
//	preBlockHead  *BlockHead                   //最高区块的上一个区块
//	syncBlock     = make(chan *BlockHeadVO, 1) //连续导入区块
//)

///*
//	从数据库加载区块
//*/
//func LoadBlockInDB() {

//}

/*
	区块头
*/
type BlockHead struct {
	Hash              []byte             `json:"Hash"`              //区块头hash
	Height            uint64             `json:"Height"`            //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
	GroupHeight       uint64             `json:"GroupHeight"`       //矿工组高度
	Previousblockhash []byte             `json:"Previousblockhash"` //上一个区块头hash
	Nextblockhash     []byte             `json:"Nextblockhash"`     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
	NTx               uint64             `json:"NTx"`               //交易数量
	MerkleRoot        []byte             `json:"MerkleRoot"`        //交易默克尔树根hash
	Tx                [][]byte           `json:"Tx"`                //本区块包含的交易id
	Time              int64              `json:"Time"`              //出块时间，unix时间戳
	Witness           crypto.AddressCoin `json:"Witness"`           //此块见证人地址
	Sign              []byte             `json:"sign"`              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	// Nonce             uint64             `json:"nonce"`             //随机数，用以调整当前区块头hash
}

/*
	构建默克尔树根
*/
func (this *BlockHead) BuildMerkleRoot() {
	this.MerkleRoot = utils.BuildMerkleRoot(this.Tx)
}

/*
	将需要hash的字段序列化，不包括Sign变量
*/
func (this *BlockHead) Serialize() *[]byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(utils.Uint64ToBytes(this.Height))
	buf.Write(utils.Uint64ToBytes(this.GroupHeight))
	buf.Write(this.Previousblockhash)

	buf.Write(utils.Uint64ToBytes(this.NTx))
	buf.Write(this.MerkleRoot)
	for _, one := range this.Tx {
		buf.Write(one)
	}
	buf.Write(utils.Uint64ToBytes(uint64(this.Time)))
	buf.Write(this.Witness)
	bs := buf.Bytes()
	return &bs
}

/*
	区块签名
*/
func (this *BlockHead) BuildSign(key crypto.AddressCoin) {

	bs := this.Serialize()

	_, prk, _, err := keystore.GetKeyByAddr(key, config.Wallet_keystore_default_pwd)
	if err != nil {
		return
	}
	signBs := keystore.Sign(prk, *bs)

	this.Sign = signBs
}

/*
	检查区块头合法性
*/
func (this *BlockHead) CheckBlockHead(puk []byte) bool {
	//检查签名是否正确
	bs := this.Serialize()
	pkey := ed25519.PublicKey(puk)
	if !ed25519.Verify(pkey, *bs, this.Sign) {
		return false
	}

	old := this.Hash
	// fmt.Println("检查区块头前", hex.EncodeToString(old))
	this.BuildHash()
	// fmt.Println("检查区块头后", hex.EncodeToString(this.Hash))
	if !bytes.Equal(old, this.Hash) {
		return false
	}
	if this.Height <= 1 {
		return true
	}

	return true

}

/*
	寻找幸运数字
	@zoroes        uint64       难度，前导零数量
	@stopSignal    chan bool    停止信号 true=已经找到；false=未找到，被终止；
*/
// func (this *BlockHead) FindNonce(zoroes uint64, stopSignal chan bool) chan bool {
// 	// fmt.Println("start 开始工作，寻找区块高度", this.Height, "幸运数字。请等待...")
// 	result := make(chan bool, 1)

// 	//TODO 测试区块分叉使用，发布版本可以删除
// 	// this.Nonce = uint64(utils.GetRandNum(20000))

// 	stop := false
// 	for !stop {
// 		this.Nonce++
// 		this.BuildHash()
// 		if utils.CheckNonce(this.Hash, zoroes) {
// 			result <- true
// 			// fmt.Println("end 停止工作，找到幸运数字", this.Height)
// 			return result
// 		}
// 		select {
// 		case <-stopSignal:
// 			// fmt.Println("end 停止工作，因外部中断", this.Height)
// 			// close(stopSignal)
// 			stop = true
// 		default:
// 		}
// 	}
// 	result <- false
// 	return result
// }

/*
	构建区块头hash
*/
func (this *BlockHead) BuildHash() {

	buf := bytes.NewBuffer(*this.Serialize())
	// buf.Write(utils.Uint64ToBytes(this.Height))
	// buf.Write(utils.Uint64ToBytes(this.GroupHeight))
	// buf.Write(this.Previousblockhash)

	// buf.Write(utils.Uint64ToBytes(this.NTx))
	// buf.Write(this.MerkleRoot)
	// for _, one := range this.Tx {
	// 	buf.Write(one)
	// }
	// buf.Write(utils.Uint64ToBytes(uint64(this.Time)))
	// buf.Write(this.Witness)
	// buf.Write(this.Sign)
	bs := buf.Bytes()

	this.Hash = utils.Hash_SHA3_256(bs)
}

/*
	构建区块头hash
*/
func (this *BlockHead) CheckHashExist() bool {
	return db.CheckHashExist(this.Hash)
}

// /*
// 	验证区块签名
// */
// func (this *BlockHead) Check() {

// }

/*
	保存到本地磁盘
*/
func (this *BlockHead) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, nil
}

/*
	解析区块头
*/
func ParseBlockHead(bs *[]byte) (*BlockHead, error) {
	bh := new(BlockHead)
	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err := decoder.Decode(bh)
	// err := json.Unmarshal(*bs, bh)
	if err != nil {
		return nil, err
	}
	return bh, nil
}

type BlockHeadVO struct {
	BH  *BlockHead `json:"bh"`  //区块
	Txs []TxItr    `json:"txs"` //交易明细
}

/*
	json格式化
*/
func (this *BlockHeadVO) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, nil
}

/*
	验证区块合法性
*/
func (this *BlockHeadVO) Verify() bool {
	for i, one := range this.Txs {
		one.BuildHash()
		if !bytes.Equal(this.BH.Tx[i], *one.GetHash()) {
			// fmt.Println("区块不合法", i, hex.EncodeToString(this.BH.Tx[i]), hex.EncodeToString(*one.GetHash()))
			//区块不合法
			engine.Log.Info("Illegal block")
			return false
		}
	}
	return true

}

/*
	创建
*/
func CreateBlockHeadVO(bh *BlockHead, txs []TxItr) *BlockHeadVO {
	bhvo := BlockHeadVO{
		BH:  bh,  //
		Txs: txs, //交易明细
	}
	return &bhvo
}

type BlockHeadVOParse struct {
	BH  *BlockHead    `json:"bh"`  //区块
	Txs []interface{} `json:"txs"` //交易明细
	BM  *BackupMiners `json:"bm"`  //见证人投票结果
}

/*
	解析区块头
*/
func ParseBlockHeadVO(bs *[]byte) (*BlockHeadVO, error) {
	bh := new(BlockHeadVOParse)

	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err := decoder.Decode(bh)

	// err := json.Unmarshal(*bs, bh)
	if err != nil {
		return nil, err
	}

	txitrs := make([]TxItr, 0)
	for _, one := range bh.Txs {

		oneMap := one.(map[string]interface{})
		bs, err := json.Marshal(oneMap)

		// fmt.Println("1234567890", one.(string))
		// bs, err := json.Marshal(one)
		if err != nil {
			return nil, err
		}
		txitr, err := ParseTxBase(&bs)
		if err != nil {
			return nil, err
		}
		txitrs = append(txitrs, txitr)
	}
	bhvo := BlockHeadVO{
		BH:  bh.BH,  //区块
		Txs: txitrs, //交易明细
		//		BM:  bh.BM,  //见证人投票结果
	}
	return &bhvo, nil
}
