package einx

import (
	"github.com/Cyinx/einx/console"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/slog"
	"os"
	"os/signal"
)

const MAIN_MODULE string = module.MAIN_MODULE

var _einx_default = &einx{
	main_module: module.GetModule(MAIN_MODULE),
	close_chan:  make(chan bool),
	agent_map:   make(map[AgentID]Agent),
}

func GetModule(name string) Module {
	return module.GetModule(name)
}

func AddTcpServer(addr string) {
	_einx_default.add_tcp_server(addr)
}

func StartTcpClient(addr string) {
	_einx_default.start_tcp_client(addr)
}

func RegisterNewAgentHandler(handler NewAgentHandler) {
	_einx_default.on_new_agent = handler
}

func StartRunModules(args ...string) {
	for _, name := range args {
		module := module.GetModule(name)
		_einx_default.start_run_module(module)
	}
}

func Run() {
	go _einx_default.einx_loop()
	go console.Run()

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
