/*
	保存历史转账纪录
*/
package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/json"
	"math/big"
	"strconv"
)

/*
	历史记录管理
*/
type BalanceHistory struct {
	GenerateMaxId *big.Int //自增长最高id，保存最新生成的id，可以直接来拿使用
	ForkNo        uint64   //分叉链id
}

type HistoryItem struct {
	GenerateId *big.Int              //
	IsIn       bool                  //资金转入转出方向，true=转入;false=转出;
	Type       uint64                //交易类型
	InAddr     []*crypto.AddressCoin //输入地址
	OutAddr    []*crypto.AddressCoin //输出地址
	Value      uint64                //交易金额
	Txid       []byte                //交易id
	Height     uint64                //区块高度
	//OutIndex   uint64                //交易输出index，从0开始
}

/*
	添加一个交易历史记录
*/
func (this *BalanceHistory) Add(hi HistoryItem) error {
	// fmt.Println("添加交易历史记录", hi)
	// fmt.Println(hi.GenerateId)
	if hi.GenerateId == nil {
		hi.GenerateId = this.GenerateMaxId
		this.GenerateMaxId = new(big.Int).Add(this.GenerateMaxId, big.NewInt(1))
	} else {
		if hi.GenerateId.Cmp(this.GenerateMaxId) == 0 {
			this.GenerateMaxId = new(big.Int).Add(this.GenerateMaxId, big.NewInt(1))
		}
	}
	bs, err := json.Marshal(hi)
	if err != nil {
		return err
	}

	key := []byte(config.LEVELDB_Head_history_balance + strconv.Itoa(int(this.ForkNo)) + "_" + hi.GenerateId.String())
	// fmt.Println("key", string(key), "\n", string(bs))
	return db.Save(key, &bs)
}

/*
	获取交易历史记录
*/
func (this *BalanceHistory) Get(start *big.Int, total int) []HistoryItem {
	if total == 0 {
		total = config.Wallet_balance_history
	}
	if start == nil {
		start = new(big.Int).Sub(this.GenerateMaxId, big.NewInt(1))
	}
	his := make([]HistoryItem, 0)

	key := config.LEVELDB_Head_history_balance + strconv.Itoa(int(this.ForkNo)) + "_"
	for i := 0; i < total; i++ {
		keyOne := key + new(big.Int).Sub(start, big.NewInt(int64(i))).String()
		bs, err := db.Find([]byte(keyOne))
		if err != nil {
			continue
		}
		hi := new(HistoryItem)

		decoder := json.NewDecoder(bytes.NewBuffer(*bs))
		decoder.UseNumber()
		err = decoder.Decode(hi)

		// err = json.Unmarshal(*bs, hi)
		if err != nil {
			continue
		}
		his = append(his, *hi)
	}
	return his
}

func NewBalanceHistory() *BalanceHistory {
	return &BalanceHistory{
		// ForkNo:        forkNo,
		GenerateMaxId: big.NewInt(0),
	}
}
