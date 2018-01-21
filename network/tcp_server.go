package network

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"net"
)

type TcpServer struct {
	listener net.Listener
}

func NewTcpServer() Server {
	tcp_server := &TcpServer{}
	return tcp_server
}

func (this *TcpServer) GetType() ServerType {
	return ServerType_TCP
}

func (this *TcpServer) Start(addr string) {

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.LogError("tcp_server", "ListenTCP addr:[%s],Error:%s", addr, err.Error())
		return
	}
	this.listener = listener
	go this.do_tcp_accept()
}

func (this *TcpServer) Close() {
	this.listener.Close()
}

func (this *TcpServer) do_tcp_accept() {

	for {
		raw_conn, err := this.listener.Accept()
		if err != nil {
			slog.LogError("tcp_server", "Accept Error:%s", err.Error())
			return
		}

		tcp_agent := NewTcpConnAgent(raw_conn)
		_event_module.PostEvent(event.EVENT_TCP_ACCEPTED, tcp_agent)
		go func() {
			tcp_agent.Run()
			_event_module.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent)
		}()
	}
}
