package network

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/slog"
	"github.com/Cyinx/einx/timer"
	"runtime/debug"
	"sync"
	"time"
)

type EventQueue = event.EventQueue

const KEEP_ALIVE_POLL_COUNT = 128

type PingEventMsg struct {
	Sender Linker
	Op     int
	Tick   int64
}

func (this *PingEventMsg) GetType() event.EventType {
	return 0
}

func (this *PingEventMsg) GetSender() Agent {
	return nil
}

func (this *PingEventMsg) Reset() {
	this.Sender = nil
	this.Op = 0
}

const (
	PING_OP_TYPE_ADD_PING = iota
	PING_OP_TYPE_REMOVE_PING
	PING_OP_TYPE_DO_PONG
)

var (
	//PINGTIME        int64 = 5 * 1000 //Millisecond
	//PONGTIME        int64 = PINGTIME * 2
	CHECKTIME int64 = 256 //Millisecond
)

type TimerHandler = timer.TimerHandler
type TimerManager = timer.TimerManager

type PingMgr struct {
	timer_manager *TimerManager
	linkers       map[Linker]uint64
	ev_queue      *EventQueue
	event_pool    *sync.Pool
	event_list    []interface{}
	event_count   uint32
	event_index   uint32
}

var pingMgr = &PingMgr{
	timer_manager: timer.NewTimerManager(),
	linkers:       make(map[Linker]uint64),
	ev_queue:      event.NewEventQueue(),
	event_pool:    &sync.Pool{New: func() interface{} { return new(PingEventMsg) }},
	event_list:    make([]interface{}, KEEP_ALIVE_POLL_COUNT),
}

var beginTime = time.Now()

func UnixTS() int64 {
	tick := time.Since(beginTime).Nanoseconds()
	return int64(tick / 1e6)
}

func (p *PingMgr) OnPing(args []interface{}) {
	linker := args[0].(Linker)
	if linker.Ping() == true {
		timer_id := p.timer_manager.AddTimer(uint64(linker.GetOption().ping_time), p.OnPing, linker)
		p.linkers[linker] = timer_id
	} else {
		delete(p.linkers, linker)
	}
}

func (p *PingMgr) AddPing(linker Linker) {
	if linker.GetOption().enable_ping == false {
		return
	}
	event_msg := p.event_pool.Get().(*PingEventMsg)
	event_msg.Sender = linker
	event_msg.Op = PING_OP_TYPE_ADD_PING
	p.ev_queue.Push(event_msg)
}

func (p *PingMgr) RemovePing(linker Linker) {
	event_msg := p.event_pool.Get().(*PingEventMsg)
	event_msg.Sender = linker
	event_msg.Op = PING_OP_TYPE_REMOVE_PING
	p.ev_queue.Push(event_msg)
}

func (p *PingMgr) DoPong(linker Linker, tick int64) {
	event_msg := p.event_pool.Get().(*PingEventMsg)
	event_msg.Sender = linker
	event_msg.Op = PING_OP_TYPE_DO_PONG
	event_msg.Tick = tick
	p.ev_queue.Push(event_msg)
}

func (p *PingMgr) recover() {
	if r := recover(); r != nil {
		slog.LogError("ping_recovery", "recover error :%v", r)
		slog.LogError("ping_recovery", "%s", string(debug.Stack()))
		go p.Run() // continue to run
	}
}

func (p *PingMgr) Run() {
	defer p.recover()
	var ticker = time.NewTicker(time.Duration(CHECKTIME) * time.Millisecond)
	timer_manager := p.timer_manager
	ev_queue := p.ev_queue
	event_chan := ev_queue.SemaChan()
	event_list := p.event_list
	for {

		if p.event_index >= p.event_count {
			p.event_count = ev_queue.Get(event_list, uint32(KEEP_ALIVE_POLL_COUNT))
			p.event_index = 0
		}

		for p.event_index < p.event_count {
			ping_event := event_list[p.event_index].(*PingEventMsg)
			event_list[p.event_index] = nil
			p.event_index++
			p.handle_event(ping_event)
			ping_event.Reset()
			p.event_pool.Put(ping_event)
		}

		timer_manager.Execute(256)

		if ev_queue.WaitNotify() == false {
			continue
		}

		select {
		case <-event_chan:
		case <-ticker.C:
		}

		ev_queue.WaiterWake()
	}
}

func (p *PingMgr) handle_event(e *PingEventMsg) {
	linker := e.Sender
	switch e.Op {
	case PING_OP_TYPE_ADD_PING:
		timer_id := p.timer_manager.AddTimer(uint64(linker.GetOption().ping_time), p.OnPing, linker)
		p.linkers[linker] = timer_id
	case PING_OP_TYPE_REMOVE_PING:
		if timer_id, ok := p.linkers[linker]; ok == true {
			delete(p.linkers, linker)
			p.timer_manager.DeleteTimer(timer_id)
		}
	case PING_OP_TYPE_DO_PONG:
		linker.DoPong(e.Tick)
	default:
		slog.LogDebug("pingMgr", "unknown ping event type [%v]", e.Op)
	}
}