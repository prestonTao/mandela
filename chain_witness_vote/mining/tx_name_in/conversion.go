package tx_name_in

// import (
// 	"mandela/chain_witness_vote/mining"
// 	"mandela/core/nodeStore"
// 	"mandela/core/utils/crypto"
// 	"mandela/protobuf"

// 	"github.com/golang/protobuf/proto"
// )

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ConversionAccountBufVO(txbase Tx_account) ([]byte, error) {
// 	vins := make([]*protobuf.VinVO, 0)
// 	for i, _ := range txbase.Vin {
// 		vinOne := protobuf.VinVO{
// 			Txid: txbase.Vin[i].Txid,
// 			Vout: txbase.Vin[i].Vout,
// 			Puk:  txbase.Vin[i].Puk,
// 			Sign: txbase.Vin[i].Sign,
// 		}
// 		vins = append(vins, &vinOne)
// 	}
// 	vouts := make([]*protobuf.VoutVO, 0)
// 	for i, _ := range txbase.Vout {
// 		voutOne := protobuf.VoutVO{
// 			Value:   txbase.Vout[i].Value,
// 			Address: crypto.AddressCoin(txbase.Vout[i].Address),
// 			Tx:      txbase.Vout[i].Tx,
// 		}
// 		vouts = append(vouts, &voutOne)

// 	}

// 	addrNets := make([][]byte, 0)
// 	for i, _ := range txbase.NetIds {
// 		addrNets = append(addrNets, txbase.NetIds[i])
// 	}

// 	addrCoins := make([][]byte, 0)
// 	for i, _ := range txbase.AddrCoins {
// 		addrCoins = append(addrCoins, txbase.AddrCoins[i])
// 	}

// 	txvoteinVO := protobuf.TxAccountVO{
// 		Hash:       txbase.Hash,
// 		Type:       txbase.Type,
// 		VinTotal:   txbase.Vin_total,
// 		Vin:        vins,
// 		VoutTotal:  txbase.Vout_total,
// 		Vout:       vouts,
// 		Gas:        txbase.Gas,
// 		LockHeight: txbase.LockHeight,
// 		Payload:    txbase.Payload,
// 		BlockHash:  txbase.BlockHash,

// 		Account:             txbase.Account,
// 		NetIds:              addrNets,
// 		NetIdsMerkleHash:    txbase.NetIdsMerkleHash,
// 		AddrCoins:           addrCoins,
// 		AddrCoinsMerkleHash: txbase.AddrCoinsMerkleHash,
// 	}
// 	return proto.Marshal(&txvoteinVO)
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ParseAccountBufVO(bs []byte) (*Tx_account, error) {
// 	txvoteinVO := new(protobuf.TxAccountVO)
// 	err := proto.Unmarshal(bs, txvoteinVO)
// 	if err != nil {
// 		return nil, err
// 	}

// 	vins := make([]mining.Vin, 0)
// 	for i, _ := range txvoteinVO.Vin {
// 		vinOne := mining.Vin{
// 			Txid: txvoteinVO.Vin[i].Txid,
// 			Vout: txvoteinVO.Vin[i].Vout,
// 			Puk:  txvoteinVO.Vin[i].Puk,
// 			Sign: txvoteinVO.Vin[i].Sign,
// 		}
// 		vins = append(vins, vinOne)
// 	}

// 	vouts := make([]mining.Vout, 0)
// 	for i, _ := range txvoteinVO.Vout {
// 		voutVO := mining.Vout{
// 			Value:   txvoteinVO.Vout[i].Value,
// 			Address: txvoteinVO.Vout[i].Address,
// 			Tx:      txvoteinVO.Vout[i].Tx,
// 		}
// 		vouts = append(vouts, voutVO)
// 	}

// 	txbase := mining.TxBase{
// 		Hash:       txvoteinVO.Hash,
// 		Type:       txvoteinVO.Type,
// 		Vin_total:  txvoteinVO.VinTotal,
// 		Vin:        vins,
// 		Vout_total: txvoteinVO.VoutTotal,
// 		Vout:       vouts,
// 		Gas:        txvoteinVO.Gas,
// 		LockHeight: txvoteinVO.LockHeight,
// 		Payload:    txvoteinVO.Payload,
// 		BlockHash:  txvoteinVO.BlockHash,
// 	}

// 	netids := make([]nodeStore.AddressNet, 0)
// 	for i, _ := range txvoteinVO.NetIds {
// 		netids = append(netids, txvoteinVO.NetIds[i])
// 	}

// 	addrCoins := make([]crypto.AddressCoin, 0)
// 	for i, _ := range txvoteinVO.AddrCoins {
// 		addrCoins = append(addrCoins, txvoteinVO.AddrCoins[i])
// 	}

// 	txvin := Tx_account{
// 		TxBase:              txbase,
// 		Account:             txvoteinVO.Account,             //账户名称
// 		NetIds:              netids,                         //网络地址列表
// 		NetIdsMerkleHash:    txvoteinVO.NetIdsMerkleHash,    //网络地址默克尔树hash
// 		AddrCoins:           addrCoins,                      //网络地址列表
// 		AddrCoinsMerkleHash: txvoteinVO.AddrCoinsMerkleHash, //网络地址默克尔树hash
// 	}
// 	return &txvin, nil
// }
