package pubstore

//获取公共keystore
func GetPubStore(pwd, seed string) (*Keystore, error) {
	keystore := NewKeystore()
	ss, err := Decode(seed)
	if err != nil {
		return nil, err
	}
	err = keystore.SeedtoFile(pwd, ss)
	if err != nil {
		return nil, err
	}
	return keystore, nil
}
