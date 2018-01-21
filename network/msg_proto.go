package network

import (
	"github.com/golang/protobuf/proto"
	"reflect"
	//"sync"
)

type ProtoTypeID = uint16
type MsgProto struct {
	msg_id   ProtoTypeID
	msg_type reflect.Type
}

var MsgProtoMap = make(map[ProtoTypeID]*MsgProto)

func (this *MsgProto) GetMsgID() ProtoTypeID {
	return this.msg_id
}

func MakeMsgProto(msg_type uint8, msg_id uint8, body interface{}) *MsgProto {
	var msg_proto_id uint16 = uint16(msg_type)
	var msg_body_type reflect.Type = reflect.TypeOf(body)

	msg_proto_id = msg_proto_id << 8
	msg_proto_id = msg_proto_id | uint16(msg_id)

	msg_proto := &MsgProto{
		msg_id:   msg_proto_id,
		msg_type: msg_body_type,
	}

	MsgProtoMap[msg_proto.msg_id] = msg_proto
	return msg_proto
}

func MsgProtoUnmarshal(type_id ProtoTypeID, data []byte) interface{} {
	msg_proto, ok := MsgProtoMap[type_id]
	if !ok {
		return nil
	}

	msg := reflect.New(msg_proto.msg_type).Interface()
	proto.UnmarshalMerge(data, msg.(proto.Message))

	return msg
}

func MsgProtoMarshal(msg interface{}) ([]byte, error) {
	data, err := proto.Marshal(msg.(proto.Message))
	return data, err
}
