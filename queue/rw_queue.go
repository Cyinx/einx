package queue

import (
	"sync"
)

type RWQueue struct {
	read_queue  *Queue
	write_queue *Queue

	queue_1 Queue
	queue_2 Queue

	read_lock  sync.Mutex
	write_lock sync.Mutex
}

func NewRWQueue() *RWQueue {
	queue := &RWQueue{
		queue_1: *NewQueue(),
		queue_2: *NewQueue(),
	}
	queue.init()
	return queue
}

func (this *RWQueue) init() {
	this.read_queue = &this.queue_1
	this.write_queue = &this.queue_2
}

func (this *RWQueue) exchange() bool {
	temp_queue := this.write_queue
	if temp_queue.empty() == true {
		return false
	}
	this.write_queue = this.read_queue
	this.read_queue = temp_queue
	return true
}

func (this *RWQueue) Push(event interface{}) {
	this.write_lock.Lock()
	this.write_queue.push(event)
	this.write_lock.Unlock()
}

func (this *RWQueue) Get(event_list []interface{}, count uint32) (uint32, int) {
	this.read_lock.Lock()

	if this.read_queue.empty() == true {
		this.write_lock.Lock()
		if this.exchange() == false {
			this.write_lock.Unlock()
			this.read_lock.Unlock()
			return 0, 0
		}
		this.write_lock.Unlock()
	}

	read_queue := this.read_queue
	var read_count uint32 = 0
	var left_count int = 0
	for {
		val, ret := read_queue.pop()
		if ret == false {
			break
		}
		event_list[read_count] = val
		read_count++
		if read_count == count {
			break
		}
	}
	left_count = read_queue.count()
	this.read_lock.Unlock()
	return read_count, left_count
}

func (this *RWQueue) GetOne() interface{} {
	this.read_lock.Lock()
	defer this.read_lock.Unlock()

	if this.read_queue.empty() == true {
		this.write_lock.Lock()
		if this.exchange() == false {
			this.write_lock.Unlock()
			return 0
		}
		this.write_lock.Unlock()
	}

	read_queue := this.read_queue
	val, ret := read_queue.pop()
	if ret == false {
		return nil
	}
	return val
}

func (this *RWQueue) Empty() bool {
	this.read_lock.Lock()
	defer this.read_lock.Unlock()

	return this.read_queue.empty()
}
