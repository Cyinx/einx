package network

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/component"
	"github.com/Cyinx/einx/module"
)

type Agent = agent.Agent
type AgentID = agent.AgentID
type ServerType uint32

type ModuleEventer = module.ModuleEventer

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

type Connection interface {
	RemoteAddr() string
}

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

const (
	COMPONENT_TYPE_TCP_SERVER = component.COMPONENT_TYPE_TCP_SERVER
	COMPONENT_TYPE_TCP_CLIENT = component.COMPONENT_TYPE_TCP_CLIENT
)
