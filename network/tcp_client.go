package network

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"net"
)

type TcpClientCom struct {
	name         string
	component_id ComponentID
	module       ModuleEventer
}

func NewTcpClientCom(name string, m ModuleEventer) Component {
	tcp_client := &TcpClientCom{
		name:         name,
		module:       m,
		component_id: GenComponentID(),
	}
	return tcp_client
}

func (this *TcpClientCom) GetID() ComponentID {
	return this.component_id
}

func (this *TcpClientCom) GetType() ComponentType {
	return ClientType_TCP
}

func (this *TcpClientCom) Start() {

}

func (this *TcpClientCom) Close() {

}

func (this *TcpClientCom) Connect(addr string) {
	go this.connect(addr)
}

func (this *TcpClientCom) connect(addr string) {
	raw_conn, err := net.Dial("tcp", addr)
	if err != nil {
		slog.LogWarning("tcp_client", "tcp connect failed %v", err)
		this.module.PostEvent(event.EVENT_TCP_CONNECT_FAILED, nil, this.component_id)
		return
	}

	tcp_agent := NewTcpConn(raw_conn, this.module)
	this.module.PostEvent(event.EVENT_TCP_CONNECTED, tcp_agent, this.component_id)

	go func() {
		AddPing(tcp_agent.(*TcpConn))
		tcp_agent.Run()
		RemovePing(tcp_agent.(*TcpConn))
		this.module.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent, this.component_id)
	}()
}
