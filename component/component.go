package component

import (
	"github.com/Cyinx/einx/agent"
)

type ComponentID = agent.AgentID
type ComponentType uint16
type EventType = int

type Component interface {
	GetID() ComponentID
	GetType() ComponentType
	Start()
	Close()
}

func GenComponentID() ComponentID {
	return agent.GenAgentID()
}

const (
	COMPONENT_TYPE_BEGIN = iota
	COMPONENT_TYPE_TCP_SERVER
	COMPONENT_TYPE_TCP_CLIENT
	COMPONENT_TYPE_DB_MONGODB
	COMPONENT_TYPE_DB_MYSQL
)
