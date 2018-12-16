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
	name          string
	listener      net.Listener
	component_id  ComponentID
	module        EventReceiver
	agent_handler SessionHandler
	addr          string
	close_flag    int32
	option        TransportOption
}

func NewTcpServerMgr(opts ...Option) Component {
	tcp_server := &TcpServerMgr{
		component_id: GenComponentID(),
		close_flag:   0,
		option: TransportOption{
			msg_max_length: MSG_MAX_BODY_LENGTH,
			msg_max_count:  MSG_DEFAULT_COUNT,
		},
	}

	for _, opt := range opts {
		opt(tcp_server)
	}

	if tcp_server.agent_handler == nil {
		panic("option agent handler is nil")
	}

	if tcp_server.module == nil {
		panic("option agent handler is nil")
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
		if this.listener != nil {
			this.listener.Close()
		}
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

		tcp_agent := newTcpConn(raw_conn, h, AgentType_TCP_InComming, &this.option)
		m.PostEvent(event.EVENT_TCP_ACCEPTED, tcp_agent, this.component_id)

		go func() {
			AddPong(tcp_agent)
			tcp_agent.Run()
			RemovePong(tcp_agent)
			m.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent, this.component_id)
		}()
	}
}
