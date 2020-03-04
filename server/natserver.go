package server

import (
	"github.com/google/uuid"
	"github/zhouwy1994/simple-nat-server/common"
	"net"
	"sync"
)

// NatServer NAT服务器结构
type NatServer struct {
	listener *net.Listener // NAT客户端注册监听器
	clients []*NatClient // 在线NAT客户端集合
	locker sync.Mutex // 线程安全锁
}

// NewServer 获取NAT服务器实例
func NewServer() *NatServer {
	return &NatServer{}
}

// Start 启动NAT服务器，开始监听客户端注册服务
func (s *NatServer) Start(srvAddr string) error {
	listener,err := net.Listen("tcp4", srvAddr)
	if err != nil {
		return err
	}

	go s.startAccept()

	s.listener = &listener

	common.Logger.Printf("服务启动，监听地址:%s", srvAddr)

	return nil
}

// startAccept 启动NAT客户端注册来连接
func (s *NatServer) startAccept() {
	for {
		conn,err := (*s.listener).Accept()
		if err != nil {
			break
		}

		c := &NatClient{id:uuid.New().String()[:8],conn:&conn, exceptionCb:s.deleteClient}
		s.addClient(c)
		c.Start()
	}
}

func (s *NatServer) addClient(c *NatClient) {
	defer s.locker.Unlock()
	s.locker.Lock()
	s.clients = append(s.clients, c)

	common.Logger.Printf("客户端连接池剩余连接数:%d", len(s.clients))
}

func (s *NatServer) deleteClient(c *NatClient) {
	defer s.locker.Unlock()
	s.locker.Lock()

	for i,client :=range s.clients{
		if client == c {
			c.Stop()
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
		}
	}

	common.Logger.Printf("客户端连接池剩余连接数:%d", len(s.clients))
}

func (s *NatServer) Stop() {
	defer s.locker.Unlock()
	s.locker.Lock()

	if s.listener != nil {
		defer (*s.listener).Close()
	}

	for _,c := range s.clients {
		c.Stop()
	}

	s.clients = s.clients[0:0]

	common.Logger.Printf("服务停止")
}