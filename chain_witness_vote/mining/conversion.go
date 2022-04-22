package mining

// import (
// 	"mandela/core/utils/crypto"
// 	"mandela/protobuf"

// 	"github.com/golang/protobuf/proto"
// )

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ConversionTxBaseBufVO(txbase TxBase) ([]byte, error) {
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

// 	txvoteinVO := protobuf.TxBaseVO{
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
// 	}
// 	return proto.Marshal(&txvoteinVO)
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ParseTxBaseBufVO(bs []byte) (*TxBase, error) {
// 	txvoteinVO := new(protobuf.TxVoteInVO)
// 	err := proto.Unmarshal(bs, txvoteinVO)
// 	if err != nil {
// 		return nil, err
// 	}

// 	vins := make([]Vin, 0)
// 	for i, _ := range txvoteinVO.Vin {
// 		vinOne := Vin{
// 			Txid: txvoteinVO.Vin[i].Txid,
// 			Vout: txvoteinVO.Vin[i].Vout,
// 			Puk:  txvoteinVO.Vin[i].Puk,
// 			Sign: txvoteinVO.Vin[i].Sign,
// 		}
// 		vins = append(vins, vinOne)
// 	}

// 	vouts := make([]Vout, 0)
// 	for i, _ := range txvoteinVO.Vout {
// 		voutVO := Vout{
// 			Value:   txvoteinVO.Vout[i].Value,
// 			Address: txvoteinVO.Vout[i].Address,
// 			Tx:      txvoteinVO.Vout[i].Tx,
// 		}
// 		vouts = append(vouts, voutVO)
// 	}

// 	txbase := TxBase{
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

// 	return &txbase, nil
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ConversionDepositInBufVO(txbase Tx_deposit_in) ([]byte, error) {
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

// 	txvoteinVO := protobuf.TxDepositInVO{
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
// 		Puk:        txbase.Puk,
// 	}
// 	return proto.Marshal(&txvoteinVO)
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ParseDepositInBufVO(bs []byte) (*Tx_deposit_in, error) {
// 	txvoteinVO := new(protobuf.TxDepositInVO)
// 	err := proto.Unmarshal(bs, txvoteinVO)
// 	if err != nil {
// 		return nil, err
// 	}

// 	vins := make([]Vin, 0)
// 	for i, _ := range txvoteinVO.Vin {
// 		vinOne := Vin{
// 			Txid: txvoteinVO.Vin[i].Txid,
// 			Vout: txvoteinVO.Vin[i].Vout,
// 			Puk:  txvoteinVO.Vin[i].Puk,
// 			Sign: txvoteinVO.Vin[i].Sign,
// 		}
// 		vins = append(vins, vinOne)
// 	}

// 	vouts := make([]Vout, 0)
// 	for i, _ := range txvoteinVO.Vout {
// 		voutVO := Vout{
// 			Value:   txvoteinVO.Vout[i].Value,
// 			Address: txvoteinVO.Vout[i].Address,
// 			Tx:      txvoteinVO.Vout[i].Tx,
// 		}
// 		vouts = append(vouts, voutVO)
// 	}

// 	txbase := TxBase{
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

// 	txvin := Tx_deposit_in{
// 		TxBase: txbase,
// 		Puk:    txvoteinVO.Puk,
// 	}
// 	return &txvin, nil
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ConversionVoteInBufVO(txbase Tx_vote_in) ([]byte, error) {
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

// 	txvoteinVO := protobuf.TxVoteInVO{
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
// 		Vote:       txbase.Vote,
// 		VoteType:   uint32(txbase.VoteType),
// 		VoteAddr:   txbase.VoteAddr,
// 	}
// 	return proto.Marshal(&txvoteinVO)
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ParseVoteInBufVO(bs []byte) (*Tx_vote_in, error) {
// 	txvoteinVO := new(protobuf.TxVoteInVO)
// 	err := proto.Unmarshal(bs, txvoteinVO)
// 	if err != nil {
// 		return nil, err
// 	}

// 	vins := make([]Vin, 0)
// 	for i, _ := range txvoteinVO.Vin {
// 		vinOne := Vin{
// 			Txid: txvoteinVO.Vin[i].Txid,
// 			Vout: txvoteinVO.Vin[i].Vout,
// 			Puk:  txvoteinVO.Vin[i].Puk,
// 			Sign: txvoteinVO.Vin[i].Sign,
// 		}
// 		vins = append(vins, vinOne)
// 	}

// 	vouts := make([]Vout, 0)
// 	for i, _ := range txvoteinVO.Vout {
// 		voutVO := Vout{
// 			Value:   txvoteinVO.Vout[i].Value,
// 			Address: txvoteinVO.Vout[i].Address,
// 			Tx:      txvoteinVO.Vout[i].Tx,
// 		}
// 		vouts = append(vouts, voutVO)
// 	}

// 	txbase := TxBase{
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

// 	txvin := Tx_vote_in{
// 		TxBase:   txbase,
// 		Vote:     txvoteinVO.Vote,
// 		VoteType: uint16(txvoteinVO.VoteType),
// 		VoteAddr: txvoteinVO.VoteAddr,
// 	}
// 	return &txvin, nil
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ConversionTokenPublishBufVO(txbase Tx_vote_in) ([]byte, error) {
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

// 	txvoteinVO := protobuf.TxVoteInVO{
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
// 		Vote:       txbase.Vote,
// 		VoteType:   uint32(txbase.VoteType),
// 		VoteAddr:   txbase.VoteAddr,
// 	}
// 	return proto.Marshal(&txvoteinVO)
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ParseTokenPublishBufVO(bs []byte) (*Tx_vote_in, error) {
// 	txvoteinVO := new(protobuf.TxVoteInVO)
// 	err := proto.Unmarshal(bs, txvoteinVO)
// 	if err != nil {
// 		return nil, err
// 	}

// 	vins := make([]Vin, 0)
// 	for i, _ := range txvoteinVO.Vin {
// 		vinOne := Vin{
// 			Txid: txvoteinVO.Vin[i].Txid,
// 			Vout: txvoteinVO.Vin[i].Vout,
// 			Puk:  txvoteinVO.Vin[i].Puk,
// 			Sign: txvoteinVO.Vin[i].Sign,
// 		}
// 		vins = append(vins, vinOne)
// 	}

// 	vouts := make([]Vout, 0)
// 	for i, _ := range txvoteinVO.Vout {
// 		voutVO := Vout{
// 			Value:   txvoteinVO.Vout[i].Value,
// 			Address: txvoteinVO.Vout[i].Address,
// 			Tx:      txvoteinVO.Vout[i].Tx,
// 		}
// 		vouts = append(vouts, voutVO)
// 	}

// 	txbase := TxBase{
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

// 	txvin := Tx_vote_in{
// 		TxBase:   txbase,
// 		Vote:     txvoteinVO.Vote,
// 		VoteType: uint16(txvoteinVO.VoteType),
// 		VoteAddr: txvoteinVO.VoteAddr,
// 	}
// 	return &txvin, nil
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ConversionTokenPaymentBufVO(txbase Tx_vote_in) ([]byte, error) {
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

// 	txvoteinVO := protobuf.TxVoteInVO{
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
// 		Vote:       txbase.Vote,
// 		VoteType:   uint32(txbase.VoteType),
// 		VoteAddr:   txbase.VoteAddr,
// 	}
// 	return proto.Marshal(&txvoteinVO)
// }

// /*
// 	将交易格式化成ProtoBuf字节流，用于持久化
// */
// func ParseTokenPaymentBufVO(bs []byte) (*Tx_vote_in, error) {
// 	txvoteinVO := new(protobuf.TxVoteInVO)
// 	err := proto.Unmarshal(bs, txvoteinVO)
// 	if err != nil {
// 		return nil, err
// 	}

// 	vins := make([]Vin, 0)
// 	for i, _ := range txvoteinVO.Vin {
// 		vinOne := Vin{
// 			Txid: txvoteinVO.Vin[i].Txid,
// 			Vout: txvoteinVO.Vin[i].Vout,
// 			Puk:  txvoteinVO.Vin[i].Puk,
// 			Sign: txvoteinVO.Vin[i].Sign,
// 		}
// 		vins = append(vins, vinOne)
// 	}

// 	vouts := make([]Vout, 0)
// 	for i, _ := range txvoteinVO.Vout {
// 		voutVO := Vout{
// 			Value:   txvoteinVO.Vout[i].Value,
// 			Address: txvoteinVO.Vout[i].Address,
// 			Tx:      txvoteinVO.Vout[i].Tx,
// 		}
// 		vouts = append(vouts, voutVO)
// 	}

// 	txbase := TxBase{
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

// 	txvin := Tx_vote_in{
// 		TxBase:   txbase,
// 		Vote:     txvoteinVO.Vote,
// 		VoteType: uint16(txvoteinVO.VoteType),
// 		VoteAddr: txvoteinVO.VoteAddr,
// 	}
// 	return &txvin, nil
// }
