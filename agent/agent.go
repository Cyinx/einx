package agent

import (
	"sync/atomic"
)

type AgentID = uint64
type ProtoTypeID = uint32
type EventType = int

type Agent interface {
	GetID() AgentID
	Close()
}

const (
	AgentType_TCP_InComming = iota
	AgentType_TCP_OutGoing
)

var agent_id uint64 = 0

func GenAgentID() AgentID {
	return AgentID(atomic.AddUint64(&agent_id, 1))
}
