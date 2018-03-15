package network

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"net"
)

type TcpServerCom struct {
	listener     net.Listener
	component_id ComponentID
	module       ModuleEventer
	addr         string
}

func NewTcpServerCom(addr string, m ModuleEventer) Component {
	tcp_server := &TcpServerCom{
		component_id: GenComponentID(),
		addr:         addr,
		module:       m,
	}
	return tcp_server
}

func (this *TcpServerCom) GetID() ComponentID {
	return this.component_id
}

func (this *TcpServerCom) GetType() ComponentType {
	return ServerType_TCP
}

func (this *TcpServerCom) Start() {

	listener, err := net.Listen("tcp", this.addr)
	if err != nil {
		slog.LogError("tcp_server", "ListenTCP addr:[%s],Error:%s", this.addr, err.Error())
		return
	}
	this.listener = listener
	go this.do_tcp_accept()
}

func (this *TcpServerCom) Close() {
	this.listener.Close()
}

func (this *TcpServerCom) do_tcp_accept() {

	for {
		raw_conn, err := this.listener.Accept()
		if err != nil {
			slog.LogError("tcp_server", "Accept Error:%s", err.Error())
			return
		}

		tcp_agent := NewTcpConn(raw_conn, this.module)
		this.module.PostEvent(event.EVENT_TCP_ACCEPTED, tcp_agent, this.component_id)
		go func() {
			AddPong(tcp_agent.(*TcpConn))
			tcp_agent.Run()
			RemovePong(tcp_agent.(*TcpConn))
			this.module.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent, this.component_id)
		}()
	}
}
