package network

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	MSG_KEY_LENGTH         = 32
	MSG_HEADER_LENGTH      = 4
	MSG_ID_LENGTH          = 4
	MSG_MAX_BODY_LENGTH    = 8096
	MSG_DEFAULT_BUF_LENGTH = 1024
	MSG_DEFAULT_COUNT      = 100
	MSG_COUNT_CHECK_TIME   = 3000
)

var bigEndian = binary.BigEndian

// --------------------------------------------------------------------------------------------------------
// |                                header              |                body                          |
// | type byte | body_length uint16 | packet_flag uint8 | msg_id uint32| msg_data []byte               |
// --------------------------------------------------------------------------------------------------------
var msgHeaderLength int = MSG_HEADER_LENGTH

type transPacket struct {
	MsgType    byte
	BodyLength uint16
	PacketFlag uint8
}

type tcpTransport = TcpConn

func (n *tcpTransport) Write() bool {
	tcpConn := n.conn
	wq := n.writeQueue
	msgList := make([]interface{}, 16)
	for {
		c := wq.Get(msgList, 16)
		for i := uint32(0); i < c; i++ {
			m := msgList[i]
			if m == nil {
				goto writeClose
			}
			wg := m.(ITransportMsg)
			if n.WriteMsgPacket(tcpConn, wg) == false {
				goto writeClose
			}
			msgList[i] = nil
			wg.reset()
		}
	}
writeClose:
	buf := n.writeBuf
	n.writeBuf = nil
	bufferPool.Put(buf)
	return false
}

func (n *tcpTransport) WriteMsgPacket(conn net.Conn, msg ITransportMsg) bool {
	switch msg.GetType() {
	case 'P', 'R':
		tsBuf := msg.(*TransportMsgPack)
		n.packMsgBuf(tsBuf)
	case 'T':
		n.packPingMsg()
	default:
		return false
	}

	buf := n.writeBuf
	rb := buf.ReadBuf(buf.Count())
	_, err := conn.Write(rb)
	if err == nil {
		return true
	}
	return false
}

func (n *tcpTransport) packMsgBuf(msg *TransportMsgPack) {
	mBytes := msg.Buf
	var bodyLength int = len(mBytes) + MSG_ID_LENGTH
	var msgLength int = bodyLength + MSG_HEADER_LENGTH

	buf := n.writeBuf
	buf.Reserve(msgLength)

	var wl int = 0
	wl += buf.WriteUint8(msg.msgType)
	wl += buf.WriteUint32(uint32(bodyLength))
	wl += buf.WriteUint8(0)

	wl += buf.WriteUint32(msg.msgID)
	wl += buf.WriteBytes(mBytes)
}

func (n *tcpTransport) packPingMsg() {
	buf := n.writeBuf
	buf.Reserve(MSG_HEADER_LENGTH)
	buffer := buf.WriteBuf()
	buffer[0] = 'T'
	buffer[1] = 0
	buffer[2] = 0
	buffer[3] = 0
	buf.Write(MSG_HEADER_LENGTH)
}

func (n *tcpTransport) Recv() bool {
	tcpConn := n.conn
	serve := n.serveHandler
	msgPacket := &n.msgPacket

	for {
		msgID, msg, err := n.ReadMsgPacket(tcpConn)
		if err != nil {
			goto waitClose
		}

		nowTick := UnixTS()

		switch msgPacket.MsgType {
		case 'P':
			serve.ServeHandler(n, msgID, msg)
		case 'R':
			serve.ServeRpc(n, msgID, msg)
		case 'T':
			n.Pong(nowTick)
		default:
			goto waitClose
		}
	}

waitClose:
	buf := n.recvBuf
	n.recvBuf = nil
	bufferPool.Put(buf)
	return false
}

func (n *tcpTransport) ReadMsgPacket(conn net.Conn) (ProtoTypeID, []byte, error) {
	buf := n.writeBuf
	buf.Reserve(MSG_HEADER_LENGTH)

	mBytes := buf.WriteBuf()
	var rl int = 0
	rl, err := io.ReadFull(conn, mBytes)

	if err != nil {
		return 0, nil, err
	}

	buf.write(rl)

	msgPacket := &n.msgPacket

	msgPacket.MsgType = mBytes[0]
	msgPacket.BodyLength = bigEndian.Uint16(mBytes[1:])
	msgPacket.PacketFlag = mBytes[3]

	if msgPacket.MsgType == 'T' {
		return 0, nil, nil
	}

	if uint32(msgPacket.BodyLength) >= n.option.msg_max_length {
		return 0, nil, errors.New("msg packet length too long.")
	}

	buf.Reserve(int(msgPacket.BodyLength))

	mBytes = buf.WriteBuf()

	rl, err = io.ReadFull(conn, mBytes)
	if err != nil {
		return 0, nil, err
	}

	buf.write(rl)

	if rl < MSG_ID_LENGTH {
		return 0, nil, errors.New("msg packet length error")
	}

	var msgID ProtoTypeID = 0
	msgID = bigEndian.Uint32(mBytes)
	msgBody := mBytes[MSG_ID_LENGTH:]
	return msgID, msgBody, nil
}
