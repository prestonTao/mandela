package payment

// import (
// 	"mandela/chain_witness_vote/mining"
// 	"mandela/core/utils/crypto"
// 	"mandela/protobuf"

// 	"github.com/golang/protobuf/proto"
// )

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ConversionTokenPaymentBufVO(txbase TxToken) ([]byte, error) {
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

// 	tokenVins := make([]*protobuf.VinVO, 0)
// 	for i, _ := range txbase.Token_Vin {
// 		vinOne := protobuf.VinVO{
// 			Txid: txbase.Token_Vin[i].Txid,
// 			Vout: txbase.Token_Vin[i].Vout,
// 			Puk:  txbase.Token_Vin[i].Puk,
// 			Sign: txbase.Token_Vin[i].Sign,
// 		}
// 		tokenVins = append(tokenVins, &vinOne)
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

// 	txvoteinVO := protobuf.TxTokenPaymentVO{
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

// 		Token_VinTotal:   txbase.Token_Vin_total,
// 		Token_Vin:        tokenVins,
// 		Token_VoutTotal:  txbase.Token_Vout_total,
// 		Token_Vout:       tokenVouts,
// 		TokenPublishTxid: txbase.Token_publish_txid,
// 	}
// 	return proto.Marshal(&txvoteinVO)
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ParseTokenPaymentBufVO(bs []byte) (*TxToken, error) {
// 	txvoteinVO := new(protobuf.TxTokenPaymentVO)
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

// 	tokenVins := make([]mining.Vin, 0)
// 	for i, _ := range txvoteinVO.Token_Vin {
// 		vinOne := mining.Vin{
// 			Txid: txvoteinVO.Token_Vin[i].Txid,
// 			Vout: txvoteinVO.Token_Vin[i].Vout,
// 			Puk:  txvoteinVO.Token_Vin[i].Puk,
// 			Sign: txvoteinVO.Token_Vin[i].Sign,
// 		}
// 		vins = append(vins, vinOne)
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

// 		Token_Vin_total:    txvoteinVO.Token_VinTotal,   //输入交易数量
// 		Token_Vin:          tokenVins,                   //交易输入
// 		Token_Vout_total:   txvoteinVO.Token_VoutTotal,  //输出交易数量
// 		Token_Vout:         tokenVouts,                  //交易输出
// 		Token_publish_txid: txvoteinVO.TokenPublishTxid, //token的合约地址
// 	}
// 	return &txvin, nil
// }
