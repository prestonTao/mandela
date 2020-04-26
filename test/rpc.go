package main

import (
	"fmt"
	"mandela/rpc"
)

func main() {
	fmt.Println("start...")
	rpc.RegisterRpcServer()
}
