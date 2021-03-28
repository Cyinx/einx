package queue

import (
	//"github.com/Cyinx/einx/slog"
	"sync"
	"sync/atomic"
)

type nullMutex struct {
	refCount int32
}

func (l *nullMutex) AddWaiter() {
	atomic.AddInt32(&l.refCount, 1)
}

func (l *nullMutex) RemoveWaiter() {
	atomic.AddInt32(&l.refCount, -1)
}

func (l *nullMutex) Lock() {
}

func (l *nullMutex) Unlock() {
}

type CondQueue struct {
	rwQueue *RWQueue
	lock    *nullMutex
	cond    *sync.Cond
	count   int32
}

func NewCondQueue() *CondQueue {
	q := &CondQueue{
		rwQueue: NewRWQueue(),
		lock:    &nullMutex{refCount: 0},
	}
	q.cond = sync.NewCond(q.lock)
	return q

}

func (c *CondQueue) Push(msg interface{}) {
	c.rwQueue.Push(msg)
	atomic.AddInt32(&c.count, 1)
	lock := c.lock
	waitCount := atomic.LoadInt32(&lock.refCount)
	if waitCount <= 0 {
		return
	}
	c.cond.Signal()
}

func (c *CondQueue) Get(list []interface{}, count uint32) uint32 {
	lock := c.lock
	lock.AddWaiter()
	if atomic.LoadInt32(&c.count) <= 0 {
		c.cond.Wait()
	}
	lock.RemoveWaiter()
	readCount, _ := c.rwQueue.Get(list, count)
	atomic.AddInt32(&c.count, 0-int32(readCount))
	return readCount
}
