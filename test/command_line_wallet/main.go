package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/engine"
	"mandela/rpc"
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	scan()
}

func scan() {
	engine.Log.Info("start init leveldb")
	db.InitDB(filepath.Join("wallet", "data"), filepath.Join("wallet", "temp"))
	input := bufio.NewScanner(os.Stdin)

	engine.Log.Info("Please type in something:")
	// 逐行扫描
	for input.Scan() {
		line := input.Text()
		engine.Log.Info("command: %s", line)
		ok := parseHeight(line)
		if ok {
			continue
		}
		ok = parsehash(line)
		if ok {
			continue
		}

	}
}

func parseHeight(line string) bool {

	height, err := strconv.Atoi(line)
	if err == nil {
		//通过高度查
		bhash, err := db.LevelDB.Find([]byte(config.BlockHeight + strconv.Itoa(int(height))))
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		bs, err := db.LevelDB.Find(*bhash)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		bh, err := mining.ParseBlockHeadProto(bs)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		txs := make([]string, 0)
		for _, one := range bh.Tx {
			txs = append(txs, hex.EncodeToString(one))
		}
		bhvo := rpc.BlockHeadVO{
			Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
			Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
			GroupHeight:       bh.GroupHeight,                           //矿工组高度
			GroupHeightGrowth: bh.GroupHeightGrowth,                     //
			Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
			Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
			NTx:               bh.NTx,                                   //交易数量
			MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
			Tx:                txs,                                      //本区块包含的交易id
			Time:              bh.Time,                                  //出块时间，unix时间戳
			Witness:           bh.Witness.B58String(),                   //此块见证人地址
			Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		}
		*bs, _ = json.Marshal(bhvo)
		engine.Log.Info(string(*bs))
		engine.Log.Info("finish!")
		return true
	}
	return false
}

func parsehash(line string) bool {
	hash, err := hex.DecodeString(line)
	if err != nil {
		engine.Log.Info(err.Error())
		return false
	}

	bs, err := db.LevelDB.Find(hash)
	if err != nil {
		engine.Log.Info(err.Error())
		return false
	}

	bh, err := mining.ParseBlockHeadProto(bs)
	if err != nil {
		//可能是查交易
		txBase, err := mining.ParseTxBaseProto(mining.ParseTxClass(hash), bs)
		if err != nil {
			engine.Log.Info(err.Error())
			return false
		}

		itr := txBase.GetVOJSON()
		bs, _ := json.Marshal(itr)
		engine.Log.Info(string(bs))

		//查询这个交易属于哪个区块
		blockHash, err := db.LevelTempDB.Find(config.BuildTxToBlockHash(hash))
		if err == nil {
			engine.Log.Info("blockhash:%s", hex.EncodeToString(*blockHash))
		}

		engine.Log.Info("finish!")
		return true
	}

	if bh.Height <= 0 {
		//可能是查交易
		txBase, err := mining.ParseTxBaseProto(mining.ParseTxClass(hash), bs)
		if err != nil {
			engine.Log.Info(err.Error())
			return false
		}

		itr := txBase.GetVOJSON()
		bs, _ := json.Marshal(itr)
		engine.Log.Info(string(bs))

		//查询这个交易属于哪个区块
		blockHash, err := db.LevelTempDB.Find(config.BuildTxToBlockHash(hash))
		if err == nil {
			engine.Log.Info("blockhash:%s", hex.EncodeToString(*blockHash))
		}

		engine.Log.Info("正在验证签名")
		err = txBase.Check()
		if err != nil {
			engine.Log.Info(err.Error())
		}

		engine.Log.Info("finish!")
		return true
	}

	txs := make([]string, 0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
	}
	bhvo := rpc.BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		GroupHeightGrowth: bh.GroupHeightGrowth,                     //
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	}
	*bs, _ = json.Marshal(bhvo)
	engine.Log.Info(string(*bs))

	engine.Log.Info("finish!")
	return true
}
