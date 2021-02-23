/**
  @Time    : 2021/2/20 20:28
  @Author  : zhaoxfan
*/
package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/utils"
	"zinx/ziface"
)

//链接模块
type Connection struct {
	//当前Conn率属于的Server  在conn初始化的时候添加即可
	TcpServer ziface.IServer
	//当前链接的socket TCP套接字
	Conn *net.TCPConn
	//链接的ID
	ConnID uint32
	//当前的链接状态
	isClosed bool
	//告知当前链接已经退出的/停止 channel
	ExitChan chan bool
	//无缓冲的管理，用于读写Goroutine之间的消息通信
	msgChan chan []byte
	//消息的管理MsgID和对应的处理业务API关系
	MsgHandle ziface.IMsgHandle
	//====================================
	//链接属性集合
	property map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
	//====================================
}

//初始化链接模块的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandle ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer: server,
		Conn: conn,
		ConnID: connID,
		MsgHandle: msgHandle,
		isClosed: false,
		msgChan: make(chan []byte),
		ExitChan: make(chan bool, 1),
		property: make(map[string]interface{}), //对链接属性map初始化
	}
	//将conn加入到ConnManager中
	c.TcpServer.GetConnMgr().Add(c)

	return c
}

//链接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("[Reader Goroutine is running]")
	defer fmt.Println("[Reader is exit] connID=", c.ConnID, " , remote addr is ", c.RemoteAddr().String())
	defer c.Stop()
	for{
		//读取客户端的数据到buf中，最大512字节
		//buf := make([]byte, utils.GlobalObject.MaxPacketSize)
		//_, err := c.Conn.Read(buf)
		//if err != nil {
		//	fmt.Println("recv conn err", err)
		//	continue
		//}
		//创建一个拆包解包对象
		dp := NewDataPack()
		//读取客户端的Msg Head 二进制流 8个字节，
		headData := make([]byte, dp.GetHeadLen())
		//拆包 得到msgID和msgDatalen放在msg消息中
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			break
		}
		//拆包，得到msgID和msgDatalen放在msg消息中
		msg, err := dp.Unpack(headData)
		if err != nil{
			fmt.Println("unpack error", err)
			break
		}
		//根据dataLen 再次读取Data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				break
			}
		}
		msg.SetData(data)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg: msg,
		}

		if utils.GlobalObject.WorkerPoolSize > 0{
			//已经开启了工作池机制  将消息发送给Worker工作池处理即可
			c.MsgHandle.SendMsgToTaskQueue(&req)
		}else {
			//从路由中，找到注册绑定的Conn对应的router调用
			//根据绑定好的MsgID找到对应处理API业务执行
			go c.MsgHandle.DoMsgHandler(&req)
		}
	}
}
//写消息Goroutine,专门发送给客户端消息的模块
func (c *Connection) StartWriter() {
	fmt.Println("[Write Goroutine is running]")
	defer fmt.Println(c.RemoteAddr(), " [conn Write exit!]")
	//不断的阻塞的等待channel的消息，进行写给客户端
	for{
		select {
		case data := <- c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil{
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		case <- c.ExitChan:
			//代表Reader已经退出 此时Write也要退出  conn已经关闭
			return
		}
	}
}

//启动链接，让当前的链接准备开始工作
func (c *Connection) Start(){
	fmt.Println("Conn Start()...ConnID=", c.ConnID)
	// 启动从当前链接读数据的业务
	go c.StartReader()
	// 启动从当前链接写数据的业务
	go c.StartWriter()
	//按照开发者传递进来的 创建连接之后需要调用的处理业务  执行对应Hook函数
	c.TcpServer.CallOnConnStart(c)
}
//停止链接，结束当前链接的工作
func (c *Connection) Stop(){
	fmt.Println("Conn stop()... ConnID=", c.ConnID)
	//如果当前链接已经关闭
	if c.isClosed == true{
		return
	}
	c.isClosed = true
	//调用开发者注册的 在销毁链接之前需要执行的业务Hook函数
	c.TcpServer.CallOnConnStop(c)
	//关闭socket链接
	c.Conn.Close()
	//告知Writer关闭
	c.ExitChan <- true
	//将当前链接从ConnMgr中摘除掉
	c.TcpServer.GetConnMgr().Remove(c)
	//回收资源
	close(c.ExitChan)
	close(c.msgChan)
}
//获取当前链接的绑定socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn{
	return c.Conn
}
//获取当前链接模块的链接ID
func (c *Connection) GetConnID() uint32{
	return c.ConnID
}
//获取远程客户端的TCP状态 IP Port
func (c *Connection) RemoteAddr() net.Addr{
	return c.Conn.RemoteAddr()
}
//提供一个SendMsg方法 将我们要发送给客户端的数据，先进行封包，再发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return  errors.New("Pack error msg ")
	}
	//写回客户端
	c.msgChan <- msg
	return nil
}

//设置链接属性
func (c *Connection) SetProperty(key string, value interface{}){
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	//添加一个链接属性
	c.property[key] = value
}
//获取连接属性
func (c *Connection) GetProperty(key string) (interface{}, error){
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	//读取属性
	if value, ok := c.property[key]; ok{
		return value, nil
	}else {
		return nil, errors.New("no property found")
	}
}
//移除连接属性
func (c *Connection) RemoveProperty(key string){
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	//删除属性
	delete(c.property, key)
}