package main

import (
	mc "mandela/core/message_center"
	"mandela/core/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mr-tron/base58/base58"
)

func main() {
	//	example3()
	now := time.Now()
	_, offset := now.Zone()
	now.Unix() - offset

	fmt.Println(now.Zone())
	fmt.Println(now.In(time.Local).String())
	fmt.Println(now.String())
	fmt.Println(now.Unix())

	fmt.Println(now.UTC().Unix())
	fmt.Println(now.Local().Unix())
	// e5()
}

func e5() {
	a := "BTC"

	b := base58.Encode([]byte(a)) // base64.StdEncoding.EncodeToString([]byte(a))

	fmt.Println(":", []byte(a), b)

	c, _ := base58.Decode(a)
	fmt.Println(":", string(c))

	for i := 0; i <= 255; i++ {

		x := base58.Encode([]byte{byte(i)})
		fmt.Println(i, x)
	}
}

func e4() {
	str := `{"c_time":"2018-03-12 18:04:01.0333927","r_time":"2018-03-12 18:04:04.3865845","hash":"ERTVyjD8uhymjpnoeghIRXNmzxBzjw==","r_hash":"ERTg8gj0OWxOB7CHnXr577RXdG5RhQ==","s_rand":981,"r_rand":1115,"content":"eyJGaWxlSGFzaCI6IkVSVG1IdVNpcHNCb1g0VXlaenpEY2p2QzFiOUNhdz09IiwiTm8iOjk0MSwiQ2h1bmtIYXNoIjoiYVF2OEF1TjVUY1IrWStZbW5ucFBqWGhKaVowPSIsIkluZGV4IjowLCJMZW5ndGgiOjk5OTk5OTk5OTk5LCJDb250ZW50IjoiWHJHQjFxQloyN3VNU0VqMHhSMW1qWFRoTStvTzlVYTV0d05JaVVObnVEUkRrd1BMaVUrV1FoVDZBbit0RWowTHNyWTNHL1ROVWxhTVowcG5QUUNoUlNPc21lbTNhcm9jbWJLRzU1WS8xbks0M0dvMWozZ2pBeEpMdDA3MGlDekZqcDZETEFlbFVSc2RkaWZTb0xIdjRzeXBzaWtnSHFvemo4cW8zRWJmbTZwTFdQQjcvd0RqYWxKbVZqZ1k5Q1Bhb1ptejFPMytDcURZZnhLM01iQmJZSjA4eXNSL1dsSnVLTGhsU1NWVG5ZK2xKNUxNVjA0ejBPTjhWUjRkOHJqQjZqdlNrS0xweGJpSURmZkQweWNuYnQzcVY0dmVyczNuK2hOQ0VVeXFHV05pM1VZcUpGa1gvd0NhNmdiRWIxT0tNSEZoSk9LM1E4NFFzT3VNNG9FdkhycFZaY0VJT2gxMUFFemJhQ3ZtMjk2RGRoeXVTbzIvR3B3TmJpeGFYeERjYXQzSXlmTGpmRlYvOXkzYTRDTVNEMTgzV2tiMjFkdW1Cdld1bHRKUVRwemo5NnJnYVdtamZyNGx1MWJEU2JEWWpQU3Ivd0R1bWNLVkRmaG5hdVJibkIvTUNQV2hpUTRZOSt3ejFweG94VW1kaWZFc3VrNG1ZWkhRSEZCZnhMZGtqWEpuQjZESUZjcnpOUzUzOXhVYzZuRW5KMmRTZkUxeVdHRzZiNUJPeC9Hc244Unp5SzRhVTRkU0NkVzNUcFhKYzk4Qmd0UUxpWE9UMDdlMVpxQmViTlhjS3ZOa0F5Q0hQNzB0cDAvTXY0MDdjb3haM3hnVXRxYlNSMVgwcnBpOUd5UDVJVlprSUoyMnFOSVlFazBXUlZZWng5TVVQUnQ4Mkt5TXFCc20rY1ZnWHBWaHE5UVJWZFp4OHZRMURFenZtc3djZWJGU1pGMjJJK29xZVl1bkRZK3RHV2lta05uYmVoU1JnN2c0cG9LcEd6VkJUUHI3VUxXaEl4dG4rbFlVTk9CTjZycHdRS1dZMEtCVC90M3JQWHkwM2hlbWFyb1gyRkxSZUFzUHhyUE5rMDN5MXhVY3NkZS9hbG9qZ3hYMXhVcmtHbWhHUCthamxERkxLb2V3QWIxRlFmMHBrd3J0dFVDQVo5UlVIRVd6V2RQTlRKZ3lSOWF6a0Nsb2NiRjk5dlNvelRBaHJPUnZSRTRzV3puNlZQWUNtT1R2aXM1UFluYjYxYlJLRjFPOVQzMjZVd0ljVkt3TG1zYk11TFlzYy9oVmdXQjlzVXdJVjlUbnQ3MWZrcGoxOWFXT0xGQUNSaXJxcHozcHBVUWZTbzA0YmJwU3k4UUFqL09wQ0FHajZSdDF6V0JSOUtsbVZBdVdweDc5NkdZVzdZcGtqSTcxSVVVVExRbVkyR1BMVUtyRHFEVCtrZDZvRlU1RzJjZHF2SXg0aWZteDdWSTlNVTFwN0FmaldDUGFwWVVSUXFjNzVGU0l5UFhGTTZmWTFaVjN5S2NpOFJZUnR2V0NNNEhscHZUdGtDcXNwQTZiK2xHeXBBZVVjYmtVUVI3ZTFFd1Bhc0hsSU9CanRtcFpseEtKSDdDcklvOWdhdG5xZTNTc0REVDAzOWFQb0pJc1ZYb0J2aXJLb0hyaW9UY2lpSHlqSHZVTXRGQnZnanBSRlhmSFNzUWdMdjB6UlZDbmNkL1dvQm5oa0t6WGNFTFpJZVJWYkhvVFgwUHdxMDRmWjhPanQ0NFVPQU1uSHRYaFhoZTNadUt3U3RrSkd3WTE2dEJ4WlZpODJRQUIyMnJYTXF1emV2WUhtQT09In0="}`

	body := new(mc.MessageBody)
	err := json.Unmarshal([]byte(str), body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("end")
}

func example3() {
	m := new(sync.Map)
	m.Store("tao", "hongfei")
	_, ok := m.Load("tao")
	fmt.Println(ok)
}

func example2() {
	tm := utils.NewBackoffTimer(1, 2, 4, 8, 16)
	for {
		n := tm.Wait()
		fmt.Println(n)
	}
}

func example1() {
	str := "c4037266f637f017a15b8428fae46075503f9f27"
	bs, err := hex.DecodeString(str)
	if err != nil {
		fmt.Println("格式化失败1", err)
	}

	for range utils.Names {
		//		bs, err := utils.EncodeName(bs, key)
		//		if err != nil {
		//			fmt.Println("格式化失败2", err)
		//			continue
		//		}
		base58.Encode(bs)
		//		utils.Multihash(bs).B58String()
		//		names := utils.Multihash(bs).B58String()
		fmt.Println("转化后", base58.Encode(bs))
	}

}
