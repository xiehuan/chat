package main

import (
	"log"
	"os"
	"os/signal"
	"simpleChat/server/logic"
	"syscall"
)

func main() {
	log.Printf("service begin")
	// 初始化
	service := &logic.Service{}
	service.Start()
	log.Printf("service start ok")

	// 等待终止
	signalKill()

	// 回收
	service.Stop()
	log.Printf("service stop ok")
}

func signalKill() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	select {
	case sig := <-stopChan:
		log.Printf("rev kill signal %v ...", sig)
	}
}
