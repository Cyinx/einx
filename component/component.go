package component

import (
	"sync/atomic"
)

type ComponentID = uint32
type ComponentType uint16
type EventType = int

type Component interface {
	GetID() ComponentID
	GetType() ComponentType
	Start()
	Close()
}

var component_id ComponentID = 0

func GenComponentID() ComponentID {
	return atomic.AddUint32(&component_id, 1)
}

type ComponentMgr interface {
	OnComponentCreate(ComponentID, Component)
	OnComponentError(Component, error)
}

const (
	COMPONENT_TYPE_BEGIN = iota
	COMPONENT_TYPE_TCP_SERVER
	COMPONENT_TYPE_TCP_CLIENT
	COMPONENT_TYPE_DB_MONGODB
	COMPONENT_TYPE_DB_MYSQL
)
