package network

import (
	"encoding/binary"
	"errors"
	"github.com/Cyinx/einx/slog"
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
)

var bigEndian = binary.BigEndian

// --------------------------------------------------------------------------------------------------------
// |                                header              | body                          |
// | type byte | body_length uint16 | packet_flag uint8 | msg_id uint32| msg_data []byte|
// --------------------------------------------------------------------------------------------------------
var msg_header_length int = MSG_HEADER_LENGTH

type transPacket struct {
	MsgType    byte
	BodyLength uint16
	PacketFlag uint8
}

type tcpTransport = TcpConn

func (t *tcpTransport) reseveWriteBuf(size int) {
	if cap(t.write_buf) < int(size) {
		buf := t.write_buf
		t.write_buf = make([]byte, size)
		buffer_pool.Put(buf)
	} else {
		t.write_buf = t.write_buf[0:size]
	}
}

func (t *tcpTransport) Write() bool {
	tcp_conn := t.conn
	write_queue := t.write_queue
	msg_list := make([]interface{}, 16)
	for {
		c := write_queue.Get(msg_list, 16)
		for i := uint32(0); i < c; i++ {
			write_msg := msg_list[i].(*WriteWrapper)
			if write_msg == nil {
				goto write_close
			}
			if t.WriteMsgPacket(tcp_conn, write_msg) == false {
				goto write_close
			}
			msg_list[i] = nil
			write_msg.reset()
			write_pool.Put(write_msg)
		}
	}
write_close:
	buf := t.write_buf
	t.write_buf = nil
	buffer_pool.Put(buf)
	return false
}

func (t *tcpTransport) WriteMsgPacket(conn net.Conn, msg *WriteWrapper) bool {
	switch msg.msg_type {
	case 'P', 'R':
		t.packMsgBuf(msg)
	case 'T':
		t.packPingMsg()
	default:
		return false
	}

	_, err := conn.Write(t.write_buf)
	if err == nil {
		return true
	}
	return false
}

func (t *tcpTransport) packMsgBuf(msg *WriteWrapper) {
	msg_buf := msg.buffer
	var msg_body_length int = len(msg_buf) + MSG_ID_LENGTH
	var msg_length int = msg_body_length + MSG_HEADER_LENGTH

	t.reseveWriteBuf(msg_length)

	buffer := t.write_buf
	buffer[0] = msg.msg_type
	bigEndian.PutUint16(buffer[1:], uint16(msg_body_length))
	buffer[3] = 0

	bigEndian.PutUint32(buffer[4:], msg.msg_id)

	copy(buffer[MSG_HEADER_LENGTH+MSG_ID_LENGTH:], msg_buf)
}

func (t *tcpTransport) packPingMsg() {
	t.reseveWriteBuf(MSG_HEADER_LENGTH)
	buffer := t.write_buf
	buffer[0] = 'T'
	buffer[1] = 0
	buffer[2] = 0
	buffer[3] = 0
}

func (t *tcpTransport) Recv() bool {

	tcp_conn := t.conn
	serve := t.servehander
	msg_packet := &t.msg_packet

	msg_recv_count := int64(0)
	check_duration := PONGTIME / 1000

	for {
		msg_id, msg, err := t.ReadMsgPacket(tcp_conn)
		if err != nil {
			goto wait_close
		}

		switch msg_packet.MsgType {
		case 'P':
			serve.ServeHandler(t, msg_id, msg)
			msg_recv_count++
		case 'R':
			serve.ServeRpc(t, msg_id, msg)
			msg_recv_count++
		case 'T':
			t.OnPing()
		default:
			goto wait_close
		}

		nowTime := GetNowTick()
		duration := nowTime - t.recv_check_time

		if duration < check_duration {
			continue
		}

		max_msg_count := duration * int64(t.option.msg_max_count)

		if msg_recv_count >= max_msg_count {
			slog.LogError("tcp_conn", "tcp conn [%v] recv beyond max msg count. recv msg count [%v].duration [%v].option count [%v]",
				t.RemoteAddr().String(), msg_recv_count, duration, t.option.msg_max_count)
			goto wait_close
		}

		msg_recv_count = 0
		t.recv_check_time = nowTime
	}

wait_close:
	buf := t.recv_buf
	t.recv_buf = nil
	buffer_pool.Put(buf)
	return false
}

func (t *tcpTransport) reseveRecvBuf(size int) {
	if cap(t.recv_buf) < int(size) {
		buf := t.recv_buf
		t.recv_buf = make([]byte, size)
		buffer_pool.Put(buf)
	} else {
		t.recv_buf = t.recv_buf[0:size]
	}
}

func (t *tcpTransport) ReadMsgPacket(conn net.Conn) (ProtoTypeID, []byte, error) {

	t.reseveRecvBuf(MSG_HEADER_LENGTH)

	if _, err := io.ReadFull(conn, t.recv_buf); err != nil {
		return 0, nil, err
	}

	msg_packet := &t.msg_packet

	msg_packet.MsgType = t.recv_buf[0]
	msg_packet.BodyLength = bigEndian.Uint16(t.recv_buf[1:])
	msg_packet.PacketFlag = t.recv_buf[3]

	if msg_packet.MsgType == 'T' {
		return 0, nil, nil
	}

	if msg_packet.BodyLength >= t.option.msg_max_length {
		return 0, nil, errors.New("msg packet length too long.")
	}

	t.reseveRecvBuf(int(msg_packet.BodyLength))

	if _, err := io.ReadFull(conn, t.recv_buf); err != nil {
		return 0, nil, err
	}

	if len(t.recv_buf) < MSG_ID_LENGTH {
		return 0, nil, errors.New("msg packet length error")
	}

	var msg_id ProtoTypeID = 0
	msg_id = bigEndian.Uint32(t.recv_buf)
	msg_body := t.recv_buf[MSG_ID_LENGTH:]
	return msg_id, msg_body, nil
}
