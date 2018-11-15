package timer

type timerPool struct {
	p []*xtimer
}

func newTimerPool() *timerPool {
	t := &timerPool{
		p: make([]*xtimer, 0, 4096),
	}
	return t
}

func (t *timerPool) Get() *xtimer {
	var x *xtimer = nil
	last := len(t.p) - 1
	if last >= 0 {
		x = t.p[last]
		t.p = t.p[:last]
	} else {
		x = newXTimer()
	}
	return x
}

func (t *timerPool) Put(x *xtimer) {
	x.reset()
	t.p = append(t.p, x)
}
