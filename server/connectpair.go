package server

import (
	"io"
	"io/ioutil"
)

type ConnectPair struct {
	internalDP *InternalDataPoint // 内部连接，与客户端导出端口的连接
	externalDP *ExternalDataPoint // 外部连接，与映射到外网的端口的连接
	exceptionCb func(* ConnectPair) // 异常回调函数
}

func(p *ConnectPair)Start() {
	er := io.TeeReader(*p.externalDP.conn, *p.internalDP.conn)
	ir := io.TeeReader(*p.internalDP.conn, *p.externalDP.conn)

	go func() {
		ioutil.ReadAll(er)
		p.exceptionCb(p)
	}()

	go func() {
		ioutil.ReadAll(ir)
		p.exceptionCb(p)
	}()
}

func(p *ConnectPair)Stop() {
	if p.internalDP != nil {
		p.internalDP.Stop()
	}

	if p.externalDP != nil {
		p.externalDP.Stop()
	}
}

