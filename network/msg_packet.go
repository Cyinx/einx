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
	MSG_KEY_LENGTH      = 32
	MSG_HEADER_LENGTH   = 4
	MSG_ID_LENGTH       = 4
	MSG_MAX_BODY_LENGTH = 8096
)

var bigEndian = binary.BigEndian

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

func ReadMsgPacket(r io.Reader, msg_packet *PacketHeader, header_buffer []byte, b *[]byte) (ProtoTypeID, []byte, error) {
	if _, err := io.ReadFull(r, header_buffer); err != nil {
		return 0, nil, err
	}

	msg_packet.MsgType = header_buffer[0]
	msg_packet.BodyLength = bigEndian.Uint16(header_buffer[1:])
	msg_packet.PacketFlag = header_buffer[3]

	if msg_packet.MsgType == 'T' {
		return 0, nil, nil
	}

	if msg_packet.BodyLength >= MSG_MAX_BODY_LENGTH {
		return 0, nil, errors.New("msg packet length too long.")
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
	msg_id = bigEndian.Uint32(body)
	msg_body := body[MSG_ID_LENGTH:]
	return msg_id, msg_body, nil
}

func MarshalMsgBinary(msg_type byte, msg_id ProtoTypeID, msg_buffer []byte, b *[]byte) bool {
	var msg_body_length int = len(msg_buffer) + MSG_ID_LENGTH
	var msg_length int = msg_body_length + MSG_HEADER_LENGTH

	if cap(*b) < msg_length {
		*b = make([]byte, msg_length)
	} else {
		*b = (*b)[:msg_length]
	}

	buffer := *b
	//packet header
	buffer[0] = msg_type
	bigEndian.PutUint16(buffer[1:], uint16(msg_body_length))
	buffer[3] = 0
	//msg wrapper
	bigEndian.PutUint32(buffer[4:], msg_id)

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
