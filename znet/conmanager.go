/**
  @Time    : 2021/2/22 11:25
  @Author  : zhaoxfan
*/
package znet

import (
	"errors"
	"fmt"
	"sync"
	"zinx/ziface"
)

//连接管理模块
type ConnManager struct {
	//管理的链接信息
	connections map[uint32]ziface.IConnection
	//读写链接集合的读写锁
	connLock sync.RWMutex
}

//创建一个连接管理
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

//添加连接
func (connMgr *ConnManager) Add(conn ziface.IConnection){
	//保护共享资源Map  加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//将conn连接添加到connManager中
	connMgr.connections[conn.GetConnID()] = conn
	fmt.Println("connection add to ConnManager successfully: conn num = ", connMgr.Len())
}
//删除连接
func (connMgr *ConnManager) Remove(conn ziface.IConnection){
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//删除连接信息
	delete(connMgr.connections, conn.GetConnID())
	fmt.Println("connection Remove ConnID=",conn.GetConnID(), " successfully: conn num = ", connMgr.Len())
}
//利用ConnID获取链接
func (connMgr *ConnManager) Get(connID uint32) (ziface.IConnection, error){
	//保护共享资源Map 加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok{
		//找到了
		return conn, nil
	}else {
		return nil, errors.New("connection not FOUND!")
	}
}
//获取当前连接总数
func (connMgr *ConnManager) Len() int{
	return len(connMgr.connections)
}
//清除并停止所有连接
func (connMgr *ConnManager) ClearConn() {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//停止并删除全部的连接信息
	for connID, conn := range connMgr.connections{
		//停止
		conn.Stop()
		//删除
		delete(connMgr.connections, connID)
	}
	fmt.Println("Clear All Connections successfully: conn num = ", connMgr.Len())
}