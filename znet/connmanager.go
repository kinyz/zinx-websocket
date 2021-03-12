package znet

import (
	"errors"
	"log"
	"sync"

	"zinx-websocket/ziface"
)

//ConnManager 连接管理模块
type ConnManager struct {
	connections map[uint32]ziface.IConnection //管理的连接信息
	connLock    sync.RWMutex                  //读写连接的读写锁
}

//NewConnManager 创建一个链接管理
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

//Add 添加链接
func (connMgr *ConnManager) Add(conn ziface.IConnection) {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//将conn连接添加到ConnMananger中
	connMgr.connections[conn.GetConnID()] = conn

	log.Printf("connection add to ConnManager successfully: conn num = ", connMgr.Len())
}

//Remove 删除连接
func (connMgr *ConnManager) Remove(conn ziface.IConnection) {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//删除连接信息
	delete(connMgr.connections, conn.GetConnID())

	log.Printf("connection Remove ConnID=", conn.GetConnID(), " successfully: conn num = ", connMgr.Len())
}

//Get 利用ConnID获取链接
func (connMgr *ConnManager) Get(connID uint32) (ziface.IConnection, error) {
	//保护共享资源Map 加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	}

	return nil, errors.New("connection not found")

}

//Len 获取当前连接
func (connMgr *ConnManager) Len() int {
	return len(connMgr.connections)
}

//ClearConn 清除并停止所有连接
func (connMgr *ConnManager) ClearConn() {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//停止并删除全部的连接信息
	for connID, conn := range connMgr.connections {
		//停止
		conn.Stop()
		//删除
		delete(connMgr.connections, connID)
	}

	log.Printf("Clear All Connections successfully: conn num = ", connMgr.Len())
}

//ClearOneConn  利用ConnID获取一个链接 并且删除
func (connMgr *ConnManager) ClearOneConn(connID uint32) {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	if conn, ok := connMgr.connections[connID]; !ok {
		//停止
		conn.Stop()
		//删除
		delete(connMgr.connections, connID)
		log.Printf("Clear Connections ID:  ", connID, "succeed")
		return
	}

	log.Printf("Clear Connections ID:  ", connID, "err")
	return
}

//获取所有连接
func (connMgr *ConnManager) GetAllConn() map[uint32]ziface.IConnection {
	return connMgr.connections
}
