package network

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"io"
	"net"
	"sync/atomic"
)

type Agent = agent.Agent
type AgentID = agent.AgentID

type TcpConnAgent struct {
	agent_id   AgentID
	conn       net.Conn
	close_flag uint32
	write_chan chan *WriteWrapper
	write_stop chan struct{}
}

func NewTcpConnAgent(raw_conn net.Conn) Agent {
	tcp_agent := &TcpConnAgent{
		conn:       raw_conn,
		close_flag: 0,
		write_chan: make(chan *WriteWrapper, 128),
		write_stop: make(chan struct{}),
	}
	return tcp_agent
}

func (this *TcpConnAgent) GetID() AgentID {
	return this.agent_id
}

func (this *TcpConnAgent) SetID(id AgentID) {
	this.agent_id = id
}

func (this *TcpConnAgent) ReadMsg() ([]byte, error) {
	return nil, nil
}

func (this *TcpConnAgent) IsClosed() bool {
	return atomic.CompareAndSwapUint32(&this.close_flag, 1, 1) == true
}

func (this *TcpConnAgent) WriteMsg(msg_id ProtoTypeID, msg interface{}) bool {
	if this.IsClosed() == true {
		return false
	}

	msg_buffer, err := MsgProtoMarshal(msg)
	if err != nil {
		return false
	}

	wrapper := &WriteWrapper{
		msg_id: msg_id,
		buffer: msg_buffer,
	}

	select {
	case <-this.write_stop:
		return false
	case this.write_chan <- wrapper:
	}
	return true
}

func (this *TcpConnAgent) LocalAddr() net.Addr {
	return nil
}

func (this *TcpConnAgent) RemoteAddr() net.Addr {
	return nil
}

func (this *TcpConnAgent) Close() {
	if atomic.CompareAndSwapUint32(&this.close_flag, 0, 1) == true {
		this.write_chan <- nil
		this.conn.Close()
		this.conn = nil
	}
}

func (this *TcpConnAgent) Destroy() {

}

func (this *TcpConnAgent) Run() {
	go this.WriteGoroutine()

	tcp_conn := this.conn
	var packet PacketHeader
	header_buffer := make([]byte, MSG_HEADER_LENGTH)
	body_buffer := make([]byte, MSG_HEADER_LENGTH)
	for {
		header_buffer = header_buffer[0:]
		msg_id, msg, err := ReadMsgPacket(tcp_conn, &packet, header_buffer, &body_buffer)
		if err != nil {
			slog.LogWarning("tcp", "read msg packet error : %s", err.Error())
			goto wait_close
		}
		switch packet.MsgType {
		case 'P':
			_event_module.PostData(event.EVENT_TCP_READ_MSG, msg_id, this, msg)
		case 'R':
			msg = nil
		default:
			goto wait_close
		}
	}

wait_close:
	slog.LogInfo("tcp", "Recv Close")
	this.Close()

}

func (this *TcpConnAgent) WriteGoroutine() {
	write_buffer := make([]byte, 512)
	tcp_conn := this.conn

	for msg := range this.write_chan {
		if msg == nil || this.IsClosed() == true {
			goto wait_close
		}

		write_buffer = write_buffer[0:]
		if this.do_write(tcp_conn, msg, &write_buffer) == true {
			continue
		}
	}

wait_close:
	slog.LogInfo("tcp", "Writer Close")
	close(this.write_stop)
}

func (this *TcpConnAgent) do_write(w io.Writer, msg *WriteWrapper, write_buffer *[]byte) bool {
	if MarshalMsgBinary(msg.msg_id, msg.buffer, write_buffer) == true {
		_, err := w.Write(*write_buffer)
		if err == nil {
			return true
		}
		slog.LogInfo("tcp", "write msg error %s", err.Error())
	}
	return false
}
