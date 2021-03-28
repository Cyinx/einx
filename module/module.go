package module

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/component"
	"github.com/Cyinx/einx/context"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/network"
	"github.com/Cyinx/einx/slog"
	"github.com/Cyinx/einx/timer"
)

type Agent = agent.Agent
type AgentID = agent.AgentID
type Component = component.Component
type ComponentID = component.ComponentID
type EventMsg = event.EventMsg
type EventType = event.EventType
type ComponentEventMsg = event.ComponentEventMsg
type SessionEventMsg = event.SessionEventMsg
type DataEventMsg = event.DataEventMsg
type RpcEventMsg = event.RpcEventMsg
type AwaitRpcEventMsg = event.AwaitRpcEventMsg
type CustomActionEventMsg = event.CustomActionEventMsg
type EventQueue = event.EventQueue
type TimerHandler = timer.TimerHandler
type TimerManager = timer.TimerManager
type SessionMgr = network.SessionMgr
type Module = context.Module
type Context = context.Context
type ProtoTypeID = uint32
type MsgHandler func(Context, interface{})
type RpcHandler func(Context, *ArgsVar)

type ModuleRouter interface {
	RouterMsg(Agent, ProtoTypeID, interface{})
	RouterRpc(Agent, string, []interface{})
	RegisterHandler(ProtoTypeID, MsgHandler)
	RegisterRpcHandler(string, RpcHandler)
}

type ModuleWoker interface {
	Close()
	Run(*sync.WaitGroup)
}

var (
	MODULE_TIMER_INTERVAL = 1
	MODULE_EVENT_LENGTH   = 128
)

type module struct {
	id            AgentID
	evQueue       *EventQueue
	name          string
	msgHandlerMap map[ProtoTypeID]MsgHandler
	rpcHandlerMap map[string]RpcHandler
	agentMap      map[AgentID]Agent
	commgrMap     map[ComponentID]ComponentMgr
	componentMap  map[ComponentID]Component
	awaitMsgPool  *sync.Pool
	rpcMsgPool    *sync.Pool
	dataMsgPool   *sync.Pool
	eventMsgPool  *sync.Pool
	timerManager  *TimerManager
	opCount       int64
	closeChan     chan bool
	context       *ModuleContext
	args          ArgsVar
	eventList     []interface{}
	eventCount    uint32
	eventIndex    uint32
	beginTime     int64
}

func (m *module) GetID() AgentID {
	return m.id
}

func (m *module) GetName() string {
	return m.name
}
func (m *module) PushEventMsg(ev EventMsg) {
	m.evQueue.Push(ev)
}

func (m *module) Close() {
	m.closeChan <- true
	slog.LogWarning("module", "module [%s] will close!", m.name)
}

func (m *module) AddTimer(delay uint64, op TimerHandler, args ...interface{}) uint64 {
	return m.timerManager.AddTimer(delay, op, args...)
}

func (m *module) RemoveTimer(timerId uint64) bool {
	return m.timerManager.DeleteTimer(timerId)
}

func (m *module) PostEvent(eventType EventType, agent Agent, cid ComponentID, args ...interface{}) {
	e := m.eventMsgPool.Get().(*SessionEventMsg)
	e.MsgType = eventType
	e.Sender = agent
	e.Cid = cid
	e.Args = args
	m.evQueue.Push(e)
}

func (m *module) PostData(eventType EventType, typeID ProtoTypeID, agent Agent, data interface{}) {
	event := m.dataMsgPool.Get().(*DataEventMsg)
	event.MsgType = eventType
	event.Sender = agent
	event.TypeID = typeID
	event.MsgData = data
	m.evQueue.Push(event)
}

func (m *module) RpcCall(name string, args ...interface{}) {
	rpc_msg := m.rpcMsgPool.Get().(*RpcEventMsg)
	rpc_msg.MsgType = event.EVENT_MODULE_RPC
	rpc_msg.Sender = m
	rpc_msg.Data = args
	rpc_msg.RpcName = name
	m.evQueue.Push(rpc_msg)
}

func (m *module) AwaitRpcCall(name string, args ...interface{}) []interface{} {
	rpc_msg := m.awaitMsgPool.Get().(*AwaitRpcEventMsg)
	rpc_msg.MsgType = event.EVENT_MODULE_AWAITRPC
	rpc_msg.Sender = m
	rpc_msg.Data = args
	rpc_msg.RpcName = name
	rpc_msg.Await = make(chan []interface{})
	m.evQueue.Push(rpc_msg)
	return <-rpc_msg.Await
}

func (m *module) RouterMsg(agent Agent, msgID ProtoTypeID, msg interface{}) {
	m.PostData(event.EVENT_TCP_READ_MSG, msgID, agent, msg)
}

func (m *module) RouterRpc(agent Agent, name string, args []interface{}) {
	rpc_msg := m.rpcMsgPool.Get().(*RpcEventMsg)
	rpc_msg.MsgType = event.EVENT_MODULE_RPC
	rpc_msg.Sender = agent
	rpc_msg.Data = args
	rpc_msg.RpcName = name
	m.evQueue.Push(rpc_msg)
}

func (m *module) RegisterHandler(typeID ProtoTypeID, handler MsgHandler) {
	_, ok := m.msgHandlerMap[typeID]
	if ok == true {
		slog.LogWarning("module", "MsgID[%d] has been registered", typeID)
		return
	}
	m.msgHandlerMap[typeID] = handler
}

func (m *module) RegisterRpcHandler(rpcName string, handler RpcHandler) {
	_, ok := m.rpcHandlerMap[rpcName]
	if ok == true {
		slog.LogWarning("module", "Rpc[%s] has been registered", rpcName)
		return
	}
	m.rpcHandlerMap[rpcName] = handler
}

func (m *module) recover(wait *sync.WaitGroup) {
	if r := recover(); r != nil {
		slog.LogError("module_recovery", "recover error :%v", r)
		slog.LogError("module_recovery", "%s", string(debug.Stack()))
		go m.Run(wait) // continue to run
	}
}

func (m *module) Run(wait *sync.WaitGroup) {
	runtime.LockOSThread()
	defer m.recover(wait)
	wait.Add(1)
	defer wait.Done()
	m.beginTime = time.Now().UnixNano() / 1e9
	timerManager := m.timerManager
	var (
		eventMsg  EventMsg = nil
		closeFlag bool     = false
		evQueue            = m.evQueue
		eventChan          = evQueue.SemaChan()
		timerTick          = time.NewTimer(time.Duration(MODULE_TIMER_INTERVAL) * time.Millisecond)
		tickC              = timerTick.C
		nextWake           = 0
		eventList          = m.eventList
	)

	for {
		for {
			if m.eventIndex >= m.eventCount {
				m.eventCount = evQueue.Get(eventList, uint32(MODULE_EVENT_LENGTH))
				m.eventIndex = 0
			}
			for m.eventIndex < m.eventCount {
				eventMsg = eventList[m.eventIndex].(EventMsg)
				eventList[m.eventIndex] = nil
				m.eventIndex++
				m.handleEvent(eventMsg)
				m.opCount++
			}
			nextWake = timerManager.Execute(100)
			if m.eventCount <= 0 && nextWake > 0 {
				break
			}
		}

		if evQueue.WaitNotify() == false {
			continue
		}

		timerTick.Reset(time.Duration(nextWake) * time.Millisecond)
		select {
		case closeFlag = <-m.closeChan:
			if closeFlag == true {
				goto runClose
			}
		case <-eventChan:
		case <-tickC:
		}

		evQueue.WaiterWake()
	}

runClose:
	m.doClose(wait)
}

func (m *module) doClose(wait *sync.WaitGroup) {
	for _, c := range m.componentMap {
		c.Close()
	}
	for _, a := range m.agentMap {
		a.Close()
	}
	if PerfomancePrint == true {
		elaspTime := time.Now().UnixNano()/1e9 - m.beginTime
		slog.LogError("perfomance", "module perfomance [%s] %d %d %d", m.name, elaspTime, m.opCount, m.opCount/elaspTime)
	}
	//slog.LogWarning("module", "module [%s] closed!", m.name)
}

func (m *module) handleEvent(eventMsg EventMsg) {
	switch eventMsg.GetType() {
	case event.EVENT_TCP_READ_MSG:
		m.handleDataEvent(eventMsg)
	case event.EVENT_COMPONENT_CREATE:
		m.handleComponentEvent(eventMsg)
	case event.EVENT_COMPONENT_ERROR:
		m.handleComponentError(eventMsg)
	case event.EVENT_TCP_ACCEPTED:
		m.handleAgentEnter(eventMsg)
	case event.EVENT_TCP_CONNECTED:
		m.handleAgentEnter(eventMsg)
	case event.EVENT_TCP_CLOSED:
		m.handleAgentClosed(eventMsg)
	case event.EVENT_MODULE_RPC:
		m.handleRpc(eventMsg)
	case event.EVENT_MODULE_AWAITRPC:
		m.handleAwaitRpc(eventMsg)
	case event.EVENT_COMPONENT_CUSTOM:
		m.handleCustomAction(eventMsg)
	default:
		slog.LogError("einx", "handleEvent unknow event msg [%v]", eventMsg.GetType())
	}
}

func (m *module) handleDataEvent(eventMsg EventMsg) {
	dataEventMsg := eventMsg.(*DataEventMsg)
	handler, ok := m.msgHandlerMap[dataEventMsg.TypeID]
	if ok == true {
		ctx := m.context
		ctx.s = dataEventMsg.Sender
		handler(ctx, dataEventMsg.MsgData)
		ctx.Reset()
	} else {
		slog.LogError("module", "module [%s] unregister msg handler msg type id[%d] %v!", m.name, dataEventMsg.TypeID, ok)
	}
	eventMsg.Reset()
	m.dataMsgPool.Put(eventMsg)
}

func (m *module) handleComponentEvent(eventMsg EventMsg) {
	comEvent := eventMsg.(*ComponentEventMsg)
	c := comEvent.Sender
	if _, ok := m.componentMap[c.GetID()]; ok == true {
		slog.LogError("pakage", "module[%v] register pakage[%v]", m.name, c.GetID())
		return
	}
	mgr := comEvent.Attach.(ComponentMgr)
	m.commgrMap[c.GetID()] = mgr
	m.componentMap[c.GetID()] = c
	ctx := m.context
	ctx.c = c
	mgr.OnComponentCreate(ctx, c.GetID())
	ctx.Reset()
}

func (m *module) handleComponentError(eventMsg EventMsg) {
	comEvent := eventMsg.(*ComponentEventMsg)
	c := comEvent.Sender
	if mgr, ok := m.commgrMap[c.GetID()]; ok == true {
		ctx := m.context
		ctx.c = c
		ctx.t = comEvent.Attach
		mgr.OnComponentError(ctx, comEvent.Err)
		ctx.Reset()
		return
	}
	slog.LogError("pakage", "module[%v] not register pakage[%v] manager cant handle error:%v", m.name, c.GetID(), comEvent.Attach)
}

func (m *module) handleAgentEnter(eventMsg EventMsg) {
	s := eventMsg.(*SessionEventMsg)
	a := s.Sender.(Agent)
	m.agentMap[a.GetID()] = a

	if sesMgr, ok := m.commgrMap[s.Cid]; ok == true {
		sesMgr.(SessionMgr).OnLinkerConnected(a.GetID(), a)
		return
	}

	slog.LogError("agent", "module[%v] agent enter not found pakage[%v]", m.name, s.Cid)
}

func (m *module) handleAgentClosed(eventMsg EventMsg) {
	s := eventMsg.(*SessionEventMsg)
	sender := s.Sender.(Agent)
	delete(m.agentMap, sender.GetID())
	if sesMgr, ok := m.commgrMap[s.Cid]; ok == true {
		var err error = nil
		if len(s.Args) > 0 {
			err, _ = s.Args[0].(error)
		}
		sesMgr.(SessionMgr).OnLinkerClosed(sender.GetID(), sender, err)
		return
	}

	slog.LogError("agent", "module[%v] agent closed not found pakage[%v]", m.name, s.Cid)
}

func (m *module) handleRpc(eventMsg EventMsg) {
	rpcMsg := eventMsg.(*RpcEventMsg)
	if handler, ok := m.rpcHandlerMap[rpcMsg.RpcName]; ok == true {
		ctx := m.context
		args := &m.args
		ctx.s = rpcMsg.Sender
		args.ref(rpcMsg.Data)
		handler(ctx, args)
		args.clear()
		ctx.Reset()
	} else {
		slog.LogError("module", "module [%v] unregister rpc handler! rpc name:[%v]", m.name, rpcMsg.RpcName)
	}
	eventMsg.Reset()
	m.rpcMsgPool.Put(rpcMsg)
}

func (m *module) handleAwaitRpc(eventMsg EventMsg) {
	rpcMsg := eventMsg.(*AwaitRpcEventMsg)
	if handler, ok := m.rpcHandlerMap[rpcMsg.RpcName]; ok == true {
		ctx := &ModuleContext{}
		args := &m.args
		ctx.s = rpcMsg.Sender
		ctx.u = rpcMsg.Await
		args.ref(rpcMsg.Data)
		handler(ctx, args)
		args.clear()
	} else {
		slog.LogError("module", "module [%v] unregister rpc handler! rpc name:[%v]", m.name, rpcMsg.RpcName)
	}
	eventMsg.Reset()
	m.awaitMsgPool.Put(rpcMsg)
}

func (m *module) handleCustomAction(eventMsg EventMsg) {
	customMsg := eventMsg.(CustomActionEventMsg)
	action := customMsg.GetAction()
	if action != nil {
		action(customMsg)
	}
}
