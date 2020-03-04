package server

import (
	"github.com/google/uuid"
	"github/zhouwy1994/simple-nat-server/common"
	"net"
	"sync"
)

// PortMapper 端口映射器结构
type PortMapper struct {
	id string // 端口映射器唯一标识
	cid string // 客户端ID
	srcPort int // 外网端口，映射到外网的端口
	dstPort int // 目的端口，客户端需要映射到外网的端口
	client *NatClient // 所属客户端句柄
	listener *net.Listener // 端口映射监听器
	pairs []*ConnectPair // 连接Pair
	locker sync.Mutex // 线程安全锁
}

func (p *PortMapper) Start() {
	p.listener,p.srcPort = common.GetAvailableListener()

	common.Logger.Printf("端口映射成功：外网端口(%d)<=====>内部端口(%d)", p.srcPort, p.dstPort)

	go p.startAccept()
}

func (p *PortMapper) Stop() {
	defer p.locker.Unlock()
	p.locker.Lock()

	if p.listener != nil {
		defer (*p.listener).Close()
	}

	for _,pair := range p.pairs {
		pair.Stop()
	}

	p.pairs = p.pairs[0:0]
}

func (p *PortMapper) startAccept() {
	for {
		conn,err := (*p.listener).Accept()
		if err != nil {
			break
		}

		edp := ExternalDataPoint{id:uuid.New().String()[:8], conn:&conn}
		edp.Start()

		idp := p.client.notifyAndGetConnect(edp.id, p.dstPort)
		if idp == nil {
			common.Logger.Printf("获取内部端点连接失败")
			edp.Stop()
			continue
		}

		pair := ConnectPair{externalDP:&edp, internalDP:idp, exceptionCb:p.deleteConnectPair}
		p.addConnectPair(&pair)
		pair.Start()
	}
}

func (p *PortMapper) addConnectPair(cp *ConnectPair) {
	defer p.locker.Unlock()
	p.locker.Lock()
	p.pairs = append(p.pairs, cp)
	common.Logger.Printf("%s连接对数量:%d", p.id, len(p.pairs))

}

func (p *PortMapper) deleteConnectPair(cp *ConnectPair) {
	defer p.locker.Unlock()
	p.locker.Lock()

	for i,c :=range p.pairs{
		if c == cp {
			c.Stop()
			// 此处有待解耦合
			p.client.deleteDataPoint(c.internalDP)
			p.pairs = append(p.pairs[:i], p.pairs[i+1:]...)
		}
	}

	common.Logger.Printf("%s连接对数量:%d", p.id, len(p.pairs))
}
