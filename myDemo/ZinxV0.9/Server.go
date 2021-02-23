/**
  @Time    : 2021/2/20 16:29
  @Author  : zhaoxfan
*/
package main

import (
	"fmt"
	"zinx/ziface"
	"zinx/znet"
)

/*
	基于Zinx框架开发的服务端应用程序
 */

//ping test 自定义路由
type PingRouter struct {
	//先基础BaseRouter
	znet.BaseRouter
}
//Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle...")
	//先读取客户端的数据，再回写ping..ping..ping
	fmt.Println("recv from client:msgID = ", request.GetMsgID(),
		", data = ", string(request.GetData()))
	//回写数据
	err := request.GetConnection().SendMsg(200, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
	}
}

//ping test 自定义路由
type HelloZinxRouter struct {
	//先基础BaseRouter
	znet.BaseRouter
}
//Test Handle
func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call HelloZinxRouter Handle...")
	//先读取客户端的数据，再回写ping..ping..ping
	fmt.Println("recv from client:msgID = ", request.GetMsgID(),
		", data = ", string(request.GetData()))
	//回写数据
	err := request.GetConnection().SendMsg(201, []byte("Hello Welcome To Zinx!"))
	if err != nil {
		fmt.Println(err)
	}
}

//创建链接之后执行的钩子函数
func DoConnectionBegin(conn ziface.IConnection)  {
	fmt.Println("====>DoConnectionBegin is Called...")
	if err := conn.SendMsg(202, []byte("DoConnection BEGIN")); err != nil{
		fmt.Println(err)
	}
}
//链接断开之前的需要执行的函数
func DoConnectionLost(conn ziface.IConnection) {
	fmt.Println("DoConneciotnLost is Called ... ")
	fmt.Println("conn ID = ", conn.GetConnID(), " is Lost...")
}

func main()  {
	//1.创建一个Server句柄，使用Zinx的api
	s := znet.NewServer("[zinx V0.9]")

	//2.注册链接Hook钩子函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	//3. 给当前zinx框架添加自定义的router
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})


	//4.启动Server
	s.Serve()
}
