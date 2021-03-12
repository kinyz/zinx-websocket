package znet

import (
	"CardExpert/library/zinx-websocket/utils"
	"CardExpert/library/zinx-websocket/ziface"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
)

//Server 接口实现，定义一个Server服务类
type Server struct {
	//服务器的名称
	Name string
	//服务器协议 ws,wss
	Scheme string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgID和对应的处理方法
	msgHandler ziface.IMsgHandle
	//当前Server的链接管理器
	ConnMgr ziface.IConnManager
	//该Server的连接创建时Hook函数
	OnConnStart func(conn ziface.IConnection)
	//该Server的连接断开时的Hook函数
	OnConnStop func(conn ziface.IConnection)
	//协议
	Path string
}

//NewServer 创建一个服务器句柄
func NewServer() ziface.IServer {
	printLogo()

	s := &Server{
		Name:       utils.GlobalObject.Name,
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TCPPort,
		msgHandler: NewMsgHandle(),
		Scheme:     utils.GlobalObject.Scheme,
		Path:       utils.GlobalObject.Path, // 比如 /echo
		ConnMgr:    NewConnManager(),
	}
	return s
}

//连接信息
var upgrader = &websocket.Upgrader{
	ReadBufferSize:  int(utils.GlobalObject.MaxPacketSize), //读取最大值
	WriteBufferSize: int(utils.GlobalObject.MaxPacketSize), //写最大值
	//解决跨域问题
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//============== 实现 ziface.IServer 里的全部接口方法 ========

//websocket回调
func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {

	//log.Println("query",r.Header.Values("Sec-WebSocket-Protocol"))
	protocol := r.Header.Get("Sec-WebSocket-Protocol")
	// 新增自定义协议头
	upgrader.Subprotocols = r.Header.Values("Sec-WebSocket-Protocol")
	//log.Println(upgrader.Subprotocols)
	hh := http.Header{}
	hh.Set("Sec-WebSocket-Protocol", protocol)
	//hh["Sec-WebSocket-Protocol"] = r.Header.Values("Sec-WebSocket-Protocol")
	conn, err := upgrader.Upgrade(w, r, hh)
	if err != nil {
		log.Println("server wsHandler upgrade err:", err)
		return
	}
	// defer log.Println("server wsHandler client is closed")
	// defer conn.Close()

	// 判断是否超出个数
	if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
		//todo 给用户发一个关闭连接消息
		log.Println("server wsHandler too many connection")

		conn.Close()
		return
	}

	log.Println("server wsHandler a new client coming ip:", conn.RemoteAddr())
	//处理新连接业务方法
	cid++

	dealConn := NewConntion(s, conn, cid, protocol, s.msgHandler)
	go dealConn.Start()
}

//全局conectionid 后续使用uuid生成
var cid uint32

//Start 开启网络服务
func (s *Server) Start() {
	fmt.Printf("[START] Server name: %s,listenner at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)

	//开启一个go去做服务端Linster业务
	go func() {
		//0 启动worker工作池机制
		s.msgHandler.StartWorkerPool()

		//已经监听成功
		fmt.Println("start ", s.Name, " succ, now listenning...")

		//TODO server.go 应该有一个自动生成ID的方法
		//	var cid uint32
		cid = 0
		http.HandleFunc("/"+s.Path, s.wsHandler)
		err := http.ListenAndServe(s.IP+":"+strconv.Itoa(int(s.Port)), nil)
		if err != nil {
			log.Println("server start listen error:", err)
		}
	}()
}

//Stop 停止服务
func (s *Server) Stop() {
	log.Printf("[STOP] Zinx server , name ", s.Name)

	//将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.ConnMgr.ClearConn()
}

//Serve 运行服务
func (s *Server) Serve() {
	s.Start()

	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则主Go退出， listenner的go将会退出
	select {}
}

//路由功能：给当前服务注册一个自定义头 找不到id时调用，供客户端链接处理使用
func (s *Server) AddCustomHandle(handle ziface.IHandle) {
	s.msgHandler.AddCustomHandle(handle)
}

//AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) AddRouter(msgID uint64, router ziface.IRouter) {
	s.msgHandler.AddRouter(msgID, router)
}

//GetConnMgr 得到链接管理
func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

//SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

//SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

//CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		log.Printf("---> CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

//CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		log.Printf("---> CallOnConnStop....")
		s.OnConnStop(conn)
	}
}

func printLogo() {

	log.Printf("[Zinx-websocket] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)
}

func init() {
}
