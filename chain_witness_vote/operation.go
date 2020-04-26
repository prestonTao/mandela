package chain_witness_vote

//import (
//	"mandela/core/utils"
//	"mandela/chain_witness_vote/mining"
//	"time"
//)

///*
//	付款
//	@value    uint64    付款金额
//	@tip      uint64    交易手续费，旷工费
//	@addr     string    付款地址
//*/
//func Pay(value, tip uint64, addr utils.Multihash) {
//	//TODO 查找钱包地址余额的UTXO交易id
//	var selfAddr utils.Multihash
//	var count uint64 = 10    //钱包余额
//	puk := []byte{}          //公钥
//	txid := []byte{}         //前交易id
//	var voutIndex uint64 = 0 //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
//	sign := []byte{}         //前一个交易的签名

//	//构建输出
//	vouts := make([]mining.Vout, 0)
//	vout := mining.Vout{
//		Value:   value, //输出金额 = 实际金额 * 100000000
//		Address: addr,  //钱包地址
//	}
//	vouts = append(vouts, vout)
//	//如果未配平，把剩余的钱打到自己地址上
//	if (count - value - tip) > 0 {
//		vout := mining.Vout{
//			Value:   count - value - tip, //输出金额 = 实际金额 * 100000000
//			Address: selfAddr,            //钱包地址
//		}
//		vouts = append(vouts, vout)
//	}

//	//构建输入
//	vins := make([]mining.Vin, 0)
//	vin := mining.Vin{
//		Txid: txid,      //UTXO 前一个交易的id
//		Vout: voutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
//		Puk:  puk,       //公钥
//		Sign: sign,      //签名
//		//	VoutSign string //对输出签名
//	}
//	vins = append(vins, vin)

//	base := mining.TxBase{
//		Type:       0,
//		Vin_total:  uint64(len(vins)),  //输入交易数量
//		Vin:        vins,               //交易输入
//		Vout_total: uint64(len(vouts)), //输出交易数量
//		Vout:       vouts,              //交易输出
//		CreateTime: time.Now().Unix(),  //创建时间
//	}

//	tx := mining.Tx_Pay{
//		TxBase: base, //交易类型，默认0=普通转账到地址交易
//	}

//	tx.BuildHash()

//}
