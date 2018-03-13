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

type Server interface {
	GetID() ComponentID
	GetType() ComponentType
	Start(addr string)
	Close()
}

type Linker interface {
	Ping()
	Pong()
}

type Client interface {
	GetID() ComponentID
	GetType() ComponentType
	Start()
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
