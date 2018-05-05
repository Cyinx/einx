package network

type ProtoTypeID = uint32

type iserializer interface {
	UnmarshalMsg(ProtoTypeID, []byte) interface{}
	MarshalMsg(interface{}) ([]byte, error)
	UnmarshalRpc(ProtoTypeID, []byte) interface{}
	MarshalRpc(interface{}) ([]byte, error)
}

var Serializer iserializer = nil

func SetMsgSerializer(m interface{}) {
	Serializer = m.(iserializer)
}
