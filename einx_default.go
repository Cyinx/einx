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

var _einxDefault = &einx{
	closeChan: make(chan bool),
	onClose:   nil,
}

func Init(opts ...Option) {
	for _, opt := range opts {
		opt()
	}
}

func Run() {
	console.Run()
	network.Run()
	module.Start()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	slog.LogWarning("einx", "einx will close down (signal: %v)", sig)
	_einxDefault.doClose()
}

func ContinueClose() {
	_einxDefault.continueClose()
}

func Close() {
	_einxDefault.close()
	slog.Close()
}

func GetModule(name string) Module {
	return module.GetModule(name)
}

func NewLuaStae() *lua_state.LuaRuntime {
	return lua_state.NewLuaStae()
}

func AddTcpServerMgr(m module.Module, addr string, mgr interface{}, opts ...Option) {
	er := m.(event.EventReceiver)

	opts = append(opts, NetworkOption.ListenAddr(addr))
	opts = append(opts, network.Module(er))
	opts = append(opts, NetworkOption.ServeHandler(mgr.(SessionHandler)))

	tcpServer := network.NewTcpServerMgr(opts...)

	e := &event.ComponentEventMsg{}
	e.MsgType = event.EVENT_COMPONENT_CREATE
	e.Sender = tcpServer
	e.Attach = mgr
	er.PushEventMsg(e)
}

func StartTcpClientMgr(m module.Module, name string, mgr interface{}, opts ...Option) {
	er := m.(event.EventReceiver)

	opts = append(opts, NetworkOption.Name(name))
	opts = append(opts, network.Module(er))
	opts = append(opts, NetworkOption.ServeHandler(mgr.(SessionHandler)))

	tcpClient := network.NewTcpClientMgr(opts...)

	e := &event.ComponentEventMsg{}
	e.MsgType = event.EVENT_COMPONENT_CREATE
	e.Sender = tcpClient
	e.Attach = mgr
	er.PushEventMsg(e)
}

func AddModuleComponent(m module.Module, c Component, mgr interface{}) {
	er := m.(event.EventReceiver)
	e := &event.ComponentEventMsg{}
	e.MsgType = event.EVENT_COMPONENT_CREATE
	e.Sender = c
	e.Attach = mgr
	er.PushEventMsg(e)
}

func CreateModuleWorkers(name string, size int) WorkerPool {
	return module.CreateWorkers(name, size)
}

func GetWorkerPool(name string) WorkerPool {
	return module.GetWorkerPool(name)
}
