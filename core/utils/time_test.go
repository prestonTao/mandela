package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeToken(*testing.T) {
	class1 := "1"
	class2 := "2"
	SetTimeToken(class1, time.Second)
	SetTimeToken(class1, time.Second)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Second / 2)
		allow1 := GetTimeToken(class1)
		allow2 := GetTimeToken(class2)
		fmt.Println(allow1, allow2)
	}
}
