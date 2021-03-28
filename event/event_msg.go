package event

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/component"
)

type Agent = agent.Agent
type ProtoTypeID = agent.ProtoTypeID
type Component = component.Component
type ComponentID = component.ComponentID
type EventType = int

const (
	EVENT_NONE EventType = iota
	EVENT_TCP_CONNECTED
	EVENT_TCP_CONNECT_FAILED
	EVENT_TCP_ACCEPTED
	EVENT_TCP_READ_MSG
	EVENT_TCP_READ
	EVENT_TCP_WRITE
	EVENT_TCP_CLOSED
	EVENT_MODULE_RPC
	EVENT_MODULE_AWAITRPC
	EVENT_COMPONENT_CREATE
	EVENT_COMPONENT_ERROR
	EVENT_COMPONENT_CUSTOM
)

type EventMsg interface {
	GetType() EventType
	Reset()
}

type ComponentEventMsg struct {
	MsgType EventType
	Sender  Component
	Attach  interface{}
	Err     error
}

func (m *ComponentEventMsg) GetType() EventType {
	return m.MsgType
}

func (m *ComponentEventMsg) GetSender() interface{} {
	return m.Sender
}

func (m *ComponentEventMsg) Reset() {
	m.MsgType = 0
	m.Sender = nil
	m.Attach = nil
	m.Err = nil
}

type SessionEventMsg struct {
	MsgType EventType
	Sender  Agent
	Cid     ComponentID
	Args    []interface{}
}

func (m *SessionEventMsg) GetType() EventType {
	return m.MsgType
}

func (m *SessionEventMsg) GetSender() Agent {
	return m.Sender
}

func (m *SessionEventMsg) Reset() {
	m.MsgType = 0
	m.Sender = nil
	m.Args = nil
	m.Cid = 0
}

type DataEventMsg struct {
	MsgType EventType
	Sender  Agent
	TypeID  ProtoTypeID
	MsgData interface{}
}

func (m *DataEventMsg) GetType() EventType {
	return m.MsgType
}

func (m *DataEventMsg) GetSender() Agent {
	return m.Sender
}

func (m *DataEventMsg) Reset() {
	m.MsgType = 0
	m.Sender = nil
	m.TypeID = 0
	m.MsgData = nil
}

type RpcEventMsg struct {
	MsgType EventType
	Sender  Agent
	RpcName string
	Data    []interface{}
}

func (m *RpcEventMsg) GetType() EventType {
	return m.MsgType
}

func (m *RpcEventMsg) GetSender() Agent {
	return m.Sender
}

func (m *RpcEventMsg) Reset() {
	m.MsgType = 0
	m.Sender = nil
	m.RpcName = ""
	m.Data = nil
}

type CustomActionEventMsg interface {
	GetType() EventType
	GetSender() Agent
	GetAction() func(CustomActionEventMsg)
}

type EventReceiver interface {
	PostEvent(EventType, Agent, ComponentID, ...interface{})
	PostData(EventType, ProtoTypeID, Agent, interface{})
	PushEventMsg(ev EventMsg)
}

type AwaitRpcEventMsg struct {
	MsgType EventType
	Sender  Agent
	RpcName string
	Data    []interface{}
	Await   chan []interface{}
}

func (m *AwaitRpcEventMsg) GetType() EventType {
	return m.MsgType
}

func (m *AwaitRpcEventMsg) GetSender() Agent {
	return m.Sender
}

func (m *AwaitRpcEventMsg) Reset() {
	m.MsgType = 0
	m.Sender = nil
	m.RpcName = ""
	m.Data = nil
	m.Await = nil
}