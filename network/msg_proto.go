package network

import (
	"sync"
)

type ProtoTypeID = uint32

type iserializer interface {
	UnmarshalMsg(ProtoTypeID, []byte) interface{}
	MarshalMsg(*sync.Pool, interface{}) ([]byte, error, bool)
	UnmarshalRpc(ProtoTypeID, []byte) interface{}
	MarshalRpc(*sync.Pool, interface{}) ([]byte, error, bool)
}

var Serializer iserializer = nil

func SetMsgSerializer(m interface{}) {
	Serializer = m.(iserializer)
}
