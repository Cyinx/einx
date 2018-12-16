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

type TransportOption struct {
	msg_max_length uint16
	msg_max_count  int32 //max msg count per seconds
}

type Option = func(...interface{})

func Name(name string) Option {
	return func(args ...interface{}) {
		t := args[0]
		switch v := t.(type) {
		case *TcpServerMgr:
			v.name = name
		case *TcpClientMgr:
			v.name = name
		default:
			panic("option network name unknown type")
		}
	}
}

func Module(m EventReceiver) Option {
	return func(args ...interface{}) {
		t := args[0]
		switch v := t.(type) {
		case *TcpServerMgr:
			v.module = m
		case *TcpClientMgr:
			v.module = m
		default:
			panic("option network module unknown type")
		}
	}
}

func ListenAddr(addr string) Option {
	return func(args ...interface{}) {
		t := args[0]
		switch v := t.(type) {
		case *TcpServerMgr:
			v.addr = addr
		default:
			panic("option network listen addr unknown type")
		}
	}
}

func ServeHandler(serve_handler SessionHandler) Option {
	return func(args ...interface{}) {
		t := args[0]
		switch v := t.(type) {
		case *TcpServerMgr:
			v.agent_handler = serve_handler
		case *TcpClientMgr:
			v.agent_handler = serve_handler
		default:
			panic("option network serve handler unknown type")
		}
	}
}
