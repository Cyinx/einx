package queue

import (
	"sync"
)

var node_pool *sync.Pool = &sync.Pool{New: func() interface{} { return new(Node) }}
var p_pool *sync.Pool = &sync.Pool{New: func() interface{} { return make([]*Node, 0, NODE_POOL_CELL_COUNT*2) }}

const NODE_POOL_CELL_COUNT = 8
const NODE_POOL_MAX_COUNT = 4096

type nodePool struct {
	p []*Node
	c int
}

func newNodePool() *nodePool {
	t := &nodePool{
		p: p_pool.Get().([]*Node),
	}
	return t
}

func (t *nodePool) Get() *Node {
	var x *Node = nil
	last := len(t.p) - 1
	if last >= 0 {
		x = t.p[last]
		t.p = t.p[:last]
	} else {
		x = node_pool.Get().(*Node)
	}
	return x
}

func (t *nodePool) Put(x *Node) {
	x.reset()
	if len(t.p) < cap(t.p) {
		t.p = append(t.p, x)
		t.c = 0
	} else {
		t.c++
		if t.c < NODE_POOL_CELL_COUNT || cap(t.p) >= NODE_POOL_MAX_COUNT {
			node_pool.Put(x)
			return
		}
		newCap := (2*NODE_POOL_CELL_COUNT + cap(t.p))
		s := make([]*Node, len(t.p), newCap)
		copy(s, t.p)
		o := t.p[:0]
		t.p = s
		p_pool.Put(o)
		t.c = 0
		t.p = append(t.p, x)
	}
}

type Node struct {
	data interface{}
	next *Node
}

func (n *Node) reset() {
	n.data = nil
	n.next = nil
}

type Queue struct {
	head      *Node
	end       *Node
	node_pool *nodePool
	c         int
}

func NewQueue() *Queue {
	q := &Queue{
		head:      nil,
		end:       nil,
		node_pool: newNodePool(),
		c:         0,
	}
	return q
}

func (q *Queue) push(data interface{}) {
	n := q.node_pool.Get()
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
