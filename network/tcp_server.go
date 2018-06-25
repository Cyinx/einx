package network

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"net"
	"sync/atomic"
	"time"
)

const TCP_ACCEPT_SLEEP = 150

type TcpServerMgr struct {
	listener      net.Listener
	component_id  ComponentID
	module        ModuleEventer
	agent_handler AgentHandler
	addr          string
	close_flag    int32
}

func NewTcpServerMgr(addr string, m ModuleEventer, h AgentHandler) Component {
	tcp_server := &TcpServerMgr{
		component_id:  GenComponentID(),
		addr:          addr,
		module:        m,
		agent_handler: h,
		close_flag:    0,
	}
	return tcp_server
}

func (this *TcpServerMgr) GetID() ComponentID {
	return this.component_id
}

func (this *TcpServerMgr) GetType() ComponentType {
	return COMPONENT_TYPE_TCP_SERVER
}

func (this *TcpServerMgr) Start() {

	listener, err := net.Listen("tcp", this.addr)
	if err != nil {
		slog.LogError("tcp_server", "ListenTCP addr:[%s],Error:%s", this.addr, err.Error())
		return
	}
	this.listener = listener
	go this.do_tcp_accept()
}

func (this *TcpServerMgr) Close() {
	if atomic.CompareAndSwapInt32(&this.close_flag, 0, 1) == true {
		this.listener.Close()
	}
}

func (this *TcpServerMgr) isRunning() bool {
	close_flag := atomic.LoadInt32(&this.close_flag)
	return close_flag == 0
}

func (this *TcpServerMgr) do_tcp_accept() {
	m := this.module
	h := this.agent_handler
	listener := this.listener

	for this.isRunning() {
		raw_conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(TCP_ACCEPT_SLEEP)
			} else if this.isRunning() {
				slog.LogError("tcp_server", "Accept Error: %v", err)
			}
			continue
		}

		tcp_agent := NewTcpConn(raw_conn, m, h, AgentType_TCP_InComming)
		m.PostEvent(event.EVENT_TCP_ACCEPTED, tcp_agent, this.component_id)

		go func() {
			AddPong(tcp_agent.(*TcpConn))
			tcp_agent.Run()
			RemovePong(tcp_agent.(*TcpConn))
			m.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent, this.component_id)
		}()
	}
}
