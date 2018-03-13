package network

import (
	"github.com/Cyinx/einx/timer"
	"sync"
	"time"
)

const PINGTIME = 15 * 1000
const PONGTIME = PINGTIME * 2

type TimerHandler = timer.TimerHandler
type TimerManager = timer.TimerManager

type aliveManager struct {
	timer_manager *TimerManager
	linkers       map[Linker]uint64
	locker        *sync.Mutex
}

var alive_manager = &aliveManager{
	timer_manager: timer.NewTimerManager(),
	linkers:       make(map[Linker]uint64),
	locker:        new(sync.Mutex),
}

var NowKeepAliveTick int64 = 0

func init() {
	go OnUpdate()
}

func OnPing(args []interface{}) {
	linker := args[0].(Linker)
	linker.Ping()
	timer_id := alive_manager.timer_manager.AddTimer(PINGTIME, OnPing, linker)
	alive_manager.linkers[linker] = timer_id
}

func AddPing(linker Linker) {
	alive_manager.locker.Lock()
	timer_id := alive_manager.timer_manager.AddTimer(PINGTIME, OnPing, linker)
	alive_manager.linkers[linker] = timer_id
	alive_manager.locker.Unlock()
}

func RemovePing(linker Linker) {
	alive_manager.locker.Lock()
	if timer_id, ok := alive_manager.linkers[linker]; ok == true {
		delete(alive_manager.linkers, linker)
		alive_manager.timer_manager.DeleteTimer(timer_id)
	}
	alive_manager.locker.Unlock()
}

func OnPong(args []interface{}) {
	linker := args[0].(Linker)
	linker.Pong()
	timer_id := alive_manager.timer_manager.AddTimer(PONGTIME, OnPong, linker)
	alive_manager.linkers[linker] = timer_id
}

func AddPong(linker Linker) {
	alive_manager.locker.Lock()
	timer_id := alive_manager.timer_manager.AddTimer(PONGTIME, OnPong, linker)
	alive_manager.linkers[linker] = timer_id
	alive_manager.locker.Unlock()
}

func RemovePong(linker Linker) {
	alive_manager.locker.Lock()
	if timer_id, ok := alive_manager.linkers[linker]; ok == true {
		delete(alive_manager.linkers, linker)
		alive_manager.timer_manager.DeleteTimer(timer_id)
	}
	alive_manager.locker.Unlock()
}

func OnUpdate() {
	var ticker = time.NewTicker(15 * time.Millisecond)
	timer_manager := alive_manager.timer_manager
	locker := alive_manager.locker
	for {
		<-ticker.C
		NowKeepAliveTick = time.Now().UnixNano() / 1e9
		locker.Lock()
		timer_manager.Execute(100)
		locker.Unlock()
	}
}
