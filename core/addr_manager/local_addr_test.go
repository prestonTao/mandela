package addr_manager

import (
	// "math/big"
	// "fmt"
	"testing"
)

func TestIdInfo(t *testing.T) {
	// ParseIdInfoExample()
}

func ParseIdInfoExample() {
	idInfo, err := NewIdInfo("tao", "taopopoo@126.com", "sichuan chengdu",
		"dcf2b63eb1e734d3da251e7ed81b6e4cc79c55171626fbfc56b8b2d895905630", "8d0e08b88ccef890")
	if err != nil {
		// fmt.Println(err.Error())
		return
	}
	id := idInfo.Build()
	err = idInfo.Parse(id)
	if err != nil {
		// fmt.Println(err.Error())
		return
	}
	// fmt.Println(idInfo)
}

func BuildIdInfoExample() {
	idInfo, err := NewIdInfo("tao", "taopopoo@126.com", "sichuan chengdu",
		"dcf2b63eb1e734d3da251e7ed81b6e4cc79c55171626fbfc56b8b2d895905630", "8d0e08b88ccef890")
	if err != nil {
		// fmt.Println(err.Error())
		return
	}
	id := idInfo.Build()
	// fmt.Println(string(id), "\n", idInfo.Id, len(idInfo.Id))
}
