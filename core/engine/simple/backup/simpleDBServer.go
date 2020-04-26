package main

import (
	// _ "./cxv"
	server "gameServerEngine"
)

func main() {
	server.ConfigPath = "config/dbServer.ini"
	server.StartUP()
}
