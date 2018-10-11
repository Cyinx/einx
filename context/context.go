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
	Store(int, interface{})
	Get(int) interface{}
}
