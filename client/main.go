package main

import (
	"log"
	"simpleChat/client/logic"
)

func main() {
	// 创建客户端
	client := logic.NewClient()
	client.CreateConn()
	log.Printf("client conn server ok")

	// 监听标准输入
	client.ReadStdin()
}
