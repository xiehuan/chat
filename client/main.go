package main

import (
	"log"
	"simpleChat/client/logic"
)

func main() {

	client := logic.NewClient()
	client.CreateConn()
	log.Printf("client conn server ok")

	// 监听标准输入
	client.ReadStdin()
}
