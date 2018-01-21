package network

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"net"
	//"sync/atomic"
	"time"
)

type TcpClient struct {
	Connect_Addr    string
	ConnectInterval uint16
	close_flag      uint32
	tcp_agent       Agent
}

func NewTcpClient(addr string) Client {
	tcp_client := &TcpClient{
		Connect_Addr: addr,
	}
	return tcp_client
}

func (this *TcpClient) GetType() ClientType {
	return ClientType_TCP
}

func (this *TcpClient) Start() {
	this.connect()
}

func (this *TcpClient) dial() net.Conn {
	for {
		conn, err := net.Dial("tcp", this.Connect_Addr)
		if err == nil {
			return conn
		}

		slog.LogWarning("tcp_client", "connect to %s error: %s", this.Connect_Addr, err)
		time.Sleep(5 * time.Second)
		continue
	}
}

func (this *TcpClient) connect() {
	raw_conn := this.dial()
	if raw_conn == nil {
		return
	}

	tcp_agent := NewTcpConnAgent(raw_conn)
	this.tcp_agent = tcp_agent
	_event_module.PostEvent(event.EVENT_TCP_CONNECTED, tcp_agent)

	go func() {
		tcp_agent.Run()

		_event_module.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent)

		time.Sleep(5 * time.Second)
		slog.LogWarning("tcp_client", "正在重连接到 %s", this.Connect_Addr)
		this.connect()
	}()
}
