package service

import (
	"common/message"
)

// func GetDatabaseId(account string) (int, int) {

// 	bs := []byte(account)
// 	firstInt := int(int8(bs[0]))
// 	lenght := int(int8(bs[len(bs)-1]))
// 	idbid := (firstInt + lenght - 1) % 4
// 	itbid := lenght % 4
// 	return idbid, itbid
// }

func GetDatabaseId(accName string) (int, int) {
	Dbid := (int(accName[0]) + int(accName[len(accName)-1])) % message.PLAYER_DATABASE_COUNT
	Tbid := int(accName[len(accName)-1]) % message.PLAYER_DATABASE_COUNT
	return Dbid, Tbid
}
