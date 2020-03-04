package server

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"github/zhouwy1994/simple-nat-server/common"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// NatClient NAT客户端结构
type NatClient struct {
	id string // 客户端唯一标识
	conn *net.Conn // 连接句柄
	listener *net.Listener // 数据端口监听器
	listPort int // 数据监听端口
	exceptionCb func(*NatClient) // 异常回调函数
	ports []*PortMapper // 端口映射器集合
	dps []*InternalDataPoint // 内部数据端点集合
	locker sync.Mutex // 线程安全锁
	connNotifyMp map[string]chan bool //  内部数据端点连接通知
}

func (c *NatClient) Start() {
	common.Logger.Printf("客户端连接:%s", (*c.conn).RemoteAddr())

	r := bufio.NewReader(*c.conn)
	// 防止DDos
	(*c.conn).SetReadDeadline(time.Now().Add(1*time.Second))
	head,_,err := r.ReadLine()
	if err != nil {
		common.Logger.Printf("客户端注册，读取消息头失败:%s", err)
		c.exceptionCb(c)
		return
	}

	if string(head) != "register" {
		common.Logger.Printf("客户端注册，消息头部格式错误:%s", string(head))
		c.exceptionCb(c)
		return
	}

	for {
		body,_,err := r.ReadLine()
		if err != nil {
			if os.IsTimeout(err) {
				break
			}
			common.Logger.Printf("客户端注册，读取消息体失败:%s", err)
			c.exceptionCb(c)
			return
		}

		portNum,err := strconv.Atoi(string(body))
		if err != nil || portNum < 0 || portNum > 65525 {
			common.Logger.Printf("客户端注册，错误的映射端口:%s", string(body))
			c.exceptionCb(c)
			return
		}
		p := PortMapper{dstPort:portNum,id:uuid.New().String()[:8],cid:c.id, client:c}
		c.ports = append(c.ports, &p)
	}

	if len(c.ports) < 1 {
		common.Logger.Printf("客户端注册，映射端口为空！！！")
		c.exceptionCb(c)
		return
	}

	c.listener,c.listPort = common.GetAvailableListener()
	go c.startAccept()

	// 停止客户端端口映射信息
	portMapInfo := "notify\r\n"
	for _,p := range c.ports {
		p.Start()
		portMapInfo += fmt.Sprintf("%d-%d|", p.srcPort, p.dstPort)
	}

	portMapInfo+="\r\n"

	(*c.conn).Write([]byte(portMapInfo))

	c.connNotifyMp = make(map[string]chan bool)

	// 检测连接状态
	go c.connStatusCheck()
}

func (c *NatClient) addDataPoint(p *InternalDataPoint) {
	defer c.locker.Unlock()
	c.locker.Lock()
	c.dps = append(c.dps, p)
	if ch,ok := c.connNotifyMp[p.id];ok {
		ch <- true
	}

	common.Logger.Printf("内部端点连接池剩余连接数:%d", len(c.dps))
}

func (c *NatClient) deleteDataPoint(p *InternalDataPoint) {
	defer c.locker.Unlock()
	c.locker.Lock()

	for i,point :=range c.dps{
		if point == p {
			point.Stop()
			c.dps = append(c.dps[:i], c.dps[i+1:]...)
		}
	}

	common.Logger.Printf("内部端点连接池剩余连接数:%d", len(c.dps))
}

func (c *NatClient) startAccept() {
	for {
		conn,err := (*c.listener).Accept()
		if err != nil {
			break
		}

		dp := InternalDataPoint{conn:&conn, addPointCb:c.addDataPoint}
		dp.Start()
	}
}

func (c *NatClient) connStatusCheck() {
	(*c.conn).SetReadDeadline(time.Now().Add(time.Hour * 0xffff))
	ioutil.ReadAll(*c.conn)
	c.exceptionCb(c)
}

func (c *NatClient) notifyAndGetConnect(id string, dstPort int) *InternalDataPoint {
	ch := make(chan bool)
	c.connNotifyMp[id] = ch
	(*c.conn).Write([]byte(fmt.Sprintf("%s-%d-%d\r\n", id, c.listPort, dstPort)))

	timer := time.NewTimer(2*time.Second)
	select {
		case <- timer.C:
			return nil
		case <- ch:
				for _,p := range c.dps {
					if p.id == id {
						return p
					}
				}
	}

	return nil
}

func (c *NatClient) Stop() {
	defer c.locker.Unlock()
	c.locker.Lock()

	if c.listener != nil {
		defer (*c.listener).Close()
	}

	if c.conn != nil {
		defer (*c.conn).Close()
	}

	for _,p := range c.ports {
		p.Stop()
	}

	for _,d:=range c.dps {
		d.Stop()
	}

	c.ports = c.ports[0:0]
	c.dps = c.dps[0:0]

	common.Logger.Printf("客户端断开:%s", (*c.conn).RemoteAddr())
}