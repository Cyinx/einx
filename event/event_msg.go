package event

import (
	"github.com/Cyinx/einx/agent"
)

type Agent = agent.Agent
type EventType int

const (
	EVENT_NONE EventType = iota
	EVENT_TCP_CONNECTED
	EVENT_TCP_ACCEPTED
	EVENT_TCP_READ_MSG
	EVENT_TCP_READ
	EVENT_TCP_WRITE
	EVENT_TCP_CLOSED
	EVENT_MODULE_RPC
)

type EventMsg interface {
	GetType() EventType
	GetSender() interface{}
}

type SessionEventMsg struct {
	MsgType EventType
	Sender  Agent
}

func (this *SessionEventMsg) GetType() EventType {
	return this.MsgType
}

func (this *SessionEventMsg) GetSender() interface{} {
	return this.Sender
}

type DataEventMsg struct {
	MsgType EventType
	Sender  Agent
	TypeID  uint16
	MsgData interface{}
}

func (this *DataEventMsg) GetType() EventType {
	return this.MsgType
}

func (this *DataEventMsg) GetSender() interface{} {
	return this.Sender
}

type RpcEventMsg struct {
	MsgType EventType
	Sender  interface{}
	RpcName string
	Data    []interface{}
}

func (this *RpcEventMsg) GetType() EventType {
	return this.MsgType
}

func (this *RpcEventMsg) GetSender() interface{} {
	return this.Sender
}
