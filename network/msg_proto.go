package network

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/slog"
	"github.com/golang/protobuf/proto"
	"reflect"
)

type ProtoTypeID = uint16

type newMsgFunc func() interface{}

var MsgNewMap = make(map[ProtoTypeID]newMsgFunc)
var ProtoIDMap = make(map[reflect.Type]ProtoTypeID)

func RegisterMsgProto(msg_type uint8, msg_id uint8, new_func newMsgFunc) uint16 {
	proto_id := uint16(msg_type) | uint16(msg_id)
	MsgProtoMap[proto_id] = new_wrapper
	msg_type := reflect.TypeOf(new_wrapper())
	ProtoIDMap[msg_type] = proto_id
	return proto_id
}

func MsgProtoUnmarshal(type_id ProtoTypeID, data []byte) interface{} {
	msg_new, ok := MsgNewMap[type_id]
	if !ok {
		return nil
	}

	msg := msg_new()
	proto.UnmarshalMerge(data, msg.GetBody().(proto.Message))
	return msg
}

func MsgProtoMarshal(msg interface{}) ([]byte, error) {
	data, err := proto.Marshal(msg.(proto.Message))
	return data, err
}
