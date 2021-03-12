package znet

import (
	"log"
	"zinx-websocket/ziface"
)

type Handle struct {
	Id uint64
	//Msg     map[uint32]func(response ziface.IResponse)
	Handle map[uint64]ziface.IHandle //存放每个MsgId 所对应的处理方法的map属性
	//Response ziface.IResponse
}

func NewIAgreement(id uint64, handle ziface.IHandle) *Handle {
	a := &Handle{
		Id: id,
		//Response: &Response{},
		//	Msg:     make(map[uint32]func(ziface.IResponse)),
		Handle: make(map[uint64]ziface.IHandle),
		//Response: make(map[uint32]ziface.IResponse),
	}
	a.Handle[id] = handle
	//a.Id[id] = handle
	return a
}

func (a *Handle) PreHandle(req ziface.IRequest) {

	handle, ok := a.Handle[req.GetMessage().GetMsgId()]
	if ok {
		handle.PreHandle(req)
		return
	}

	log.Println("Agreement PreHandle Not Fount")

}

func (a *Handle) PostHandle(req ziface.IRequest) {
	handle, ok := a.Handle[req.GetMessage().GetMsgId()]
	if ok {
		handle.PostHandle(req)
		return
	}

	log.Println("Agreement PostHandle Not Fount")

}
