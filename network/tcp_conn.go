package network

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/queue"
	"github.com/Cyinx/einx/slog"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

type TcpConn struct {
	agent_id       AgentID
	conn           net.Conn
	close_flag     uint32
	write_queue    *queue.CondQueue
	servehander    SessionHandler
	last_ping_tick int64
	remote_addr    string
	conn_type      int16
	user_type      int16

	recv_buf        []byte
	write_buf       []byte
	msg_packet      transPacket
	recv_check_time int64
	option          TransportOption
}

func newTcpConn(raw_conn net.Conn, h SessionHandler, conn_type int16, opt *TransportOption) *TcpConn {
	tcp_agent := &TcpConn{
		conn:        raw_conn,
		close_flag:  0,
		write_queue: queue.NewCondQueue(),
		agent_id:    agent.GenAgentID(),
		servehander: h,
		remote_addr: raw_conn.RemoteAddr().(*net.TCPAddr).String(),
		conn_type:   conn_type,
		user_type:   0,

		recv_buf:  buffer_pool.Get().([]byte),
		write_buf: buffer_pool.Get().([]byte),
		option:    *opt,
	}
	return tcp_agent
}

func (this *TcpConn) GetID() AgentID {
	return this.agent_id
}

func (this *TcpConn) GetType() int16 {
	return this.conn_type
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
	this.write_queue.Push(wrapper)
	return true
}

var write_pool *sync.Pool = &sync.Pool{New: func() interface{} { return new(WriteWrapper) }}

func (this *TcpConn) WriteMsg(msg_id ProtoTypeID, b []byte) bool {
	if this.IsClosed() == true {
		return false
	}

	wrapper := write_pool.Get().(*WriteWrapper)
	wrapper.msg_type = 'P'
	wrapper.msg_id = msg_id
	wrapper.buffer = b

	return this.do_push_write(wrapper)
}

func (this *TcpConn) RpcCall(msg_id ProtoTypeID, b []byte) bool {
	if this.IsClosed() == true {
		return false
	}

	wrapper := write_pool.Get().(*WriteWrapper)
	wrapper.msg_type = 'P'
	wrapper.msg_id = msg_id
	wrapper.buffer = b

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
		this.do_push_write(nil)
	}
}

func (this *TcpConn) Destroy() {
	this.conn.Close()
}

func (this *TcpConn) Run() {
	defer this.recover()

	go func() {
		defer this.recover()
		if this.Write() == false {
			this.Close()
			this.Destroy()
		}
	}()

	if this.Recv() == false {
		this.Close()
	}
}

func (this *TcpConn) BeginPing() {
	atomic.StoreInt64(&this.last_ping_tick, GetNowTick())
}

func (this *TcpConn) Pong(now_tick int64) {
	if this.last_ping_tick == now_tick {
		return
	}
	atomic.StoreInt64(&this.last_ping_tick, now_tick)
	if this.conn_type == Linker_TCP_InComming {
		this.DoPing()
	}
}

func (this *TcpConn) Ping() {
	check_duration := PINGTIME / 1000
	if GetNowTick()-this.GetLastPingTime() <= (check_duration * 2) {
		if this.conn_type == Linker_TCP_OutGoing {
			this.DoPing()
		}
		return
	}
	this.conn.Close()
}

func (this *TcpConn) GetLastPingTime() int64 {
	return atomic.LoadInt64(&this.last_ping_tick)
}

func (this *TcpConn) DoPing() {
	wrapper := &WriteWrapper{
		msg_type: 'T',
		msg_id:   0,
		buffer:   nil,
	}
	this.do_push_write(wrapper)
}

func (this *TcpConn) recover() {
	if r := recover(); r != nil {
		slog.LogError("tcp_recovery", "recover error :%v", r)
		slog.LogError("tcp_recovery", "%s", string(debug.Stack()))
		this.Close()
		this.Destroy()
	}
}
