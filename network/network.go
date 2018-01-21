package network

import (
	"github.com/Cyinx/einx/module"
)

type ServerType uint16

const (
	ServerType_TCP = iota
	ServerType_UDP
)

type Server interface {
	GetType() ServerType
	Start(addr string)
	Close()
}

type ClientType uint16

const (
	ClientType_TCP = iota
	ClientType_UDP
)

type Client interface {
	GetType() ClientType
	Start()
}

type ConnType uint16

const (
	ConnType_TCP = iota
	ConnType_UDP
)

type WriteWrapper struct {
	msg_id ProtoTypeID
	buffer []byte
}

var _event_module = module.GetModule(module.MAIN_MODULE).(module.ModuleEventer)
