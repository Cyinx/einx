package network

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"net"
)

type TcpClient struct {
	Connect_Addr    string
	ConnectInterval uint16
	close_flag      uint32
	tcp_agent       Agent
	component_id    ComponentID
	module          ModuleEventer
}

func NewTcpClient(addr string, m ModuleEventer) Component {
	tcp_client := &TcpClient{
		Connect_Addr: addr,
		module:       m,
		component_id: GenComponentID(),
	}
	return tcp_client
}

func (this *TcpClient) GetID() ComponentID {
	return this.component_id
}

func (this *TcpClient) GetType() ComponentType {
	return ClientType_TCP
}

func (this *TcpClient) Start() {
	this.connect()
}

func (this *TcpClient) Close() {
	this.tcp_agent.Close()
}

func (this *TcpClient) dial() (net.Conn, error) {
	return net.Dial("tcp", this.Connect_Addr)
}

func (this *TcpClient) connect() {
	raw_conn, err := this.dial()
	if err != nil {
		slog.LogWarning("tcp_client", "tcp connect failed %v", err)
		this.module.PostEvent(event.EVENT_TCP_CONNECT_FAILED, nil, this.component_id)
		return
	}

	tcp_agent := NewTcpConn(raw_conn, this.module)
	this.tcp_agent = tcp_agent
	this.module.PostEvent(event.EVENT_TCP_CONNECTED, tcp_agent, this.component_id)

	go func() {
		AddPing(tcp_agent.(*TcpConn))
		tcp_agent.Run()
		RemovePing(tcp_agent.(*TcpConn))
		this.module.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent, this.component_id)
	}()
}
