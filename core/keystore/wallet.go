package keystore

import (
	"mandela/config"
	"mandela/core/utils/crypto"
	"mandela/core/utils/crypto/dh"
	"bytes"
	"crypto/aes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/ed25519"
)

type Wallet struct {
	Key       []byte        `json:"key"`       //生成主密钥的随机数
	ChainCode []byte        `json:"chaincode"` //主KDF链编码
	IV        []byte        `json:"iv"`        //aes加密向量
	CheckHash []byte        `json:"checkhash"` //主私钥和链编码加密验证hash值
	Coinbase  uint64        `json:"coinbase"`  //当前默认使用的收付款地址
	Addrs     []AddressInfo `json:"addrs"`     //已经生成的地址列表
	DHKey     []DHKeyPair   `json:"dhkey"`     //DH密钥
	lock      *sync.RWMutex //
	// Puks      map[crypto.AddressCoin]ed25519.PublicKey `json:"puks"`      //存放地址和公钥；key:crypto.AddressCoin=地址;value:ed25519.PublicKey=公钥;
}

type AddressInfo struct {
	Index uint64             `json:"index"` //棘轮数量
	Addr  crypto.AddressCoin `json:"addr"`  //收款地址
	Puk   ed25519.PublicKey  `json:"puk"`   //公钥
}

type DHKeyPair struct {
	Index   uint64     `json:"index"`   //棘轮数量
	KeyPair dh.KeyPair `json:"keypair"` //
}

/*
	检查钱包是否完整
*/
func (this *Wallet) CheckIntact() bool {
	if this.Key == nil || this.ChainCode == nil || this.IV == nil || this.CheckHash == nil {
		return false
	}
	if len(this.Key) != 48 || len(this.ChainCode) != 48 || len(this.IV) != aes.BlockSize || len(this.CheckHash) != 64 {
		return false
	}
	if len(this.Addrs) <= 0 {
		return false
	}
	return true
}

/*
	获取地址列表
*/
func (this *Wallet) GetAddr() (addrs []crypto.AddressCoin) {
	addrs = make([]crypto.AddressCoin, 0)
	this.lock.RLock()
	for _, one := range this.Addrs {
		addrs = append(addrs, one.Addr)
	}
	this.lock.RUnlock()
	return
}

/*
	生成一个新的地址，需要密码
*/
func (this *Wallet) GetNewAddr(password [32]byte) (crypto.AddressCoin, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	ok, key, code, err := this.decrypt(password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, config.ERROR_password_fail
	}
	//密码验证通过

	//查找用过的最高的棘轮数量
	addrIndex := uint64(0)
	if len(this.Addrs) > 0 {
		addrInfo := this.Addrs[len(this.Addrs)-1]
		addrIndex = addrInfo.Index
	}
	dhIndex := uint64(0)
	if len(this.DHKey) > 0 {
		dhKey := this.DHKey[len(this.DHKey)-1]
		// if dhIndex < dhKey.Index {
		// }
		dhIndex = dhKey.Index
	}
	index := addrIndex
	if index < dhIndex {
		index = dhIndex
	}
	index = index + 1

	//生成新的地址
	key, _, err = crypto.GetHkdfChainCode(key, code, index)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(key)
	puk, _, err := ed25519.GenerateKey(buf)
	if err != nil {
		return nil, err
	}
	addr := crypto.BuildAddr(config.AddrPre, puk)

	addrInfo := AddressInfo{
		Index: index, //棘轮数
		Addr:  addr,  //收款地址
		Puk:   puk,   //公钥
	}
	this.Addrs = append(this.Addrs, addrInfo)
	return addr, nil
}

/*
	生成一个新的地址，需要密码
*/
func (this *Wallet) GetNewDHKey(password [32]byte) (*dh.KeyPair, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	ok, key, code, err := this.decrypt(password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("password is fail!")
	}

	//查找用过的最高的棘轮数量
	index := uint64(0)
	if len(this.Addrs) > 0 {
		addrInfo := this.Addrs[len(this.Addrs)-1]
		index = addrInfo.Index
	}
	if len(this.DHKey) > 0 {
		dhKey := this.DHKey[len(this.DHKey)-1]
		if index < dhKey.Index {
			index = dhKey.Index
		}
	}
	index = index + 1

	//密码验证通过，生成新的地址
	key, _, err = crypto.GetHkdfChainCode(key, code, index)
	if err != nil {
		return nil, err
	}
	keyPair, err := dh.GenerateKeyPair(key)
	if err != nil {
		return nil, err
	}
	dhKey := DHKeyPair{
		Index:   index,
		KeyPair: keyPair,
	}
	this.DHKey = append(this.DHKey, dhKey)
	return &keyPair, nil
}

/*
	设置默认收付款地址
*/
func (this *Wallet) SetCoinbase(index uint64) bool {
	if index < uint64(len(this.Addrs)) {
		this.Coinbase = uint64(index)
		return true
	}
	return false
}

/*
	设置默认收付款地址
*/
func (this *Wallet) GetCoinbase() crypto.AddressCoin {
	return this.Addrs[this.Coinbase].Addr
}

func (this *Wallet) GetDHbase() DHKeyPair {
	return this.DHKey[len(this.DHKey)-1]
}

/*
	设置DH密钥位置
*/
// func (this *Wallet) SetDHIndex(index uint64) bool {
// 	if index < uint64(len(this.Addrs)) {
// 		this.Coinbase = index
// 		return true
// 	}
// 	return false
// }

// /*
// 	设置DH密钥位置
// */
// func (this *Wallet) getDHIndex() bool {
// 	return this.Addrs[this.Coinbase].Addr
// }

/*
	使用密码解密种子，获得私钥和链编码
	@return    ok    bool    密码是否正确
	@return    key   []byte  生成私钥的随机数
	@return    code  []byte  链编码
*/
func (this *Wallet) decrypt(pwdbs [32]byte) (ok bool, key, code []byte, err error) {
	//密码取hash
	// pwdbs := sha256.Sum256(password)

	//先用密码解密key和链编码
	keyBs, err := crypto.DecryptCBC(this.Key, pwdbs[:], this.IV)
	if err != nil {
		return false, nil, nil, config.ERROR_password_fail
	}
	codeBs, err := crypto.DecryptCBC(this.ChainCode, pwdbs[:], this.IV)
	if err != nil {
		return false, nil, nil, config.ERROR_password_fail
	}

	//验证密码是否正确
	checkHash := append(keyBs, codeBs...)
	h := sha256.New()
	n, err := h.Write(checkHash)
	if n != len(checkHash) {
		//hash 写入失败
		return false, nil, nil, errors.New("hash Write failure")
	}
	if err != nil {
		return false, nil, nil, err
	}
	checkHash = h.Sum(pwdbs[:])
	// checkHash = sha256.Sum256(checkHash)[:]
	if !bytes.Equal(checkHash, this.CheckHash) {
		return false, nil, nil, nil
	}
	return true, keyBs, codeBs, nil
}

/*
	生成一个新的地址，需要密码
*/
// func (this *Wallet) Encrypt(password string) (crypto.Address, error) {

// }

/*
	查询地址，判断地址是否在本钱包中
*/
func (this *Wallet) FindAddress(addr crypto.AddressCoin) (ok bool) {
	ok = false
	this.lock.RLock()
	for _, one := range this.Addrs {
		if bytes.Equal(one.Addr, addr) {
			ok = true
			break
		}
	}
	this.lock.RUnlock()
	return
}

/*
	通过地址获取密钥
	@rand    []byte    hkdf链生成的随机数
*/
func (this *Wallet) GetKeyByAddr(addr crypto.AddressCoin, pwd [32]byte) (rand []byte, prk ed25519.PrivateKey, puk ed25519.PublicKey, err error) {
	ok, key, code, err := this.decrypt(pwd)
	if err != nil {
		return nil, nil, nil, err
	}
	if !ok {
		return nil, nil, nil, errors.New("Incorrect password!")
	}

	// fmt.Println("解密后的key和code", len(key), len(code))

	err = errors.New("address not found")
	this.lock.RLock()
	for _, one := range this.Addrs {
		if bytes.Equal(addr, one.Addr) {
			rand, _, err = crypto.GetHkdfChainCode(key, code, one.Index)
			if err != nil {
				return nil, nil, nil, err
			}
			// fmt.Println("随机数长度2", len(rand))
			puk, prk, err = ed25519.GenerateKey(bytes.NewBuffer(rand))
			break
		}
	}
	this.lock.RUnlock()
	return
}

/*
	通过地址获取密钥
	@rand    []byte    hkdf链生成的随机数
*/
func (this *Wallet) GetPukByAddr(addr crypto.AddressCoin) (puk ed25519.PublicKey, ok bool) {
	ok = false
	this.lock.RLock()
	// fmt.Println("寻找的地址", addr.B58String())
	for _, one := range this.Addrs {
		// fmt.Println("第", i, "个地址", one.Addr.B58String(), one.Index, hex.EncodeToString(one.Puk))
		if bytes.Equal(addr, one.Addr) {
			puk = one.Puk
			ok = true
			break
		}
	}
	this.lock.RUnlock()
	return
}

/*
	修改密码
*/
func (this *Wallet) UpdatePwd(oldpwd, newpwd [32]byte) (ok bool, err error) {
	fmt.Println(oldpwd, newpwd)
	ok = false
	ok, key, code, err := this.decrypt(oldpwd)
	if err != nil {
		return false, err
	}

	iv, err := crypto.Rand16Byte()
	if err != nil {
		return false, err
	}

	keySec, err := crypto.EncryptCBC(key[:], newpwd[:], iv[:])
	if err != nil {
		return false, err
	}
	codeSec, err := crypto.EncryptCBC(code[:], newpwd[:], iv[:])
	if err != nil {
		return false, err
	}

	hash := sha256.New()
	hash.Write(append(key[:], code[:]...))
	checkHash := hash.Sum(newpwd[:])

	this.Key = keySec
	this.ChainCode = codeSec
	this.CheckHash = checkHash
	this.IV = iv[:]

	return true, nil
}

/*
	创建一个新的钱包种子
*/
func NewWallet(key, code [32]byte, iv [16]byte, pwd [32]byte) (*Wallet, error) {
	// fmt.Println("key:", key, "\ncode:", code, "\npwd:", pwd)

	keySec, err := crypto.EncryptCBC(key[:], pwd[:], iv[:])
	if err != nil {
		return nil, err
	}
	codeSec, err := crypto.EncryptCBC(code[:], pwd[:], iv[:])
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	hash.Write(append(key[:], code[:]...))
	checkHash := hash.Sum(pwd[:])

	wallet := Wallet{
		Key:       keySec,                 //生成主密钥的随机数
		ChainCode: codeSec,                //主KDF链编码
		IV:        iv[:],                  //aes加密向量
		CheckHash: checkHash,              //主私钥和链编码加密验证hash值
		Addrs:     make([]AddressInfo, 0), //已经生成的地址列表
		Coinbase:  0,                      //当前默认使用的收付款地址
		DHKey:     make([]DHKeyPair, 0),   //dh密钥对
		lock:      new(sync.RWMutex),      //
	}
	//生成第一个地址
	wallet.GetNewAddr(pwd)
	// buf := bytes.NewBuffer(key[:])
	// puk, _, err := ed25519.GenerateKey(buf)
	// if err != nil {
	// 	return nil, err
	// }
	// addr := crypto.BuildAddr(config.NetId, puk)

	// addrInfo := AddressInfo{
	// 	Index: 0,    //
	// 	Addr:  addr, //收款地址
	// 	Puk:   puk,  //公钥
	// }
	// wallet.Addrs = append(wallet.Addrs, addrInfo)

	//生成第一个DH密钥对
	wallet.GetNewDHKey(pwd)

	return &wallet, nil
}
