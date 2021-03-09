package znet

import "zinx-websocket/pb"

type Message struct {
	Msg *pb.Message
}

func (m *Message) GetMsgId() uint64 {
	return m.Msg.MsgId
}
func (m *Message) GetData() []byte {
	return m.Msg.Data
}
func (m *Message) GetDataLen() uint32 {
	return uint32(len(m.Msg.Data))
}
func (m *Message) SetData(data []byte) {
	m.Msg.Data = data
}


func (m *Message) SetMsgId(msgId uint64) {
	m.Msg.MsgId = msgId
}
