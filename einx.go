package einx

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/component"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/network"
	"sync"
)

type EventMsg = event.EventMsg
type EventType = event.EventType
type Agent = agent.Agent
type AgentID = agent.AgentID
type AgentSessionMgr = agent.AgentSessionMgr
type Component = component.Component
type ComponentID = component.ComponentID
type ComponentMgr = component.ComponentMgr
type MsgHandler = module.MsgHandler
type RpcHandler = module.RpcHandler
type ProtoTypeID = network.ProtoTypeID
type Module = module.Module
type SessionEventMsg = event.SessionEventMsg

type einx struct {
	end_wait   sync.WaitGroup
	close_chan chan bool
}

func (this *einx) start_run_modules() {
	//module.StartRunModules(&this.end_wait)
}

func (this *einx) do_close() {
	//shutdown server
	module.Close()
}

func (this *einx) close() {
	this.end_wait.Wait()
}
