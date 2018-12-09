package network

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"net"
)

type TcpClientMgr struct {
	name          string
	component_id  ComponentID
	module        EventReceiver
	agent_handler SessionHandler
}

func NewTcpClientMgr(name string, m EventReceiver, h SessionHandler) Component {
	tcp_client := &TcpClientMgr{
		name:          name,
		module:        m,
		component_id:  GenComponentID(),
		agent_handler: h,
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
		e := &event.ComponentEventMsg{}
		e.MsgType = event.EVENT_COMPONENT_ERROR
		e.Sender = this
		e.Attach = err
		this.module.PushEventMsg(e)
		return
	}

	m := this.module
	h := this.agent_handler

	tcp_linker := NewTcpConn(raw_conn, h, AgentType_TCP_OutGoing)
	tcp_linker.SetUserType(user_type)
	m.PostEvent(event.EVENT_TCP_CONNECTED, tcp_linker.(Agent), this.component_id)

	go func() {
		AddPing(tcp_linker.(*TcpConn))
		tcp_linker.Run()
		RemovePing(tcp_linker.(*TcpConn))
		m.PostEvent(event.EVENT_TCP_CLOSED, tcp_linker.(Agent), this.component_id)
	}()
}
