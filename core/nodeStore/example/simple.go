package main

import (
	"../../peerNode"
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strconv"
)

func main() {
	peerNode.NodeIdLevel = 5
	peerNode.IsSuper = false
	peerNode.Addr = "127.0.0.1"
	peerNode.Port = map[string]int{"TCP": 0, "UDP": 0}
	StartUP()
}

func read(nodeQueue chan *peerNode.Node, inQueue chan *peerNode.Node) {
	session := 0
	for {
		session++
		node := <-nodeQueue
		// fmt.Println(node.NodeIdShould.String())
		if node.NodeId == nil {
			node.NodeId = node.NodeIdShould
		}
		// node.NodeId = node.NodeIdShould
		node.Addr = strconv.Itoa(session)
		node.Status = 1

		inQueue <- node
	}
}
func StartUP() {
	nodeStore := peerNode.NewNodeStore("", "")
	go read(nodeStore.OutFindNode, nodeStore.InNodes)
	running := true
	reader := bufio.NewReader(os.Stdin)

	for running {
		data, _, _ := reader.ReadLine()
		command := string(data)
		switch command {
		case "find":
			findNode, _ := new(big.Int).SetString("27", 10)
			// peerNode.Print(findNode)
			node := nodeStore.Get(findNode.String())
			// peerNode.Print(node.NodeId)
			fmt.Println(node)
		case "quit":
			running = false
		case "see":
			nodes := nodeStore.GetAllNodes()
			fmt.Println(nodes)
		case "self":
			fmt.Println("自己节点的id：")
			rootId := nodeStore.GetRootId()
			// peerNode.Print(rootId)
			fmt.Println(rootId)
		case "cap":
		case "odp":
		case "cdp":
		case "dump":
		}
	}
}
