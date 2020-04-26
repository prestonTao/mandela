package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/utils/crypto"
)

func FindWitnessName(addr crypto.AddressCoin) string {
	value, err := db.Find(append([]byte(config.WitnessAddr), addr...))
	if err != nil {
		return ""
	}
	return string(*value)
}

func FindWitnessAddr(name string) *crypto.AddressCoin {
	value, err := db.Find(append([]byte(config.WitnessName), []byte(name)...))
	if err != nil {
		return nil
	}
	addr := crypto.AddressCoin(*value)
	return &addr
}

func SaveWitnessName(addr crypto.AddressCoin, name string) {
	bs := []byte(name)
	addrBs := []byte(addr)
	db.Save(append([]byte(config.WitnessAddr), addr...), &bs)
	db.Save(append([]byte(config.WitnessName), bs...), &addrBs)
}
func DelWitnessName(name string) {
	addr := FindWitnessAddr(name)
	db.Remove(append([]byte(config.WitnessName), []byte(name)...))
	if addr != nil {
		db.Remove(append([]byte(config.WitnessAddr), *addr...))
	}
}
