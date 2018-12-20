package einx

import (
	"github.com/Cyinx/einx/console"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/lua"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/network"
	"github.com/Cyinx/einx/slog"
	"os"
	"os/signal"
)

var _einx_default = &einx{
	close_chan: make(chan bool),
}

func Init(opts ...Option) {
	for _, opt := range opts {
		opt()
	}
}

func Run() {
	slog.Run()
	console.Run()
	network.Run()
	module.Start()
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

func GetModule(name string) Module {
	return module.GetModule(name)
}

func NewLuaStae() *lua_state.LuaRuntime {
	return lua_state.NewLuaStae()
}

func AddTcpServerMgr(m module.Module, addr string, mgr interface{}) {
	m_eventer := m.(event.EventReceiver)
	tcp_server := network.NewTcpServerMgr(
		NetworkOption.ListenAddr(addr),
		network.Module(m_eventer),
		NetworkOption.ServeHandler(mgr.(SessionHandler)))
	e := &event.ComponentEventMsg{}
	e.MsgType = event.EVENT_COMPONENT_CREATE
	e.Sender = tcp_server
	e.Attach = mgr
	m_eventer.PushEventMsg(e)
}

func StartTcpClientMgr(m module.Module, name string, mgr interface{}) {
	m_eventer := m.(event.EventReceiver)
	tcp_client := network.NewTcpClientMgr(
		NetworkOption.Name(name),
		network.Module(m_eventer),
		NetworkOption.ServeHandler(mgr.(SessionHandler)))

	e := &event.ComponentEventMsg{}
	e.MsgType = event.EVENT_COMPONENT_CREATE
	e.Sender = tcp_client
	e.Attach = mgr
	m_eventer.PushEventMsg(e)
}

func AddModuleComponent(m module.Module, c Component, mgr interface{}) {
	m_eventer := m.(event.EventReceiver)
	e := &event.ComponentEventMsg{}
	e.MsgType = event.EVENT_COMPONENT_CREATE
	e.Sender = c
	e.Attach = mgr
	m_eventer.PushEventMsg(e)
}

func CreateModuleWorkers(name string, size int) WorkerPool {
	return module.CreateWorkers(name, size)
}

func GetWorkerPool(name string) WorkerPool {
	return module.GetWorkerPool(name)
}
