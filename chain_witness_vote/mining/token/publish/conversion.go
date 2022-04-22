package publish

// import (
// 	"mandela/chain_witness_vote/mining"
// 	"mandela/core/utils/crypto"
// 	"mandela/protobuf"

// 	"github.com/golang/protobuf/proto"
// )

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ConversionTokenPublishBufVO(txbase TxToken) ([]byte, error) {
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
// 			Address: txbase.Vout[i].Address,
// 			Tx:      txbase.Vout[i].Tx,
// 		}
// 		vouts = append(vouts, &voutOne)

// 	}

// 	tokenVouts := make([]*protobuf.VoutVO, 0)
// 	for i, _ := range txbase.Token_Vout {
// 		voutOne := protobuf.VoutVO{
// 			Value:   txbase.Token_Vout[i].Value,
// 			Address: crypto.AddressCoin(txbase.Token_Vout[i].Address),
// 			Tx:      txbase.Token_Vout[i].Tx,
// 		}
// 		tokenVouts = append(tokenVouts, &voutOne)
// 	}

// 	// addrNets := make([][]byte, 0)
// 	// for i, _ := range txbase.NetIds {
// 	// 	addrNets = append(addrNets, txbase.NetIds[i])
// 	// }

// 	// addrCoins := make([][]byte, 0)
// 	// for i, _ := range txbase.AddrCoins {
// 	// 	addrCoins = append(addrCoins, txbase.AddrCoins[i])
// 	// }

// 	txvoteinVO := protobuf.TxTokenPublishVO{
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

// 		TokenName:       txbase.Token_name,
// 		TokenSymbol:     txbase.Token_symbol,
// 		TokenSupply:     txbase.Token_supply,
// 		Token_VoutTotal: txbase.Token_Vout_total,
// 		Token_Vout:      tokenVouts,
// 	}
// 	return proto.Marshal(&txvoteinVO)
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ParseTokenPublishBufVO(bs []byte) (*TxToken, error) {
// 	txvoteinVO := new(protobuf.TxTokenPublishVO)
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

// 	tokenVouts := make([]mining.Vout, 0)
// 	for i, _ := range txvoteinVO.Token_Vout {
// 		voutOne := mining.Vout{
// 			Value:   txvoteinVO.Token_Vout[i].Value,
// 			Address: crypto.AddressCoin(txvoteinVO.Token_Vout[i].Address),
// 			Tx:      txvoteinVO.Token_Vout[i].Tx,
// 		}
// 		tokenVouts = append(tokenVouts, voutOne)
// 	}

// 	txvin := TxToken{
// 		TxBase: txbase,

// 		Token_name:       txvoteinVO.TokenName,       //名称
// 		Token_symbol:     txvoteinVO.TokenSymbol,     //单位
// 		Token_supply:     txvoteinVO.TokenSupply,     //发行总量
// 		Token_Vout_total: txvoteinVO.Token_VoutTotal, //输出交易数量
// 		Token_Vout:       tokenVouts,                 //交易输出
// 	}
// 	return &txvin, nil
// }
