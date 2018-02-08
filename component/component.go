package component

import (
	"sync/atomic"
)

type ComponentID = uint32
type ComponentType uint16

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
