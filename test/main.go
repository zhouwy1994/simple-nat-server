package main

import (
	"github/zhouwy1994/simple-nat-server/common"
	server2 "github/zhouwy1994/simple-nat-server/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// wg 负责回收所有Goroutine
	server := server2.NewServer()
	server.Start(":3333")

	sigterm := make(chan os.Signal, 1)
	// 注册进程可接收的外部信号
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-sigterm:
		common.Logger.Println("程序接收到退出信号...")
	}

	server.Stop()
}