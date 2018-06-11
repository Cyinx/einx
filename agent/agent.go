package agent

import (
	"sync/atomic"
)

type AgentID uint64
type ProtoTypeID = uint32
type EventType = int

type Agent interface {
	GetID() AgentID
	WriteMsg(msg_id ProtoTypeID, msg interface{}) bool
	Close()
	Run()
	GetType() int16
	GetUserType() int16
	SetUserType(int16)
	Destroy()
}

const (
	AgentType_TCP_InComming = iota
	AgentType_TCP_OutGoing
)

type AgentSessionMgr interface {
	OnAgentEnter(AgentID, Agent)
	OnAgentExit(AgentID, Agent)
}

type AgentHandler interface {
	ServeHandler(Agent, ProtoTypeID, []byte)
	ServeRpc(Agent, ProtoTypeID, []byte)
}

var agent_id uint64 = 0

func GenAgentID() AgentID {
	return AgentID(atomic.AddUint64(&agent_id, 1))
}
