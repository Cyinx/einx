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
	EVENT_COMPONENT_CREATE
	EVENT_COMPONENT_ERROR
)

type EventMsg interface {
	GetType() EventType
	Reset()
}

type ComponentEventMsg struct {
	MsgType EventType
	Sender  Component
	Attach  interface{}
}

func (this *ComponentEventMsg) GetType() EventType {
	return this.MsgType
}

func (this *ComponentEventMsg) GetSender() interface{} {
	return this.Sender
}

func (this *ComponentEventMsg) Reset() {
	this.MsgType = 0
	this.Sender = nil
	this.Attach = nil
}

type SessionEventMsg struct {
	MsgType EventType
	Sender  Agent
	Cid     ComponentID
}

func (this *SessionEventMsg) GetType() EventType {
	return this.MsgType
}

func (this *SessionEventMsg) GetSender() Agent {
	return this.Sender
}

func (this *SessionEventMsg) Reset() {
	this.MsgType = 0
	this.Sender = nil
	this.Cid = 0
}

type DataEventMsg struct {
	MsgType EventType
	Sender  Agent
	TypeID  ProtoTypeID
	MsgData interface{}
}

func (this *DataEventMsg) GetType() EventType {
	return this.MsgType
}

func (this *DataEventMsg) GetSender() Agent {
	return this.Sender
}

func (this *DataEventMsg) Reset() {
	this.MsgType = 0
	this.Sender = nil
	this.TypeID = 0
	this.MsgData = nil
}

type RpcEventMsg struct {
	MsgType EventType
	Sender  Agent
	RpcName string
	Data    []interface{}
}

func (this *RpcEventMsg) GetType() EventType {
	return this.MsgType
}

func (this *RpcEventMsg) GetSender() Agent {
	return this.Sender
}

func (this *RpcEventMsg) Reset() {
	this.MsgType = 0
	this.Sender = nil
	this.RpcName = ""
	this.Data = nil
}

type EventReceiver interface {
	PostEvent(EventType, Agent, ComponentID)
	PostData(EventType, ProtoTypeID, Agent, interface{})
	PushEventMsg(ev EventMsg)
}
