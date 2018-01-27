package module

import (
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"github.com/Cyinx/einx/timer"
	"sync"
	"time"
)

const MAIN_MODULE string = "Main"
const RPC_CHAN_LENGTH = 512

type Agent = agent.Agent
type EventMsg = event.EventMsg
type EventType = event.EventType
type SessionEventMsg = event.SessionEventMsg
type DataEventMsg = event.DataEventMsg
type RpcEventMsg = event.RpcEventMsg
type EventQueue = event.EventQueue
type TimerHandler = timer.TimerHandler
type TimerManager = timer.TimerManager
type ProtoTypeID = uint16
type ModuleID = int
type DispatchHandler func(event_msg EventMsg)
type MsgHandler func(Agent, interface{})
type RpcHandler func(interface{}, []interface{})

type Module interface {
	Close()
	GetName() string
	Run(*sync.WaitGroup)
	RpcCall(string, ...interface{})
	RegisterHandler(ProtoTypeID, MsgHandler)
	RegisterRpcHandler(string, RpcHandler)
	AddTimer(delay uint64, op TimerHandler, args ...interface{}) uint64
}

type ModuleIniter interface {
	SetDispatchHandler(DispatchHandler)
}

type ModuleEventer interface {
	PostEvent(EventType, Agent)
	PostData(EventType, uint16, Agent, interface{})
}

var module_map map[string]Module = make(map[string]Module)
var module_id_index ModuleID = 0

func GetModule(name string) Module {
	var m Module
	var ok bool = false
	m, ok = module_map[name]
	if ok == false {
		m = &module{
			id:              module_id_index,
			ev_queue:        event.NewEventQueue(),
			rpc_chan:        make(chan EventMsg, RPC_CHAN_LENGTH),
			close_chan:      make(chan bool),
			dispatch_event:  nil,
			name:            name,
			timer_manager:   timer.NewTimerManager(),
			msg_handler_map: make(map[ProtoTypeID]MsgHandler),
			rpc_handler_map: make(map[string]RpcHandler),
			rpc_msg_pool:    &sync.Pool{New: func() interface{} { return new(RpcEventMsg) }},
			event_msg_pool:  &sync.Pool{New: func() interface{} { return new(SessionEventMsg) }},
			data_msg_pool:   &sync.Pool{New: func() interface{} { return new(DataEventMsg) }},
		}
		module_map[name] = m
		module_id_index = module_id_index + 1
	}
	return m
}

type module struct {
	id              ModuleID
	ev_queue        *EventQueue
	rpc_chan        chan EventMsg
	close_chan      chan bool
	dispatch_event  DispatchHandler
	name            string
	msg_handler_map map[ProtoTypeID]MsgHandler
	rpc_handler_map map[string]RpcHandler
	rpc_msg_pool    *sync.Pool
	event_msg_pool  *sync.Pool
	data_msg_pool   *sync.Pool
	timer_manager   *TimerManager
}

func (this *module) GetName() string {
	return this.name
}
func (this *module) PushEventMsg(ev EventMsg) {
	this.ev_queue.Push(ev)
}

func (this *module) Close() {
	slog.LogWarning("module", "module [%s] will close!", this.name)
	this.close_chan <- true
}

func (this *module) AddTimer(delay uint64, op TimerHandler, args ...interface{}) uint64 {
	return this.timer_manager.AddTimer(delay, op, args)
}

func (this *module) PostEvent(event_type EventType, agent Agent) {
	event := this.event_msg_pool.Get().(*SessionEventMsg)
	event.MsgType = event_type
	event.Sender = agent
	this.ev_queue.Push(event)
}

func (this *module) PostData(event_type EventType, type_id uint16, agent Agent, data interface{}) {
	event := this.data_msg_pool.Get().(*DataEventMsg)
	event.MsgType = event_type
	event.Sender = agent
	event.TypeID = type_id
	event.MsgData = data
	this.ev_queue.Push(event)
}

func (this *module) RpcCall(name string, args ...interface{}) {
	rpc_msg := this.rpc_msg_pool.Get().(*RpcEventMsg)
	rpc_msg.MsgType = event.EVENT_MODULE_RPC
	rpc_msg.Sender = this
	rpc_msg.Data = args
	rpc_msg.RpcName = name
	this.rpc_chan <- rpc_msg
}

func (this *module) SetDispatchHandler(handler DispatchHandler) {
	this.dispatch_event = handler
}

func (this *module) RegisterHandler(type_id ProtoTypeID, handler MsgHandler) {
	_, ok := this.msg_handler_map[type_id]
	if ok == true {
		slog.LogWarning("module", "MsgID[%d] has been registered", type_id)
		return
	}
	this.msg_handler_map[type_id] = handler
}

func (this *module) RegisterRpcHandler(rpc_name string, handler RpcHandler) {
	_, ok := this.rpc_handler_map[rpc_name]
	if ok == true {
		slog.LogWarning("module", "Rpc[%s] has been registered", rpc_name)
		return
	}
	this.rpc_handler_map[rpc_name] = handler
}

func (this *module) Run(wait *sync.WaitGroup) {
	wait.Add(1)
	ev_queue := this.ev_queue
	rpc_chan := this.rpc_chan
	timer_manager := this.timer_manager
	var event_msg EventMsg = nil
	var rpc_msg EventMsg = nil
	var event_chan = ev_queue.GetChan()
	var close_flag bool = false
	var ticker = time.NewTicker(15 * time.Millisecond)
	for {

		select {
		case close_flag = <-this.close_chan:
			if close_flag == true {
				goto run_close
			}
		case rpc_msg = <-rpc_chan:
			this.handle_rpc(rpc_msg)
		case event_msg = <-event_chan:
			this.handle_event(event_msg)
		case <-ticker.C:
			timer_manager.Execute(100)
		}
	}
run_close:
	wait.Done()
}

func (this *module) handle_event(event_msg EventMsg) {
	if event_msg.GetType() == event.EVENT_TCP_READ_MSG {
		data_event := event_msg.(*DataEventMsg)
		handler, ok := this.msg_handler_map[data_event.TypeID]
		if ok == true {
			handler(data_event.Sender, data_event.MsgData)
			this.data_msg_pool.Put(data_event)
			return
		}
		slog.LogError("module", "module unregister msg handler!")
		this.data_msg_pool.Put(data_event)
		return
	}
	if this.dispatch_event == nil {
		slog.LogError("module", "module unknow event msg")
		this.event_msg_pool.Put(event_msg)
		return
	}
	this.dispatch_event(event_msg)
	this.event_msg_pool.Put(event_msg)
}

func (this *module) handle_rpc(event_msg EventMsg) {
	rpc_msg := event_msg.(*RpcEventMsg)
	handler, ok := this.rpc_handler_map[rpc_msg.RpcName]
	if ok == true {
		handler(rpc_msg.Sender, rpc_msg.Data)
		this.rpc_msg_pool.Put(rpc_msg)
		return
	}
	slog.LogError("module", "module unregister rpc handler!")
	this.rpc_msg_pool.Put(rpc_msg)
	return
}
