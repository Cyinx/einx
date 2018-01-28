package agent

type AgentID uint64
type ProtoTypeID = uint16

type Agent interface {
	GetID() AgentID
	SetID(AgentID)
	WriteMsg(msg interface{}) bool
	Close()
	Run()
	Destroy()
}
