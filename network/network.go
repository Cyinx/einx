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
	Address() net.Addr
}

const (
	Linker_TCP_InComming = iota
	Linker_TCP_OutGoing
)

type Linker interface {
	GetID() AgentID
	Ping() bool
	DoPong(int64)
	GetOption() *TransportOption
}

type ITcpClientMgr interface {
	GetID() ComponentID
	GetType() ComponentType
	Connect(addr string, user_type interface{})
}

type ConnType uint16

const (
	ConnType_TCP = iota
	ConnType_UDP
)

type ITransportMsg interface {
	GetType() byte
	reset()
}

type TransportMsgPack struct {
	msgType byte
	msgID   ProtoTypeID
	Buf     []byte
}

func (m *TransportMsgPack) GetType() byte {
	return m.msgType
}

func (w *TransportMsgPack) reset() {
	w.msgType = 0
	w.msgID = 0
	w.Buf = nil

	writePool.Put(w)
}

type ITransporter interface {
	IsClosed() bool
	doPushWrite(ITransportMsg) bool
}

type ITranMsgMultiple interface {
	WriteMsg(ProtoTypeID, []byte) bool
	RpcCall(ProtoTypeID, []byte) bool
	Done() bool
}

type TransportMultiple struct {
	trans    ITransporter
	count    int
	msgArray []*TransportMsgPack
}

func (m *TransportMultiple) GetType() byte {
	return 'B'
}

func (m *TransportMultiple) WriteMsg(msgID ProtoTypeID, b []byte) bool {
	if m.trans.IsClosed() == true {
		return false
	}

	w := writePool.Get().(*TransportMsgPack)
	w.msgType = 'P'
	w.msgID = msgID
	w.Buf = b
	m.count += len(b)
	m.msgArray = append(m.msgArray, w)
	return true
}

func (m *TransportMultiple) RpcCall(msgiD ProtoTypeID, b []byte) bool {
	if m.trans.IsClosed() == true {
		return false
	}

	w := writePool.Get().(*TransportMsgPack)
	w.msgType = 'R'
	w.msgID = msgiD
	w.Buf = b
	m.count += len(b)
	m.msgArray = append(m.msgArray, w)
	return true
}

func (m *TransportMultiple) Done() bool {
	if m.trans.IsClosed() == true {
		return false
	}
	m.trans.doPushWrite(m)
	return true
}

func (b *TransportMultiple) reset() {
	b.trans = nil
	msgArray := b.msgArray
	b.msgArray = nil
	for _, v := range msgArray {
		writePool.Put(v)
	}
}

const (
	COMPONENT_TYPE_TCP_SERVER = component.COMPONENT_TYPE_TCP_SERVER
	COMPONENT_TYPE_TCP_CLIENT = component.COMPONENT_TYPE_TCP_CLIENT
)

type NetLinker interface {
	GetID() AgentID
	Close()
	RemoteAddr() net.Addr
	WriteMsg(ProtoTypeID, []byte) bool
	RpcCall(ProtoTypeID, []byte) bool
	MultipleMsg() ITranMsgMultiple
	GetUserType() interface{}
	SetUserType(interface{})
	Run() error
}

type SessionMgr interface {
	OnLinkerConnected(AgentID, Agent)
	OnLinkerClosed(AgentID, Agent, error)
}

type SessionHandler interface {
	ServeHandler(Agent, ProtoTypeID, []byte)
	ServeRpc(Agent, ProtoTypeID, []byte)
}

func Run() {
	go pingMgr.Run()
}
