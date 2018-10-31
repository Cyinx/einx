package network

import (
	"github.com/Cyinx/einx/agent"
	//"github.com/Cyinx/einx/slog"
	"io"
	"net"
	"sync/atomic"
)

type TcpConn struct {
	agent_id       AgentID
	conn           net.Conn
	close_flag     uint32
	write_chan     chan *WriteWrapper
	write_stop     chan struct{}
	handler        SessionHandler
	last_ping_tick int64
	remote_addr    string
	agent_type     int16
	user_type      int16
}

func NewTcpConn(raw_conn net.Conn, h SessionHandler, agent_type int16) NetLinker {
	tcp_agent := &TcpConn{
		conn:           raw_conn,
		close_flag:     0,
		write_chan:     make(chan *WriteWrapper, 256),
		write_stop:     make(chan struct{}),
		agent_id:       agent.GenAgentID(),
		handler:        h,
		last_ping_tick: NowKeepAliveTick,
		remote_addr:    raw_conn.RemoteAddr().(*net.TCPAddr).String(),
		agent_type:     agent_type,
		user_type:      0,
	}
	return tcp_agent
}

func (this *TcpConn) GetID() AgentID {
	return this.agent_id
}

func (this *TcpConn) GetType() int16 {
	return this.agent_type
}

func (this *TcpConn) GetUserType() int16 {
	return this.user_type
}

func (this *TcpConn) SetUserType(t int16) {
	this.user_type = t
}

func (this *TcpConn) ReadMsg() ([]byte, error) {
	return nil, nil
}

func (this *TcpConn) IsClosed() bool {
	close_flag := atomic.LoadUint32(&this.close_flag)
	return close_flag == 1
}

func (this *TcpConn) do_push_write(wrapper *WriteWrapper) bool {
	select {
	case <-this.write_stop:
		return false
	case this.write_chan <- wrapper:
	}
	return true
}

func (this *TcpConn) WriteMsg(msg_id ProtoTypeID, b []byte) bool {
	if this.IsClosed() == true {
		return false
	}

	wrapper := &WriteWrapper{
		msg_type: 'P',
		msg_id:   msg_id,
		buffer:   b,
	}
	return this.do_push_write(wrapper)
}

func (this *TcpConn) RpcCall(msg_id ProtoTypeID, b []byte) bool {
	if this.IsClosed() == true {
		return false
	}

	wrapper := &WriteWrapper{
		msg_type: 'R',
		msg_id:   msg_id,
		buffer:   b,
	}

	return this.do_push_write(wrapper)
}

func (this *TcpConn) LocalAddr() net.Addr {
	return nil
}

func (this *TcpConn) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *TcpConn) Close() {
	if atomic.CompareAndSwapUint32(&this.close_flag, 0, 1) == true {
		this.write_chan <- nil
		this.conn.Close()
		this.conn = nil
		this.Destroy()
	}
}

func (this *TcpConn) Destroy() {

}

func (this *TcpConn) Run() {
	go this.WriteGoroutine()

	tcp_conn := this.conn
	var packet PacketHeader
	header_buffer := make([]byte, MSG_HEADER_LENGTH)
	body_buffer := make([]byte, MSG_HEADER_LENGTH)
	h := this.handler

	for {
		header_buffer = header_buffer[0:]
		msg_id, msg, err := ReadMsgPacket(tcp_conn, &packet, header_buffer, &body_buffer)
		if err != nil {
			goto wait_close
		}
		switch packet.MsgType {
		case 'P':
			h.ServeHandler(this, msg_id, msg)
			break
		case 'R':
			h.ServeRpc(this, msg_id, msg)
			break
		case 'T':
			this.last_ping_tick = NowKeepAliveTick
			break
		default:
			goto wait_close
		}
	}

wait_close:
	this.Close()
}

func (this *TcpConn) WriteGoroutine() {
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
	close(this.write_stop)
}

func (this *TcpConn) do_write(w io.Writer, msg *WriteWrapper, write_buffer *[]byte) bool {
	switch msg.msg_type {
	case 'P', 'R':
		if MarshalMsgBinary(msg.msg_type, msg.msg_id, msg.buffer, write_buffer) == false {
			return false
		}
	case 'T':
		MarshalKeepAliveMsgBinary(write_buffer)
	default:
		return false
	}
	_, err := w.Write(*write_buffer)
	if err == nil {
		return true
	}
	//slog.LogInfo("tcp", "write msg error %s", err.Error())
	return false
}

func (this *TcpConn) Ping() {
	wrapper := &WriteWrapper{
		msg_type: 'T',
		msg_id:   0,
		buffer:   nil,
	}
	this.do_push_write(wrapper)
}

func (this *TcpConn) Pong() {
	if NowKeepAliveTick-this.last_ping_tick >= PONGTIME {
		this.conn.Close()
	}
}
