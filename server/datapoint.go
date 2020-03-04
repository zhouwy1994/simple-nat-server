package server

import (
	"bufio"
	"github/zhouwy1994/simple-nat-server/common"
	"net"
	"sync"
	"time"
)

type InternalDataPoint struct {
	id string // 内部数据端唯一标识
	addPointCb func(* InternalDataPoint) // 内部数据端连接成功回调
	conn *net.Conn // 连接句柄
	once sync.Once // 停止函数只调用一次
}

type ExternalDataPoint struct {
	id string // 外部数据端唯一标识
	conn *net.Conn // 连接句柄
	once sync.Once // 停止函数只调用一次
}

func(dp *InternalDataPoint)Start() {
	r := bufio.NewReader(*dp.conn)
	// 防止DDos
	(*dp.conn).SetReadDeadline(time.Now().Add(2*time.Second))
	head,_,err := r.ReadLine()
	if err != nil {
		common.Logger.Printf("内部数据端点连接，读取消息头失败:%s", err)
		(*dp.conn).Close()
		return
	}

	if string(head) != "dataer" {
		common.Logger.Printf("内部数据端点连接，消息头部格式错误:%s", string(head))
		(*dp.conn).Close()
		return
	}

	body,_,err := r.ReadLine()
	if err != nil || len(body) != 8 {
		common.Logger.Printf("内部数据端点连接，读取消息体失败:%s", err)
		(*dp.conn).Close()
		return
	}

	(*dp.conn).SetReadDeadline(time.Now().Add(0xffff * time.Hour))
	dp.id = string(body)
	dp.addPointCb(dp)
	common.Logger.Printf("内部数据端点连接:%s", dp.id)
}

func(dp *InternalDataPoint)Stop() {
	dp.once.Do(func() {
		if dp.conn != nil {
			defer (*dp.conn).Close()
		}

		if len(dp.id) != 0 {
			common.Logger.Printf("内部数据端点断开:%s", dp.id)
		}
	})
}

func(dp *ExternalDataPoint)Start() {
	common.Logger.Printf("外部数据端点连接:%s", dp.id)
}

func(dp *ExternalDataPoint)Stop() {
	dp.once.Do(func() {
		if dp.conn != nil {
			defer (*dp.conn).Close()
		}

		if len(dp.id) != 0 {
			common.Logger.Printf("外部数据端点断开:%s", dp.id)
		}
	})
}