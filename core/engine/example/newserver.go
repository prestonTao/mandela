package main

import (
	"fmt"
	"net/http"
	"runtime/pprof"
	"time"

	"engine"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("goroutine")
	p.WriteTo(w, 1)
}

func main() {
	go example1()

	http.HandleFunc("/", handler)
	http.ListenAndServe(":11181", nil)
}

var CryKey = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func example1() {
	engine.GlobalInit("console", "", "debug", 1)
	//	engine.GlobalInit("file", "log.txt", "", 1024)

	engine.InitEngine("file_server")
	//	eg := engine.NewEngine("file_server")
	engine.RegisterMsg(101, hello)

	// engine.SetAuth(new(handlers.NoneAuth))
	// engine.SetCloseCallback(handlers.CloseConnHook)
	err := engine.Listen("0.0.0.0", 9981)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Hour * 1000000)
}

func hello(c engine.Controller, msg engine.Packet) {
	fmt.Println(string(msg.Data), "sever")
	data := []byte("hello")
	msg.Session.Send(101, 0, 0, &data)
}

type FindNode struct {
	Name string `json:"name"`
}
