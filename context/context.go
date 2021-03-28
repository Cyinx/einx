package context

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/component"
)

type Agent = agent.Agent
type AgentID = agent.AgentID
type Component = component.Component

type Context interface {
	GetModule() Module
	GetSender() Agent
	GetComponent() Component
	GetAttach() interface{}
	Store(int, interface{})
	Get(int) interface{}
	Done(args ...interface{})
}
