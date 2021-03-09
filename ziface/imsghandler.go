package ziface

///*
//	消息管理抽象层
//*/
//type IMsgHandle interface {
//	DoMsgHandler(request IRequest)          //马上以非阻塞方式处理消息
//	AddRouter(msgID uint32, router IRouter) //为消息添加具体的处理逻辑
//	StartWorkerPool()                       //启动worker工作池
//	SendMsgToTaskQueue(request IRequest)    //将消息交给TaskQueue,由worker进行处理
//}

/*
	消息管理抽象层
*/
type IMsgHandle interface {
	DoMsgHandler(request IRequest) //马上以非阻塞方式处理消息
	AddCustomHandle(handle IHandle)
	AddRouter(msgID uint64, handle IRouter) //为消息添加具体的处理逻辑
	StartWorkerPool()                    //启动worker工作池
	SendMsgToTaskQueue(request IRequest) //将消息交给TaskQueue,由worker进行处理

}

/*
	路由接口， 这里面路由是 使用框架者给该链接自定的 处理业务方法
	路由里的IRequest 则包含用该链接的链接信息和该链接的请求数据信息
*/
type IHandle interface {
	PreHandle(request IRequest)  //在处理conn业务之前的钩子方法
	PostHandle(request IRequest) //处理conn业务之后的钩子方法
}