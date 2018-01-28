package network

import (
	"github.com/Cyinx/einx/slog"
	"github.com/Cyinx/protobuf/proto"
	"reflect"
)

type ProtoTypeID = uint16
type Message = proto.Message

var MsgNewMap = make(map[ProtoTypeID]func() interface{})
var ProtoIDMap = make(map[reflect.Type]ProtoTypeID)

func GetMsgID(msg interface{}) (ProtoTypeID, bool) {
	t := reflect.TypeOf(msg)
	id, ok := ProtoIDMap[t]
	return id, ok
}

func RegisterMsgProto(msg_type uint8, msg_id uint8, x Message) uint16 {
	proto_id := uint16(msg_type) | uint16(msg_id)
	t := reflect.TypeOf(x)
	if proto.MessageNewFunc(t) == nil {
		slog.LogInfo("name", "format, ...")
	}
	MsgNewMap[proto_id] = proto.MessageNewFunc(t)
	ProtoIDMap[t] = proto_id
	return proto_id
}

func MsgProtoUnmarshal(type_id ProtoTypeID, data []byte) interface{} {
	msg_new, ok := MsgNewMap[type_id]
	if !ok {
		return nil
	}

	msg := msg_new()
	proto.UnmarshalMerge(data, msg.(Message))
	return msg
}

func MsgProtoMarshal(msg interface{}) ([]byte, error) {
	data, err := proto.Marshal(msg.(Message))
	return data, err
}
