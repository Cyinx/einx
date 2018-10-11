package module

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/timer"
	"sync"
)

//var module_map map[string]Module = make(map[string]Module)
var module_map sync.Map
var wait_close sync.WaitGroup

func GenModuleID() AgentID {
	return agent.GenAgentID()
}

func Close() {
	module_map.Range(func(k interface{}, m interface{}) bool {
		m.(ModuleWoker).Close()
		return true
	})

	worker_pools_map.Range(func(k interface{}, m interface{}) bool {
		m.(*ModuleWorkerPool).Close()
		return true
	})

	wait_close.Wait()
}

func NewModule(name string) Module {
	m := &module{
		id:              GenModuleID(),
		ev_queue:        event.NewEventQueue(),
		rpc_queue:       event.NewEventQueue(),
		name:            name,
		timer_manager:   timer.NewTimerManager(),
		msg_handler_map: make(map[ProtoTypeID]MsgHandler),
		rpc_handler_map: make(map[string]RpcHandler),
		agent_map:       make(map[AgentID]Agent),
		commgr_map:      make(map[ComponentID]ComponentMgr),
		component_map:   make(map[ComponentID]Component),
		rpc_msg_pool:    &sync.Pool{New: func() interface{} { return new(RpcEventMsg) }},
		data_msg_pool:   &sync.Pool{New: func() interface{} { return new(DataEventMsg) }},
		event_msg_pool:  &sync.Pool{New: func() interface{} { return new(SessionEventMsg) }},
		close_chan:      make(chan bool),
	}
	m.context = &ModuleContext{m: m}
	return m
}

func GetModule(name string) Module {
	var m Module
	v, ok := module_map.Load(name)
	if ok == false {
		m = NewModule(name)
		module_map.Store(name, m)
	} else {
		m = v.(Module)
	}
	return m
}

func FindModule(name string) Module {
	var m Module
	v, ok := module_map.Load(name)
	if ok == true {
		m = v.(Module)
		return m
	}
	return nil
}

func Start() {
	module_map.Range(func(k interface{}, m interface{}) bool {
		go func(m interface{}) { m.(ModuleWoker).Run(&wait_close) }(m)
		return true
	})

	worker_pools_map.Range(func(k interface{}, m interface{}) bool {
		m.(*ModuleWorkerPool).Start()
		return true
	})
}
