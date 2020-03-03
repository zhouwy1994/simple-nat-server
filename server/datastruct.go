package server

import "net"

type DataNode struct {
	Key string
	Conn *net.Conn
}

type NodeListener struct {
	Key string // 节点唯一标识
	ExternalAddr string // 外网监听地址
	NodeAddr string // 节点地址(客户端映射出的一个端口)
	Listener *net.Listener // 外网监听器句柄
	ConnList []*DataNode // 连接列表
}

type ClientNode struct {
	Key string // 客户端唯一标识
	Conn *net.Conn // 客户端连接句柄
	NodeList [] *NodeListener // 节点列表
}

type RegisterListener struct {
	ListenAddr string // 监听地址
	Listener *net.Listener // 监听器句柄
	ClientList []*ClientNode // 客户端列表
}

type DataListener struct {
	ListenAddr string // 监听地址
	Listener *net.Listener // 监听器句柄
	ConnList []*DataNode // 连接列表
}

type Server struct {
	registerListener *RegisterListener
	dataListener *DataListener
}
