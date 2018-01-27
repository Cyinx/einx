package agent

type AgentID uint64
type ProtoTypeID = uint16

type MsgWrapper interface {
	GetMsgID() ProtoTypeID
	GetBody() interface{}
}

type Agent interface {
	GetID() AgentID
	SetID(AgentID)
	WriteMsg(msg MsgWrapper) bool
	Close()
	Run()
	Destroy()
}
