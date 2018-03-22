package network

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"net"
)

type TcpClientMgr struct {
	name         string
	component_id ComponentID
	module       ModuleEventer
}

func NewTcpClientMgr(name string, m ModuleEventer) Component {
	tcp_client := &TcpClientMgr{
		name:         name,
		module:       m,
		component_id: GenComponentID(),
	}
	return tcp_client
}

func (this *TcpClientMgr) GetID() ComponentID {
	return this.component_id
}

func (this *TcpClientMgr) GetType() ComponentType {
	return COMPONENT_TYPE_TCP_CLIENT
}

func (this *TcpClientMgr) Start() {

}

func (this *TcpClientMgr) Close() {

}

func (this *TcpClientMgr) Connect(addr string, user_type int16) {
	go this.connect(addr, user_type)
}

func (this *TcpClientMgr) connect(addr string, user_type int16) {
	raw_conn, err := net.Dial("tcp", addr)
	if err != nil {
		slog.LogWarning("tcp_client", "tcp connect failed %v", err)
		this.module.PostEvent(event.EVENT_TCP_CONNECT_FAILED, nil, this.component_id)
		return
	}

	tcp_agent := NewTcpConn(raw_conn, this.module, AgentType_TCP_OutGoing)
	tcp_agent.SetUserType(user_type)
	this.module.PostEvent(event.EVENT_TCP_CONNECTED, tcp_agent, this.component_id)

	go func() {
		AddPing(tcp_agent.(*TcpConn))
		tcp_agent.Run()
		RemovePing(tcp_agent.(*TcpConn))
		this.module.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent, this.component_id)
	}()
}
