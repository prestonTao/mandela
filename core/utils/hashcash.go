package utils

import (
	"crypto/sha256"
	"math/big"
	"strconv"
)

/*
	引用至：
		http://www.ibm.com/developerworks/cn/linux/l-hashcash.html
	协议格式
	1:bits:date:resource:ext:salt:suffix
	例如：
		1:20:140703:taopopoo::tHODRlCK:1c45c
	戳记包括 7 个域。
	1. 版本号（版本 0 更简单，但是有一些局限性）。
	2. 声明的比特值。如果戳记没有真正地使用声明的前导零比特进行散列，那么它就是非法的。
	3. 生成戳记的日期（和时间）。可以认为当前时间之后的戳记以及那些在很久以前的戳记是非法的。
	4. 戳记为哪个资源而生成。可能是一个电子邮件地址，但是也可能是一个 URI 或者其他命名的资源。
	5. 特定应用程序可能需要的扩展。任何附加的数据都可以放置在这里，但是，在到目前为止的使用中， 这个域通常是空的。
	6. 将该戳记与其他所有人为相同的资源在同一日期生成的戳记区别开来的随机因子（salt）。例如，两个不同的人
		可以合情合理地在同一天向我的同一个地址发送电子邮件。他们不应该由于我使用了 double spend 数据库而无法发送成功。
		但是，如果他们每个人都使用一个随机因子，那么完整戳记将是不同的。
	7. 后缀是算法真正起作用的部分。假定给出了前 6 个域，为了生成一个通过期望数目的前导零 进行散列的的戳记，
		minter必须尝试很多连续的后缀值。
*/

//开始工作，猜解一段字符串hash后有足够个0
//目前（2014年）一般计算机计算出20个0大约一百万次，用时3.0121722秒
func Work(message string, zeroes int) int {
	nonce := 0
	for {
		nonce++
		digest := sha256.Sum256([]byte(message + strconv.Itoa(nonce)))
		digestHex := new(big.Int).SetBytes(digest[:])
		if digestHex.BitLen() == 256-zeroes {
			return nonce
		}
	}
}

//检查是否完成工作
func Check(message string, zeroes int, nonce int) bool {
	digest := sha256.Sum256([]byte(message + strconv.Itoa(nonce)))
	for i := zeroes; i >= 0; i-- {
		if digest[i] != 0 {
			return false
		}
	}
	return true
}

/*
	检查hash值是否有足够多的前导零
	@code      string    Hash值
	@zeroes    uint64    前导零个数
*/
func CheckNonce(code []byte, zeroes uint64) bool {
	digestHex := new(big.Int).SetBytes(code)
	surplus := new(big.Int).Rsh(digestHex, uint(256-zeroes))
	zero := big.NewInt(0)
	result := zero.Cmp(surplus)
	if result == 0 {
		return true
	}
	return false
}
