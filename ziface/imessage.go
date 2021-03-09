package ziface

/*
	将请求的一个消息封装到message中，定义抽象层接口
*/
type IMessage interface {

	//获取msgid
	GetMsgId() uint64

	//获取data长度
	GetDataLen() uint32
	// 获取data
	GetData() []byte
	//设置data
	SetData(data []byte)
	SetMsgId(msgId uint64)
}