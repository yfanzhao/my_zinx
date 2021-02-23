/**
  @Time    : 2021/2/20 20:28
  @Author  : zhaoxfan
*/
package ziface

import "net"

//定义链接模块的抽象层
type IConnection interface {
	//启动链接，让当前的链接准备开始工作
	Start()
	//停止链接，结束当前链接的工作
	Stop()
	//获取当前链接的绑定socket conn
	GetTCPConnection() *net.TCPConn
	//获取当前链接模块的链接ID
	GetConnID() uint32
	//获取远程客户端的TCP状态 IP Port
	RemoteAddr() net.Addr
	//发送数据，将数据发送给远程的客户端
	SendMsg(msgId uint32, data []byte) error
	//设置链接属性
	SetProperty(key string, value interface{})
	//获取连接属性
	GetProperty(key string) (interface{}, error)
	//移除连接属性
	RemoveProperty(key string)
}

//定义一个处理链接业务的方法
type HandleFunc func(*net.TCPConn, []byte, int) error
