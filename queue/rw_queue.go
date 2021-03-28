package queue

import (
	"sync"
)

type RWQueue struct {
	readQueue  *Queue
	writeQueue *Queue

	queue_1 Queue
	queue_2 Queue

	readLock  sync.Mutex
	writeLock sync.Mutex
}

func NewRWQueue() *RWQueue {
	queue := &RWQueue{
		queue_1: *NewQueue(),
		queue_2: *NewQueue(),
	}
	queue.init()
	return queue
}

func (q *RWQueue) init() {
	q.readQueue = &q.queue_1
	q.writeQueue = &q.queue_2
}

func (q *RWQueue) exchange() bool {
	temp_queue := q.writeQueue
	if temp_queue.empty() == true {
		return false
	}
	q.writeQueue = q.readQueue
	q.readQueue = temp_queue
	return true
}

func (q *RWQueue) Push(event interface{}) {
	q.writeLock.Lock()
	q.writeQueue.push(event)
	q.writeLock.Unlock()
}

func (q *RWQueue) Get(event_list []interface{}, count uint32) (uint32, int) {
	q.readLock.Lock()

	if q.readQueue.empty() == true {
		q.writeLock.Lock()
		if q.exchange() == false {
			q.writeLock.Unlock()
			q.readLock.Unlock()
			return 0, 0
		}
		q.writeLock.Unlock()
	}

	readQueue := q.readQueue
	var readCount uint32 = 0
	var leftCount int = 0
	for {
		val, ret := readQueue.pop()
		if ret == false {
			break
		}
		event_list[readCount] = val
		readCount++
		if readCount == count {
			break
		}
	}
	leftCount = readQueue.count()
	q.readLock.Unlock()
	return readCount, leftCount
}

func (q *RWQueue) GetOne() interface{} {
	q.readLock.Lock()
	defer q.readLock.Unlock()

	if q.readQueue.empty() == true {
		q.writeLock.Lock()
		if q.exchange() == false {
			q.writeLock.Unlock()
			return 0
		}
		q.writeLock.Unlock()
	}

	readQueue := q.readQueue
	val, ret := readQueue.pop()
	if ret == false {
		return nil
	}
	return val
}

func (q *RWQueue) Empty() bool {
	q.readLock.Lock()
	defer q.readLock.Unlock()

	return q.readQueue.empty()
}
