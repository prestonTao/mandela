package cache

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func Benchmark_tostring(b *testing.B) {
	rand.Seed(time.Now().Unix())
	for i := 0; i < b.N; i++ {
		i := rand.Int63()
		To58String([]byte("ok" + strconv.Itoa(int(i))))
	}
}
func Benchmark_tohex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ToHex([]byte("ok"))
	}
}
