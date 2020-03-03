package server

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pborman/uuid"
	"github/zhouwy1994/simple-nat/common"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

func NewServer() *Server {
	return &Server{registerListener:&RegisterListener{},dataListener:&DataListener{}}
}

func (s *Server) registerNodeConnect(conn *net.Conn,ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("有注册节点连接:%s\n", (*conn).RemoteAddr())

	(*conn).SetReadDeadline(time.Now().Add(2*time.Second))
	r := bufio.NewReader(*conn)
	data,_,err := r.ReadLine()
	if err != nil || string(data) != "register" {
		return
	}

	client := ClientNode{Conn:conn, Key:uuid.New()[:8]}
	for {
		data,_,err = r.ReadLine()
		if err != nil {
			if os.IsTimeout(err) {
				break
			}

			return
		}

		port,err := strconv.Atoi(string(data))
		if err != nil {
			return
		}

		nodeLister := NodeListener{NodeAddr:strconv.Itoa(port), Key:client.Key}
		nodeLister.Start(s)
		client.NodeList = append(client.NodeList, &nodeLister)
	}

	if len(client.NodeList) > 0 {
		(*client.Conn).SetReadDeadline(time.Now().Add(time.Hour * 2))
		s.registerListener.ClientList = append(s.registerListener.ClientList, &client)
	}
}

func (s *Server) dataNodeConnect(conn *net.Conn,ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	(*conn).SetReadDeadline(time.Now().Add(time.Second * 2))
	r := bufio.NewReader(*conn)
	data,_,err := r.ReadLine()
	if err != nil || string(data) != "dataer" {
		return
	}

	data,_,err = r.ReadLine()
	if len(data) < 8 {
		return
	}

	(*conn).SetReadDeadline(time.Now().Add(time.Hour * 2))
	fmt.Println("客户端输入:", string(data)[:8])
	dataNode := DataNode{Conn:conn, Key:string(data)[:8]}
	s.dataListener.ConnList = append(s.dataListener.ConnList, &dataNode)
	fmt.Printf("有数据节点上来了:%s\n", string(data)[:8])
}

func (s *Server) Start(registerAddr,dataAddr string,ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	rl,err := net.Listen("tcp4", registerAddr)
	if err != nil {
		common.Logger.Println("注册监听器监听失败")
		return err
	}

	dl,err := net.Listen("tcp4", dataAddr)
	if err != nil {
		common.Logger.Println("数据监听器监听失败")
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			conn,err := rl.Accept()
			if err != nil {
				break
			}

			if ctx.Err() != nil {
				break
			}

			wg.Add(1)
			go s.registerNodeConnect(&conn, ctx, wg)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			conn,err := dl.Accept()
			if err != nil {
				break
			}

			if ctx.Err() != nil {
				break
			}

			wg.Add(1)
			go s.dataNodeConnect(&conn, ctx, wg)
		}
	}()

	s.registerListener.ListenAddr = registerAddr
	s.dataListener.ListenAddr = dataAddr
	s.registerListener.Listener = &rl
	s.dataListener.Listener = &dl

	return nil
}

func (s *Server) Stop() {
	(*s.registerListener.Listener).Close()
	for _,client := range s.registerListener.ClientList {
		(*client.Conn).Close()
		for _,node := range client.NodeList {
			(*node.Listener).Close()
			for _,conn := range node.ConnList {
				(*conn.Conn).Close()
			}
		}
	}

	(*s.dataListener.Listener).Close()
	for _,dataer := range s.dataListener.ConnList {
		(*dataer.Conn).Close()
	}
}