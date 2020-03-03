package server

import (
	"fmt"
	"github.com/pborman/uuid"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func (n *NodeListener) nodeConnect(conn *net.Conn, s *Server) {
	fmt.Printf("有外部连接:%s\n", (*conn).RemoteAddr())

	dataNode := DataNode{Key:uuid.New()[:8] + "-" + n.Key, Conn:conn}
	var cc *net.Conn
	var cs *net.Conn
	for _,client := range s.registerListener.ClientList {
		if client.Key == n.Key {
			cc = client.Conn
		}
	}

	if cc == nil {
		return
	}

	fmt.Println("客户端输出:", dataNode.Key)
	(*cc).Write([]byte(dataNode.Key+"-"+n.NodeAddr+"\r\n"))
	time.Sleep(time.Second)
	for _,node := range s.dataListener.ConnList{
		if node.Key == dataNode.Key[:8] {
			cs = node.Conn
		}
	}

	if cs == nil {
		return
	}

	ncr := io.TeeReader(*conn, *cs)
	ncw := io.TeeReader(*cs,*conn)

	go ioutil.ReadAll(ncr)
	go ioutil.ReadAll(ncw)

	n.ConnList = append(n.ConnList, &dataNode)
	fmt.Printf("建立连接成功:%s-->%s\n", (*cc).RemoteAddr(), (*cs).RemoteAddr())
}

func (n *NodeListener) Start(s *Server) {
	for {
		port := rand.Intn(1109) + 3335
		listener,err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
		if err != nil {
			time.Sleep(time.Millisecond * 50)
			continue
		}

		fmt.Printf("监听端口成功,外网端口:%d,内网端口:%s\n", port, n.NodeAddr)

		n.ExternalAddr = strconv.Itoa(port)
		n.Listener = &listener
		go func() {
			conn,err := listener.Accept()
			if err != nil {
				return
			}

			go n.nodeConnect(&conn, s)
		}()

		break
	}
}
