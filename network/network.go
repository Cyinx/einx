package network

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/component"
	"github.com/Cyinx/einx/event"
	"net"
)

type Agent = agent.Agent
type AgentID = agent.AgentID
type ProtoTypeID = uint32
type ServerType uint32
type EventReceiver = event.EventReceiver
type ComponentID = component.ComponentID
type ComponentType = component.ComponentType
type Component = component.Component

func GenComponentID() ComponentID {
	return component.GenComponentID()
}

type ITcpServerMgr interface {
	GetID() ComponentID
	GetType() ComponentType
}

const (
	AgentType_TCP_InComming = agent.AgentType_TCP_InComming
	AgentType_TCP_OutGoing  = agent.AgentType_TCP_OutGoing
)

type Linker interface {
	Ping()
	Pong()
}

type ITcpClientMgr interface {
	GetID() ComponentID
	GetType() ComponentType
	Connect(addr string, user_type int16)
}

type ConnType uint16

const (
	ConnType_TCP = iota
	ConnType_UDP
)

type WriteWrapper struct {
	msg_type byte
	msg_id   ProtoTypeID
	buffer   []byte
}

func (w *WriteWrapper) reset() {
	w.msg_type = 0
	w.msg_id = 0
	w.buffer = nil
}

const (
	COMPONENT_TYPE_TCP_SERVER = component.COMPONENT_TYPE_TCP_SERVER
	COMPONENT_TYPE_TCP_CLIENT = component.COMPONENT_TYPE_TCP_CLIENT
)

type NetLinker interface {
	GetID() AgentID
	Close()
	RemoteAddr() net.Addr
	WriteMsg(msg_id ProtoTypeID, b []byte) bool
	GetUserType() int16
	SetUserType(int16)
	Run()
}

type SessionMgr interface {
	OnLinkerConneted(AgentID, Agent)
	OnLinkerClosed(AgentID, Agent)
}

type SessionHandler interface {
	ServeHandler(Agent, ProtoTypeID, []byte)
	ServeRpc(Agent, ProtoTypeID, []byte)
}

func Run() {
	go OnKeepAliveUpdate()
}
