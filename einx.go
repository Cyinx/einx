package einx

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/component"
	"github.com/Cyinx/einx/context"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/lua"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/network"
	"github.com/Cyinx/einx/timer"
	"sync"
)

type Agent = agent.Agent
type AgentID = agent.AgentID
type Module = context.Module
type Context = context.Context
type EventMsg = event.EventMsg
type EventType = event.EventType
type ArgsVar = module.ArgsVar
type MsgHandler = module.MsgHandler
type RpcHandler = module.RpcHandler
type WorkerPool = module.WorkerPool
type Component = component.Component
type ComponentID = component.ComponentID
type ModuleRouter = module.ModuleRouter
type ComponentMgr = module.ComponentMgr
type SessionEventMsg = event.SessionEventMsg
type LuaRuntime = lua_state.LuaRuntime
type NetLinker = network.NetLinker
type ProtoTypeID = network.ProtoTypeID
type SessionMgr = network.SessionMgr
type SessionHandler = network.SessionHandler
type ITcpClientMgr = network.ITcpClientMgr
type ITcpServerMgr = network.ITcpServerMgr
type TimerHandler = timer.TimerHandler
type EventReceiver = event.EventReceiver
type ITranMsgMultiple = network.ITranMsgMultiple

type einx struct {
	endWait   sync.WaitGroup
	closeChan chan bool
	onClose   func()
}

func (e *einx) doClose() {
	onClose := e.onClose
	if onClose != nil {
		onClose()
		<-e.closeChan
	}
}

func (e *einx) close() {
	module.Close()
	e.endWait.Wait()
}

func (e *einx) continueClose() {
	e.closeChan <- true
}
