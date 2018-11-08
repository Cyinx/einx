package timer

import (
	"time"
)

const TIMERWHEELCOUNT = 5

type TimerManager struct {
	seqIDIndex   uint32
	timer_wheels [5]*timerWheel
}

var beginTick uint64 = uint64(time.Now().UnixNano() / 1e6)

func UnixTS() uint64 {
	return uint64(time.Now().UnixNano()/1e6) - beginTick
}

func NewTimerManager() *TimerManager {

	timer_manager := &TimerManager{
		seqIDIndex: 0,
	}

	now := UnixTS()

	timer_wheels := &timer_manager.timer_wheels

	timer_wheels[0] = newTimerWheel(1, 0, now)
	timer_wheels[1] = newTimerWheel(0xff+1, 8, now)
	timer_wheels[2] = newTimerWheel(0xffff+1, 16, now)
	timer_wheels[3] = newTimerWheel(0xffffff+1, 24, now)
	timer_wheels[4] = newTimerWheel(0xffffffff+1, 32, now)

	timer_wheels[0].next_wheel = timer_wheels[1]
	timer_wheels[1].prev_wheel = timer_wheels[0]
	timer_wheels[1].next_wheel = timer_wheels[2]
	timer_wheels[2].prev_wheel = timer_wheels[1]
	timer_wheels[2].next_wheel = timer_wheels[3]
	timer_wheels[3].prev_wheel = timer_wheels[2]
	timer_wheels[3].next_wheel = timer_wheels[4]
	timer_wheels[4].prev_wheel = timer_wheels[3]

	return timer_manager
}

func (this *TimerManager) GetSeqID() uint32 {
	this.seqIDIndex++
	if this.seqIDIndex == 0 || this.seqIDIndex >= 0xffffff {
		this.seqIDIndex = 1
	}
	return this.seqIDIndex
}

func (this *TimerManager) AddTimer(delay uint64, op TimerHandler, args ...interface{}) uint64 {
	seqID := this.GetSeqID()

	if delay < 0 {
		delay = 0
	}

	run_tick := UnixTS() + delay

	if run_tick > 0x000000ffffffffff {
		run_tick = run_tick & 0x000000ffffffffff
	}

	xtimer := newXTimer()
	xtimer.args = args
	xtimer.handler = op
	xtimer.next = nil
	xtimer.running = false
	xtimer.seqID = seqID
	xtimer.runTick = run_tick

	this.timer_wheels[4].add_timer(xtimer)

	return xtimer.get_timer_id()
}

func (this *TimerManager) DeleteTimer(timerID uint64) {
	if timerID == 0 {
		return
	}

	this.timer_wheels[4].delete_timer(timerID>>24, uint32(timerID&0xffffff))
}

func (this *TimerManager) Execute(count uint32) int {
	now := UnixTS()
	wheel := this.timer_wheels[0]
	wheel.execute(now, count)
	return wheel.nextWake()
}
