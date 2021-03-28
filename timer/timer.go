package timer

type TimerHandler func(args []interface{})

type xtimer struct {
	seqID   uint32
	runTick uint64
	args    []interface{}
	handler TimerHandler
	next    *xtimer
	running bool
}

func newXTimer() *xtimer {
	timer := &xtimer{
		running: false,
	}
	return timer
}

func (x *xtimer) reset() {
	x.seqID = 0
	x.runTick = 0
	x.args = x.args[:0]
	x.handler = nil
	x.next = nil
	x.running = false
}

func (x *xtimer) getTimerId() uint64 {
	seqID := uint64(x.seqID)
	return ((x.runTick << 24) | (seqID & 0xffffff))
}

type timerList struct {
	head    *xtimer
	tail    *xtimer
	running *xtimer
	pool    *timerPool
}

func newTimerList(p *timerPool) *timerList {
	timerlist := &timerList{
		head: nil,
		tail: nil,
		pool: p,
	}
	return timerlist
}

func (l *timerList) addTimer(timer *xtimer) {
	timer.next = nil
	if l.tail == nil {
		l.tail = timer
		l.head = l.tail
		return
	}
	l.tail.next = timer
	l.tail = timer
}

func (l *timerList) getTimer(seqID uint32) *xtimer {
	currHead := l.head
	for currHead != nil {
		if currHead.seqID == seqID {
			return currHead
		}
		currHead = currHead.next
	}
	return nil
}

func (l *timerList) deleteTimer(seqID uint32) bool {
	running := l.running
	if running != nil && running.seqID == seqID {
		return true
	}
	currHead := l.head
	var prevHead *xtimer = nil
	for currHead != nil {
		if currHead.seqID != seqID {
			prevHead = currHead
			currHead = currHead.next
			continue
		}

		if currHead.running {
			return true
		}

		if prevHead == nil {
			l.head = currHead.next
		} else {
			prevHead.next = currHead.next
		}

		if currHead.next == nil {
			l.tail = prevHead
		}
		l.pool.Put(currHead)
		return true
	}
	return false
}

func (l *timerList) execute(now uint64, count uint32) (uint32, bool) {
	var runningTimer *xtimer = nil
	runCount := uint32(0)

	for ; l.head != nil; runCount++ {

		runningTimer = l.head
		if runCount >= count || runningTimer.runTick > now {
			return runCount, false
		}

		l.head = runningTimer.next

		if l.head == nil {
			l.tail = nil
		}

		l.running = runningTimer

		runningTimer.running = true
		runningTimer.handler(runningTimer.args)

		l.running = nil

		l.pool.Put(runningTimer)
	}

	return runCount, true
}
