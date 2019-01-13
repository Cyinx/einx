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
	option        TransportOption
}

func NewTcpClientMgr(opts ...Option) Component {
	tcp_client := &TcpClientMgr{
		component_id: GenComponentID(),
		option: TransportOption{
			msg_max_length: MSG_MAX_BODY_LENGTH,
			msg_max_count:  MSG_DEFAULT_COUNT,
		},
	}

	for _, opt := range opts {
		opt(tcp_client)
	}

	if tcp_client.agent_handler == nil {
		panic("option agent handler is nil")
	}

	if tcp_client.module == nil {
		panic("option agent handler is nil")
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
		e.Attach = user_type
		e.Err = err
		this.module.PushEventMsg(e)
		return
	}

	m := this.module
	h := this.agent_handler

	tcp_agent := newTcpConn(raw_conn, h, Linker_TCP_OutGoing, &this.option)
	tcp_agent.SetUserType(user_type)
	m.PostEvent(event.EVENT_TCP_CONNECTED, tcp_agent, this.component_id)

	go func() {
		ping_mgr.AddPing(tcp_agent)
		tcp_agent.Run()
		ping_mgr.RemovePing(tcp_agent)
		m.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent, this.component_id)
	}()
}
