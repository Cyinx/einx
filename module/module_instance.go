package module

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/timer"
	"sync"
	"sync/atomic"
)

//var module_map map[string]Module = make(map[string]Module)
var module_map sync.Map
var module_id_index ModuleID = 0
var wait_close sync.WaitGroup

func GenModuleID() ModuleID {
	return atomic.AddUint32(&module_id_index, 1)
}

func Close() {
	module_map.Range(func(k interface{}, m interface{}) bool {
		m.(Module).Close()
		return true
	})
	wait_close.Wait()
}

func GetModule(name string) Module {
	var m Module
	v, ok := module_map.Load(name)
	if ok == false {
		m = &module{
			id:              GenModuleID(),
			ev_queue:        event.NewEventQueue(),
			rpc_chan:        make(chan EventMsg, RPC_CHAN_LENGTH),
			name:            name,
			timer_manager:   timer.NewTimerManager(),
			msg_handler_map: make(map[ProtoTypeID]MsgHandler),
			rpc_handler_map: make(map[string]RpcHandler),
			agent_map:       make(map[AgentID]Agent),
			sesmgr_map:      make(map[ComponentID]AgentSesMgr),
			component_map:   make(map[ComponentID]Component),
			rpc_msg_pool:    &sync.Pool{New: func() interface{} { return new(RpcEventMsg) }},
			data_msg_pool:   &sync.Pool{New: func() interface{} { return new(DataEventMsg) }},
			event_msg_pool:  &sync.Pool{New: func() interface{} { return new(SessionEventMsg) }},
			close_chan:      make(chan bool),
		}
		module_map.Store(name, m)
		go func() { m.(Module).Run(&wait_close) }()
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
