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

func (this *xtimer) reset() {
	this.seqID = 0
	this.runTick = 0
	this.args = this.args[:0]
	this.handler = nil
	this.next = nil
	this.running = false
}

func (this *xtimer) get_timer_id() uint64 {
	seqID := uint64(this.seqID)
	return ((this.runTick << 24) | (seqID & 0xffffff))
}

type timerList struct {
	head *xtimer
	tail *xtimer

	pool *timerPool
}

func newTimerList(p *timerPool) *timerList {
	timerlist := &timerList{
		head: nil,
		tail: nil,
		pool: p,
	}
	return timerlist
}

func (this *timerList) add_timer(timer *xtimer) {
	timer.next = nil
	if this.tail == nil {
		this.tail = timer
		this.head = this.tail
		return
	}
	this.tail.next = timer
	this.tail = timer
}

func (this *timerList) get_timer(seqID uint32) *xtimer {
	curr_head := this.head
	for curr_head != nil {
		if curr_head.seqID == seqID {
			return curr_head
		}
		curr_head = curr_head.next
	}
	return nil
}

func (this *timerList) delete_timer(seqID uint32) bool {
	curr_head := this.head
	var prev_head *xtimer = nil
	for curr_head != nil {
		if curr_head.seqID != seqID {
			prev_head = curr_head
			curr_head = curr_head.next
			continue
		}

		if curr_head.running {
			return true
		}

		if prev_head == nil {
			this.head = curr_head.next
		} else {
			prev_head.next = curr_head.next
		}

		if curr_head.next == nil {
			this.tail = prev_head
		}
		this.pool.Put(curr_head)
		return true
	}
	return false
}

func (this *timerList) execute(now uint64, count uint32) (uint32, bool) {
	var running_timer *xtimer = nil
	run_count := uint32(0)

	for ; this.head != nil; run_count++ {

		running_timer = this.head
		if run_count >= count || running_timer.runTick > now {
			return run_count, false
		}

		running_timer = this.head
		running_timer.running = true
		running_timer.handler(running_timer.args)
		this.head = running_timer.next
		this.pool.Put(running_timer)
	}

	if this.head == nil {
		this.tail = nil
	}

	return run_count, true
}
