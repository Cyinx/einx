package network

import (
	//"github.com/Cyinx/einx/slog"
	"github.com/Cyinx/einx/timer"
	"sync"
	"sync/atomic"
	"time"
)

var (
	EnableKeepAlive       = true
	PINGTIME        int64 = 5 * 1000 //Millisecond
	PONGTIME        int64 = PINGTIME * 2
	CHECKTIME       int64 = 256 //Millisecond
)

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

var nowTick int64 = 0

func SetKeepAlive(open bool, pingTime int64) {
	EnableKeepAlive = open
	PINGTIME = pingTime
}

func OnPing(args []interface{}) {
	linker := args[0].(Linker)
	linker.Ping()
	timer_id := alive_manager.timer_manager.AddTimer(uint64(PINGTIME), OnPing, linker)
	alive_manager.linkers[linker] = timer_id
}

func AddPing(linker Linker) {
	if EnableKeepAlive == false {
		return
	}
	alive_manager.locker.Lock()
	timer_id := alive_manager.timer_manager.AddTimer(uint64(PINGTIME), OnPing, linker)
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
	timer_id := alive_manager.timer_manager.AddTimer(uint64(PONGTIME), OnPong, linker)
	alive_manager.linkers[linker] = timer_id
}

func AddPong(linker Linker) {
	if EnableKeepAlive == false {
		return
	}
	alive_manager.locker.Lock()
	timer_id := alive_manager.timer_manager.AddTimer(uint64(PONGTIME), OnPong, linker)
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

func GetNowTick() int64 {
	return atomic.LoadInt64(&nowTick)
}

func OnKeepAliveUpdate() {
	var ticker = time.NewTicker(time.Duration(CHECKTIME) * time.Millisecond)
	timer_manager := alive_manager.timer_manager
	locker := alive_manager.locker
	for {
		<-ticker.C
		atomic.StoreInt64(&nowTick, time.Now().Unix())
		locker.Lock()
		timer_manager.Execute(100)
		locker.Unlock()
	}
}
