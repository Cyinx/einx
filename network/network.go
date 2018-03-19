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

const (
	ServerType_TCP = iota
	ServerType_UDP
	ClientType_TCP
	ClientType_UDP
)

type ITcpServerCom interface {
	GetID() ComponentID
	GetType() ComponentType
}

const (
	AgentType_TCP_InComming = iota
	AgentType_TCP_OutGoing
)

type Connection interface {
	RemoteAddr() string
}

type Linker interface {
	Ping()
	Pong()
}

type ITcpClientCom interface {
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
