package main

import (
	"fmt"
	"mandela/wallet/keystore"
)

func main() {
	//L4mwX5c9GZ85i32bWQgoT8i79YHC5eiKdq9UDGR7PjTYYKSKH2ox
	//KwZfiVztGNBMVnDNXv6y4q52mQkNcRMwZpeyHgLSTDpCzuh177pc
	//L4gMe9Td54WBuAuAkvcZKZAMKMYu1TJvYyEskzFoGwCtq3oGjv7R
	keystore.NewLoad("L4gMe9Td54WBuAuAkvcZKZAMKMYu1TJvYyEskzFoGwCtq3oGjv7R", "123456")
	fmt.Println("end")
}
