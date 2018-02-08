package network

import (
	"github.com/Cyinx/einx/slog"
	"github.com/Cyinx/protobuf/proto"
	"reflect"
)

type ProtoTypeID = uint32
type Message = proto.Message

var MsgNewMap = make(map[ProtoTypeID]func() interface{})

func RegisterMsgProto(msg_type uint16, msg_id uint16, x Message) ProtoTypeID {
	proto_id := uint32(msg_type<<16) | uint32(msg_id)
	t := reflect.TypeOf(x)
	f := proto.MessageNewFunc(t)
	if f == nil {
		slog.LogInfo("msg_proto", "unregister message new func [%s]", t.Name())
	} else {
		MsgNewMap[proto_id] = f
	}
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
