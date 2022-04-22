package cache

var ch *Cache

func init() {
	ch = newCache()
	ch.init()
}
func To58String(str []byte) string {
	return ch.b58string(str)
}
func ToHex(str []byte) string {
	return ch.hex(str)
}
