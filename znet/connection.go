package znet

import (
	"context"
	"errors"

	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"zinx-websocket/pb"

	"sync"

	"github.com/gorilla/websocket"
	"zinx-websocket/utils"
	"zinx-websocket/ziface"
)

//Connection 链接
type Connection struct {
	//当前Conn属于哪个Server
	TCPServer ziface.IServer
	//当前连接的socket TCP套接字
	Conn *websocket.Conn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//消息管理MsgID和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle
	//告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc
	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte

	sync.RWMutex
	//链接属性
	property map[string]interface{}
	////保护当前property的锁
	propertyLock sync.Mutex
	//当前连接的关闭状态
	isClosed bool


	//消息类型 TextMessage 或 BinaryMessage之类
	MessageType int `json:"messageType"`
}

//NewConntion 创建连接的方法
func NewConntion(server ziface.IServer, conn *websocket.Conn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	//初始化Conn属性
	c := &Connection{
		TCPServer:   server,
		Conn:        conn,
		ConnID:      connID,
		isClosed:    false,
		MsgHandler:  msgHandler,
		msgChan:     make(chan []byte),
		msgBuffChan: make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:    nil,
	}

	//将新创建的Conn添加到链接管理中
	c.TCPServer.GetConnMgr().Add(c)
	return c
}

//StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) StartWriter() {
	log.Printf("[Writer Goroutine is running]")
	defer log.Printf(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端

			if err := c.Conn.WriteMessage(c.MessageType, data); err != nil {
				log.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
			//fmt.Printf("Send data succ! data = %+v\n", data)
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if err := c.Conn.WriteMessage(c.MessageType, data); err != nil {
					log.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				log.Println("msgBuffChan is Closed")
				break
			}
		case <-c.ctx.Done():
			return
		}
	}
}

//StartReader 读消息Goroutine，用于从客户端中读取数据
func (c *Connection) StartReader() {
	log.Printf("[Reader Goroutine is running]")
	defer log.Printf(c.RemoteAddr().String(), "[conn Reader exit!]")
	defer c.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// 创建拆包解包的对象
			messageType, data, err := c.Conn.ReadMessage()
			if err != nil {
				log.Println("read data error")
				return
			}

			c.MessageType = messageType //以客户端的类型为准
			msg := &pb.Message{}

			if err := proto.Unmarshal(data, msg);err!=nil{
				log.Print("unpack error:",err)
				return
			}

			req := Request{
				conn:    c,
				message: &Message{Msg: msg},
			}

			if utils.GlobalObject.WorkerPoolSize > 0 {
				//已经启动工作池机制，将消息交给Worker处理
				c.MsgHandler.SendMsgToTaskQueue(&req)
			} else {
				//从绑定好的消息和对应的处理方法中执行对应的Handle方法
				// go c.MsgHandle.DoMsgHandler(req)
				c.MsgHandler.DoMsgHandler(&req)
			}
		}
	}
}

//Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TCPServer.CallOnConnStart(c)
}

//Stop 停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	c.Lock()
	defer c.Unlock()

	c.TCPServer.CallOnConnStop(c)

	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}

	log.Printf("Conn Stop()...ConnID = ", c.ConnID)

	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TCPServer.CallOnConnStop(c)

	// 关闭socket链接
	c.Conn.Close()
	//关闭Writer
	c.cancel()

	//将链接从连接管理器中删除
	c.TCPServer.GetConnMgr().Remove(c)

	//关闭该链接全部管道
	close(c.msgBuffChan)
	//设置标志位
	c.isClosed = true

}

//GetTCPConnection 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *websocket.Conn {
	return c.Conn
}

//GetConnID 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

//RemoteAddr 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//SendMsg 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgID uint64, data []byte) error {
	c.RLock()
	defer c.RUnlock()
	if c.isClosed == true {
		return errors.New("connection closed when send msg")
	}
	message:=&pb.Message{
		MsgId: msgID,
		Data:  data,
	}
	d, err := proto.Marshal(message)
	if err!=nil{
		//log.Print("pack data error:",err)
		return err
	}

	//写回客户端
	c.msgChan <- d

	return nil
}

//SendBuffMsg  发生BuffMsg
func (c *Connection) SendBuffMsg(msgID uint64, data []byte) error {
	c.RLock()
	defer c.RUnlock()
	if c.isClosed == true {
		return errors.New("connection closed when send msg")
	}
	message:=&pb.Message{
		MsgId: msgID,
		Data:  data,
	}
	d, err := proto.Marshal(message)
	if err!=nil{
		//log.Print("pack data error:",err)
		return err
	}

	//写回客户端
	c.msgBuffChan <- d

	return nil
}

//SetProperty 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	if c.property == nil {
		c.property = make(map[string]interface{})
	}

	c.property[key] = value
}

//GetProperty 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	}

	return nil, errors.New("no property found")
}

//RemoveProperty 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

//返回ctx，用于用户自定义的go程获取连接退出状态
func (c *Connection) Context() context.Context {
	return c.ctx
}
