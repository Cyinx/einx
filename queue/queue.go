package queue

import (
	"sync"
)

type Node struct {
	data interface{}
	next *Node
}

type Queue struct {
	head      *Node
	end       *Node
	node_pool *sync.Pool
	c         int
}

func NewQueue() *Queue {
	q := &Queue{
		head:      nil,
		end:       nil,
		node_pool: &sync.Pool{New: func() interface{} { return new(Node) }},
		c:         0,
	}
	return q
}

func (q *Queue) push(data interface{}) {
	n := q.node_pool.Get().(*Node)
	n.data = data
	n.next = nil
	if q.end == nil {
		q.head = n
		q.end = n
	} else {
		q.end.next = n
		q.end = n
	}
	q.c++
	return
}

func (q *Queue) pop() (interface{}, bool) {
	if q.head == nil {
		return nil, false
	}

	n := q.head
	data := n.data
	q.head = n.next
	if q.head == nil {
		q.end = nil
	}
	q.c--
	q.node_pool.Put(n)
	return data, true
}

func (q *Queue) count() int {
	return q.c
}

func (q *Queue) empty() bool {
	if q.head == nil {
		return true
	}
	return false
}
