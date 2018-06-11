package network

import (
	//"github.com/Cyinx/einx/slog"
	"encoding/binary"
	"errors"
	"io"
)

type IReader interface {
	Read(b []byte) (n int, err error)
}

const (
	MSG_KEY_LENGTH    = 32
	MSG_HEADER_LENGTH = 4
	MSG_ID_LENGTH     = 4
)

// --------------------------------------------------------------------------------------------------------
// |                                header              | body                          |
// | type byte | body_length uint16 | packet_flag uint8 | msg_id uint32| msg_data []byte|
// --------------------------------------------------------------------------------------------------------
var msg_header_length int = MSG_HEADER_LENGTH

type PacketHeader struct {
	MsgType    byte
	BodyLength uint16
	PacketFlag uint8
}

func MsgHeaderLength() int {
	return msg_header_length
}

func ReadBinary(r io.Reader, data interface{}) error {
	return binary.Read(r, binary.LittleEndian, data)
}

func UnmarshalMsgBinary(packet *PacketHeader, b []byte) (ProtoTypeID, interface{}, error) {
	var msg_id ProtoTypeID = 0
	if len(b) < MSG_ID_LENGTH {
		return 0, nil, errors.New("msg packet length error")
	}
	msg_id = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24 //小端
	msg_body := b[MSG_ID_LENGTH:]
	var msg interface{}
	switch packet.MsgType {
	case 'P':
		msg = Serializer.UnmarshalMsg(msg_id, msg_body)
		break
	case 'R':
		msg = Serializer.UnmarshalRpc(msg_id, msg_body)
		break
	default:
		break
	}
	return msg_id, msg, nil
}

func ReadMsgPacket(r io.Reader, msg_packet *PacketHeader, header_buffer []byte, b *[]byte) (ProtoTypeID, []byte, error) {
	if _, err := io.ReadFull(r, header_buffer); err != nil {
		return 0, nil, err
	}

	msg_packet.MsgType = header_buffer[0]
	msg_packet.BodyLength = uint16(header_buffer[1]) | (uint16(header_buffer[2]) << 8) //小端
	msg_packet.PacketFlag = header_buffer[3]

	if msg_packet.MsgType == 'T' {
		return 0, nil, nil
	}

	if cap(*b) < int(msg_packet.BodyLength) {
		*b = make([]byte, msg_packet.BodyLength)
	} else {
		*b = (*b)[0:msg_packet.BodyLength]
	}

	if _, err := io.ReadFull(r, *b); err != nil {
		return 0, nil, err
	}

	body := *b
	if len(body) < MSG_ID_LENGTH {
		return 0, nil, errors.New("msg packet length error")
	}

	var msg_id ProtoTypeID = 0
	msg_id = uint32(body[0]) | uint32(body[1])<<8 | uint32(body[2])<<16 | uint32(body[3])<<24 //小端
	msg_body := body[MSG_ID_LENGTH:]
	return msg_id, msg_body, nil
}

func MarshalMsgBinary(msg_id ProtoTypeID, msg_buffer []byte, b *[]byte) bool {
	var msg_body_length int = len(msg_buffer) + MSG_ID_LENGTH
	var msg_length int = msg_body_length + MSG_HEADER_LENGTH

	if cap(*b) < msg_length {
		*b = make([]byte, msg_length)
	} else {
		*b = (*b)[:msg_length]
	}

	buffer := *b
	//packet header
	buffer[0] = 'P'
	buffer[1] = byte(msg_body_length & 0xFF)
	buffer[2] = byte(msg_body_length >> 8)
	buffer[3] = 0

	//msg wrapper
	buffer[4] = byte(msg_id)
	buffer[5] = byte(msg_id >> 8)
	buffer[6] = byte(msg_id >> 16)
	buffer[7] = byte(msg_id >> 24)

	copy(buffer[MSG_HEADER_LENGTH+MSG_ID_LENGTH:], msg_buffer)

	return true
}

func MarshalKeepAliveMsgBinary(b *[]byte) {

	if cap(*b) < MSG_HEADER_LENGTH {
		*b = make([]byte, MSG_HEADER_LENGTH)
	} else {
		*b = (*b)[:MSG_HEADER_LENGTH]
	}

	buffer := *b
	//packet header
	buffer[0] = 'T'
	buffer[1] = 0
	buffer[2] = 0
	buffer[3] = 0
}
