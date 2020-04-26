package cache

var (
	Caches *Cache
)

func initCache() {
	Caches = new(Cache)
}
func Save(key string, value []byte) {
	Caches.Save(key, value)
}
func Get(key string) []byte {
	return Caches.Get(key)
}
