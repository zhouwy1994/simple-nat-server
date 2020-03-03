package main

import (
	"context"
	"github/zhouwy1994/simple-nat/common"
	server2 "github/zhouwy1994/simple-nat/server"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// wg 负责回收所有Goroutine
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	server := server2.NewServer()
	wg.Add(1)
	server.Start(":3333",":3334",ctx,wg)

	sigterm := make(chan os.Signal, 1)
	// 注册进程可接收的外部信号
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-ctx.Done():
		common.Logger.Println("Context Error")
	case <-sigterm:
		common.Logger.Println("程序接收到退出信号...")
	}

	// 向所有运行的Goroutine发出退出信号
	cancel()

	server.Stop()

	wg.Wait()
}