package timer

const TIMERSLOTSCOUNT = 256

type timerWheel struct {
	array    [TIMERSLOTSCOUNT]*timerList
	index    uint8
	bitSize  uint32
	msUnit   uint64
	baseTick uint64

	timerCount uint64

	next_wheel *timerWheel
	prev_wheel *timerWheel
}

func newTimerWheel(ms_unit uint64, bit_size uint32, now uint64, p *timerPool) *timerWheel {
	timer_wheel := &timerWheel{
		index:      0,
		bitSize:    bit_size,
		msUnit:     ms_unit,
		next_wheel: nil,
		prev_wheel: nil,
	}

	if ms_unit == 1 {
		timer_wheel.baseTick = now
	} else {
		timer_wheel.baseTick = now + ms_unit
	}

	for i := 0; i < TIMERSLOTSCOUNT; i++ {
		timer_wheel.array[i] = newTimerList(p)
	}

	return timer_wheel
}

func (this *timerWheel) tickIdxDelta(runTick uint64) uint8 {
	idxDelta := runTick - this.baseTick
	idxDelta = idxDelta >> this.bitSize
	return uint8(idxDelta)
}

func (this *timerWheel) add_timer(timer *xtimer) {
	if this.prev_wheel != nil && timer.runTick < this.baseTick {
		this.prev_wheel.add_timer(timer)
		return
	}

	idx := uint8(this.index + this.tickIdxDelta(timer.runTick))
	timer_list := this.array[idx]
	timer_list.add_timer(timer)
	this.timerCount++
}

func (this *timerWheel) delete_timer(run_tick uint64, seq_id uint32) bool {
	if this.prev_wheel != nil && run_tick < this.baseTick {
		return this.prev_wheel.delete_timer(run_tick, seq_id)
	}
	idx := (this.index + uint8(this.tickIdxDelta(run_tick)))
	timer_list := this.array[idx]
	success := timer_list.delete_timer(seq_id)
	this.timerCount--
	return success
}

func (this *timerWheel) execute(now uint64, count uint32) uint32 {
	if this.prev_wheel != nil || now < this.baseTick {
		return 0
	}

	elapsedTime := uint64(now - this.baseTick)
	loopTimes := uint64(1 + elapsedTime)

	var run_count uint32 = 0
	for ; run_count < count && loopTimes > 0; loopTimes-- {
		timer_list := this.array[this.index]
		c, b := timer_list.execute(now, count-run_count)
		run_count = run_count + c
		if b == false {
			return run_count
		}
		this.index++
		this.baseTick += this.msUnit
		if this.index == 0 {
			this.next_wheel.TurnWheel()
		}
	}
	return run_count
}

func (this *timerWheel) nextWake() int {
	wakeDelay := 0
	wakeIndex := this.index
	for wakeDelay < 64 {
		timer_list := this.array[wakeIndex]
		if timer_list.head != nil {
			break
		}
		if wakeIndex++; wakeIndex == 0 {
			break
		}
		wakeDelay++
	}
	return wakeDelay
}

func (this *timerWheel) TurnWheel() {
	if this.prev_wheel == nil {
		return
	}

	timer_list := this.array[this.index]
	head_timer := timer_list.head
	var next_timer *xtimer = nil
	for head_timer != nil {
		next_timer = head_timer.next
		head_timer.next = nil
		this.prev_wheel.add_timer(head_timer)
		head_timer = next_timer
		this.timerCount--
	}

	timer_list.head = nil
	timer_list.tail = nil

	this.index++
	this.baseTick += this.msUnit

	if this.index == 0 && this.next_wheel != nil {
		this.next_wheel.TurnWheel()
	}
}
