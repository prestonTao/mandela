package utils

import (
	"encoding/binary"
)

/*
	构建默克尔树根
*/
func BuildMerkleRoot(tx [][]byte) []byte {
	if len(tx) == 0 {
		return []byte{}
	}

	if len(tx) == 1 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, 1)
		return Hash_SHA3_256(append(b, append(tx[0], tx[0]...)...))
	}

	txbs := merkleroot(0, tx)
	return txbs[0]
}

/*
	计算默克尔树根
*/
func merkleroot(level uint64, tx [][]byte) [][]byte {
	//	fmt.Println("计算默克尔树", len(tx))
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, level)
	//	fmt.Println("计算默克尔树", b)
	if len(tx) == 1 {
		return [][]byte{append(b, tx[0]...)}
	}

	newtx := make([][]byte, 0)
	for i := 0; i < len(tx)/2; i++ {
		newtx = append(newtx, Hash_SHA3_256(append(b, append(tx[i*2], tx[((i+1)*2)-1]...)...)))
	}
	if len(tx)%2 != 0 {
		newtx = append(newtx, Hash_SHA3_256(append(b, append(tx[0], tx[len(tx)-1]...)...)))
	}
	return merkleroot(level+1, newtx)
}
