package main

import (
	// _ "./cxv"
	server "gameServerEngine"
)

func main() {
	server.ConfigPath = "config/accServer.ini"
	server.StartUP()
}
