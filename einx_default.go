package einx

import (
	"github.com/Cyinx/einx/console"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/network"
	"github.com/Cyinx/einx/slog"
	"os"
	"os/signal"
)

var _einx_default = &einx{
	close_chan: make(chan bool),
}

func GetModule(name string) Module {
	return module.GetModule(name)
}

func AddTcpServer(m module.Module, addr string, mgr AgentSesMgr) {
	m_eventer := m.(module.ModuleEventer)
	tcp_server := network.NewTcpServer(addr, m_eventer)
	e := &event.ComponentEventMsg{}
	e.MsgType = event.EVENT_COMPONENT_CREATE
	e.Sender = tcp_server
	e.Attach = mgr
	m_eventer.PushEventMsg(e)
}

func StartTcpClient(m module.Module, addr string, mgr AgentSesMgr) {
	m_eventer := m.(module.ModuleEventer)
	tcp_client := network.NewTcpClient(addr, m_eventer)
	e := &event.ComponentEventMsg{}
	e.MsgType = event.EVENT_COMPONENT_CREATE
	e.Sender = tcp_client
	e.Attach = mgr
	m_eventer.PushEventMsg(e)
}

func Run() {
	go console.Run()
	_einx_default.start_run_modules()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	slog.LogWarning("einx", "einx will close down (signal: %v)", sig)
	_einx_default.do_close()
}

func Close() {
	_einx_default.close()
	slog.Close()
}
