package einx

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/network"
	"github.com/Cyinx/einx/slog"
	"sync"
)

type EventMsg = event.EventMsg
type EventType = event.EventType
type Agent = agent.Agent
type AgentID = agent.AgentID
type MsgHandler = module.MsgHandler
type NewAgentHandler func(EventType, AgentID, Agent)
type ProtoTypeID = network.ProtoTypeID
type Module = module.Module
type SessionEventMsg = event.SessionEventMsg
type einx struct {
	main_module  Module
	server_map   sync.Map
	client_map   sync.Map
	agent_map    map[AgentID]Agent
	module_map   sync.Map
	agent_id     AgentID
	on_new_agent NewAgentHandler
	end_wait     sync.WaitGroup
	close_chan   chan bool
}

func (this *einx) add_tcp_server(addr string) {
	tcp_server := network.NewTcpServer()
	tcp_server.Start(addr)

	this.server_map.Store(addr, tcp_server)
}

func (this *einx) start_tcp_client(addr string) {
	tcp_client := network.NewTcpClient(addr)
	tcp_client.Start()

	this.client_map.Store(addr, tcp_client)
}

func (this *einx) start_run_module(module Module) {
	this.module_map.Store(module.GetName(), module)
	go func() {
		module.Run(&this.end_wait)
	}()
}

func (this *einx) do_close() {

	//shutdown server
	this.server_map.Range(func(key, value interface{}) bool {
		s := value.(network.Server)
		s.Close()
		return true
	})

	this.module_map.Range(func(key, value interface{}) bool {
		if key != module.MAIN_MODULE {
			m := value.(Module)
			m.Close()
		}
		return true
	})
	this.main_module.Close()
}

func (this *einx) close() {
	this.end_wait.Wait()
}

func (this *einx) einx_loop() {
	this.end_wait.Add(1)
	main_module := this.main_module
	module_init := main_module.(module.ModuleIniter)
	module_init.SetDispatchHandler(this.dispatch_event)
	main_module.Run(&this.end_wait)
	this.end_wait.Done()
}

func (this *einx) dispatch_event(event_msg EventMsg) {
	session_event := event_msg.(*SessionEventMsg)
	switch session_event.MsgType {
	case event.EVENT_TCP_ACCEPTED:
		this.on_new_network_agent(session_event.MsgType, session_event)
	case event.EVENT_TCP_CONNECTED:
		this.on_new_network_agent(session_event.MsgType, session_event)
	case event.EVENT_TCP_CLOSED:
		this.on_closed_network_agent(session_event.MsgType, session_event)
	default:
		slog.LogError("einx", "unknow event msg")
	}
}

func (this *einx) on_new_network_agent(event_type EventType, event_msg *SessionEventMsg) {
	agent := event_msg.Sender.(Agent)
	this.agent_id++
	agent.SetID(this.agent_id)
	//this.agent_map.Store(this.agent_id, agent)
	this.agent_map[this.agent_id] = agent
	this.on_new_agent(event_type, this.agent_id, agent)
}

func (this *einx) on_closed_network_agent(event_type EventType, event_msg *SessionEventMsg) {
	agent := event_msg.Sender
	slog.LogInfo("einx", "delete agent")
	//this.agent_map.Delete(agent.GetID())
	delete(this.agent_map, agent.GetID())
}
