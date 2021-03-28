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
	name         string
	listener     net.Listener
	componentID  ComponentID
	module       EventReceiver
	agentHandler SessionHandler
	addr         string
	closeFlag    int32
	option       TransportOption
}

func NewTcpServerMgr(opts ...Option) Component {
	tcpServer := &TcpServerMgr{
		componentID: GenComponentID(),
		closeFlag:   0,
		option:      newTransportOption(),
	}

	for _, opt := range opts {
		opt(tcpServer)
	}

	if tcpServer.agentHandler == nil {
		panic("option agent handler is nil")
	}

	if tcpServer.module == nil {
		panic("option agent handler is nil")
	}

	return tcpServer
}

func (this *TcpServerMgr) GetID() ComponentID {
	return this.componentID
}

func (this *TcpServerMgr) GetType() ComponentType {
	return COMPONENT_TYPE_TCP_SERVER
}

func (this *TcpServerMgr) Address() net.Addr {
	if this.listener == nil {
		return nil
	}
	return this.listener.Addr()
}

func (this *TcpServerMgr) Start() bool {
	listener, err := net.Listen("tcp", this.addr)
	if err != nil {
		slog.LogError("tcp_server", "ListenTCP addr:[%s],Error:%s", this.addr, err.Error())
		return false
	}
	this.listener = listener
	go this.doTcpAccept()
	return true
}

func (this *TcpServerMgr) Close() {
	if atomic.CompareAndSwapInt32(&this.closeFlag, 0, 1) == true {
		if this.listener == nil {
			return
		}
		_ = this.listener.Close()
	}
}

func (this *TcpServerMgr) isRunning() bool {
	close_flag := atomic.LoadInt32(&this.closeFlag)
	return close_flag == 0
}

func (this *TcpServerMgr) doTcpAccept() {
	m := this.module
	h := this.agentHandler
	listener := this.listener

	for this.isRunning() {
		rawConn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(TCP_ACCEPT_SLEEP)
			} else if this.isRunning() {
				slog.LogError("tcp_server", "Accept Error: %v", err)
			}
			continue
		}

		tcpAgent := newTcpConn(rawConn, h, Linker_TCP_InComming, &this.option)
		m.PostEvent(event.EVENT_TCP_ACCEPTED, tcpAgent, this.componentID)

		go func() {
			pingMgr.AddPing(tcpAgent)
			err := tcpAgent.Run()
			pingMgr.RemovePing(tcpAgent)
			m.PostEvent(event.EVENT_TCP_CLOSED, tcpAgent, this.componentID, err)
		}()
	}
}

func (this *TcpServerMgr) GetOption() *TransportOption {
	return &this.option
}
